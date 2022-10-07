package universe

import (
	"container/heap"
	"context"
	"sort"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const SortKind = "sort"

type SortOpSpec struct {
	Columns []string `json:"columns"`
	Desc    bool     `json:"desc"`
}

func init() {
	sortSignature := runtime.MustLookupBuiltinType("universe", "sort")

	runtime.RegisterPackageValue("universe", SortKind, flux.MustValue(flux.FunctionValue(SortKind, createSortOpSpec, sortSignature)))
	plan.RegisterProcedureSpec(SortKind, newSortProcedure, SortKind)
	plan.RegisterPhysicalRules(RemoveRedundantSort{})
	execute.RegisterTransformation(SortKind, createSortTransformation)
}

func createSortOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(SortOpSpec)

	if array, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.Columns, err = interpreter.ToStringArray(array)
		if err != nil {
			return nil, err
		}
	} else {
		// Default behavior to sort by value
		spec.Columns = []string{execute.DefaultValueColLabel}
	}

	if desc, ok, err := args.GetBool("desc"); err != nil {
		return nil, err
	} else if ok {
		spec.Desc = desc
	}

	return spec, nil
}

func (s *SortOpSpec) Kind() flux.OperationKind {
	return SortKind
}

type SortProcedureSpec struct {
	plan.DefaultCost
	Columns []string
	Desc    bool
}

func newSortProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SortOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &SortProcedureSpec{
		Columns: spec.Columns,
		Desc:    spec.Desc,
	}, nil
}

func (s *SortProcedureSpec) Kind() plan.ProcedureKind {
	return SortKind
}
func (s *SortProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.Columns = make([]string, len(s.Columns))
	copy(ns.Columns, s.Columns)
	return &ns
}

func (s *SortProcedureSpec) OutputAttributes() plan.PhysicalAttributes {
	return plan.PhysicalAttributes{
		plan.CollationKey: &plan.CollationAttr{
			Columns: s.Columns,
			Desc:    s.Desc,
		},
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SortProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createSortTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SortProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewSortTransformation(id, s, a.Allocator())
}

type sortTransformation struct {
	execute.ExecutionNode
	d       *execute.PassthroughDataset
	mem     memory.Allocator
	cols    []string
	compare arrowutil.CompareFunc
}

func NewSortTransformation(id execute.DatasetID, spec *SortProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &sortTransformation{
		d:       execute.NewPassthroughDataset(id),
		mem:     mem,
		cols:    spec.Columns,
		compare: arrowutil.Compare,
	}
	if spec.Desc {
		// If descending, use the descending comparison.
		t.compare = arrowutil.CompareDesc
	}
	return t, t.d, nil
}

func (s *sortTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	sortCols := s.sortCols(tbl.Key(), tbl.Cols())
	mh := &sortTableMergeHeap{
		cols:     tbl.Cols(),
		key:      tbl.Key(),
		sortCols: sortCols,
		compare:  s.compare,
	}
	if err := tbl.Do(func(cr flux.ColReader) error {
		return s.processView(mh, cr)
	}); err != nil {
		return err
	}

	out, err := mh.Table(-1, s.mem)
	if err != nil {
		return err
	}
	return s.d.Process(out)
}

func (s *sortTransformation) sortCols(key flux.GroupKey, cols []flux.ColMeta) []int {
	sortCols := make([]int, 0, len(s.cols))
	for _, col := range s.cols {
		if idx := execute.ColIdx(col, cols); idx >= 0 {
			// If the sort key is part of the group key, skip it anyway.
			// They are all sorted anyway.
			if key.HasCol(col) {
				continue
			}
			sortCols = append(sortCols, idx)
		}
	}
	return sortCols
}

func (s *sortTransformation) processView(mh *sortTableMergeHeap, cr flux.ColReader) error {
	if cr.Len() == 0 {
		return nil
	}

	cr.Retain()
	item := &sortTableMergeHeapItem{cr: cr}
	if !s.isSorted(cr, mh.sortCols) {
		item.indices = s.sort(cr, mh.sortCols)
		item.offset = int(item.indices.Value(0))
	}
	mh.items = append(mh.items, item)
	return nil
}

func (s *sortTransformation) isSorted(cr flux.ColReader, cols []int) bool {
	// Check if the array is sorted by moving through each element and ensuring
	// that the previous one is greater than or equal to it.
	// We do not use the sort package for this because the sort package requires
	// some form of slice to operate and we want to avoid the allocations if
	// possible.
	//
	// In the future, it might be possible to skip this method and the sort method
	// if we can learn from the planner that each individual buffer is sorted.
	for i, n := 1, cr.Len(); i < n; i++ {
		for _, col := range cols {
			arr := table.Values(cr, col)
			if cmp := s.compare(arr, arr, i-1, i); cmp > 0 {
				// Not sorted return false.
				return false
			} else if cmp > 0 {
				// Sorted so move to the next row.
				break
			}
		}

		// If we get here by exiting the for loop normally, that means
		// everything was equal so technically sorted. Continue to the next row.
	}

	// If we get here, then the buffer is sorted.
	return true
}

func (s *sortTransformation) sort(cr flux.ColReader, cols []int) *array.Int {
	// Construct the indices.
	indices := mutable.NewInt64Array(s.mem)
	indices.Resize(cr.Len())

	// Retrieve the raw slice and initialize the offsets.
	offsets := indices.Int64Values()
	for i := range offsets {
		offsets[i] = int64(i)
	}

	// Sort the offsets by using the comparison method.
	sort.SliceStable(offsets, func(i, j int) bool {
		i, j = int(offsets[i]), int(offsets[j])
		for _, col := range cols {
			arr := table.Values(cr, col)
			if cmp := s.compare(arr, arr, i, j); cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})

	// Return the now sorted indices.
	return indices.NewInt64Array()
}

func (s *sortTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return s.d.RetractTable(key)
}
func (s *sortTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return s.d.UpdateWatermark(t)
}
func (s *sortTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return s.d.UpdateProcessingTime(t)
}
func (s *sortTransformation) Finish(id execute.DatasetID, err error) {
	s.d.Finish(err)
}

type sortTableMergeHeapItem struct {
	cr        flux.ColReader
	indices   *array.Int
	i, offset int
}

func (s *sortTableMergeHeapItem) Next() bool {
	s.i++
	if s.i >= s.cr.Len() {
		return false
	}
	s.offset = s.i
	if s.indices != nil {
		s.offset = int(s.indices.Value(s.i))
	}
	return true
}

func (s *sortTableMergeHeapItem) Release() {
	if s.indices != nil {
		s.indices.Release()
		s.indices = nil
	}
	if s.cr != nil {
		s.cr.Release()
		s.cr = nil
	}
}

type sortTableMergeHeap struct {
	key      flux.GroupKey
	cols     []flux.ColMeta
	items    []*sortTableMergeHeapItem
	sortCols []int
	compare  arrowutil.CompareFunc
}

func (s *sortTableMergeHeap) Len() int {
	return len(s.items)
}

func (s *sortTableMergeHeap) Less(i, j int) bool {
	x, y := s.items[i], s.items[j]
	for _, i := range s.sortCols {
		left := table.Values(x.cr, i)
		right := table.Values(y.cr, i)
		if cmp := s.compare(left, right, x.offset, y.offset); cmp != 0 {
			return cmp < 0
		}
	}
	return false
}

func (s *sortTableMergeHeap) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s *sortTableMergeHeap) Push(x interface{}) {
	s.items = append(s.items, x.(*sortTableMergeHeapItem))
}

func (s *sortTableMergeHeap) Pop() interface{} {
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

func (s *sortTableMergeHeap) ValueLen() int {
	var n int
	for _, item := range s.items {
		n += item.cr.Len() - item.i
	}
	return n
}

func (s *sortTableMergeHeap) Table(limit int, mem memory.Allocator) (flux.Table, error) {
	if s.ValueLen() == 0 {
		// Degenerate case where there are no rows to merge sort.
		for len(s.items) > 0 {
			s.Pop()
		}
		return execute.NewEmptyTable(s.key, s.cols), nil
	}

	// Construct the buffered builder that will contain the full table.
	builder := table.NewBufferedBuilder(s.key, mem)

	// Initialize the heap now that we have all of the data.
	heap.Init(s)

	// Initialize the builders. Due to the nature of how arrow builders
	// work, we can reuse these builders for every buffer as they
	// automatically reset.
	builders := make([]array.Builder, len(s.cols))
	for i, col := range s.cols {
		if s.key.HasCol(col.Label) {
			continue
		}
		builders[i] = arrow.NewBuilder(col.Type, mem)
	}
	defer func() {
		for _, b := range builders {
			if b != nil {
				b.Release()
			}
		}
	}()

	// Initialize space for the key values.
	// We will initialize these on the first buffer and then reuse
	// them after that to conserve space.
	keys := make([]array.Array, len(s.key.Cols()))
	defer func() {
		for _, key := range keys {
			if key != nil {
				key.Release()
			}
		}
	}()

	// Continue merging the tables until there are none.
	for len(s.items) > 0 && limit != 0 {
		n := s.ValueLen()
		if n > table.BufferSize {
			n = table.BufferSize
		}
		if limit > 0 && n > limit {
			n = limit
		}

		buffer := s.NextBuffer(builders, keys, n, mem)
		if limit > 0 {
			limit -= buffer.Len()
		}
		if err := builder.AppendBuffer(&buffer); err != nil {
			buffer.Release()
			return nil, err
		}
		buffer.Release()
	}

	// Release the remaining items and clear the items.
	// There are either none left or the remaining ones were filtered.
	for _, item := range s.items {
		item.Release()
	}
	s.items = s.items[:0]

	// Determine the next buffer size.
	return builder.Table()
}

func (s *sortTableMergeHeap) NextBuffer(builders []array.Builder, keys []array.Array, n int, mem memory.Allocator) arrow.TableBuffer {
	// Ensure there is enough space in each builder
	for _, b := range builders {
		if b == nil {
			continue
		}
		b.Resize(n)
	}

	for i := 0; i < n; i++ {
		// Append the next row to the builders.
		item := s.items[0]
		for j, b := range builders {
			if b == nil {
				continue
			}
			arrowutil.CopyValue(b, table.Values(item.cr, j), item.offset)
		}

		// Move to the next row.
		if item.Next() {
			// Fix the heap to ensure the next value.
			heap.Fix(s, 0)
		} else {
			// Remove this item from the heap since it
			// no longer has anymore rows.
			item.Release()
			heap.Pop(s)
		}
	}

	// Initialize the key buffers if they need to be.
	for i := range keys {
		if keys[i] == nil {
			keys[i] = arrow.Repeat(s.key.Cols()[i].Type, s.key.Value(i), n, mem)
		}
	}

	// Create the table buffer by merging our builders with the keys.
	buffer := arrow.TableBuffer{
		GroupKey: s.key,
		Columns:  s.cols,
		Values:   make([]array.Array, len(s.cols)),
	}
	for i, col := range s.cols {
		if builders[i] == nil {
			idx := execute.ColIdx(col.Label, s.key.Cols())
			arr := keys[idx]
			if arr.Len() > n {
				arr = arrow.Slice(arr, 0, int64(n))
			} else {
				arr.Retain()
			}
			buffer.Values[i] = arr
			continue
		}
		buffer.Values[i] = builders[i].NewArray()
	}
	return buffer
}

// RemoveRedundantSort is a planner rule that will remove a sort
// node from the graph if its input is already sorted.
type RemoveRedundantSort struct {
}

var _ plan.Rule = RemoveRedundantSort{}

func (r RemoveRedundantSort) Name() string {
	return "universe/RemoveRedundantSort"
}

func (r RemoveRedundantSort) Pattern() plan.Pattern {
	// Predecessor to the sort must have only the single successor (the sort node)
	// to work around https://github.com/influxdata/flux/issues/5044
	return plan.MultiSuccessor(SortKind, plan.AnySingleSuccessor())
}

func (r RemoveRedundantSort) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	pred := node.Predecessors()[0]
	inputCollation := plan.GetOutputAttribute(pred, plan.CollationKey)
	if inputCollation == nil {
		return node, false, nil // input to sort is not already sorted
	}

	sortSpec := node.ProcedureSpec().(*SortProcedureSpec)
	sortCollation := sortSpec.OutputAttributes()[plan.CollationKey]
	if sortCollation.SatisfiedBy(inputCollation) {
		return pred, true, nil
	}

	return node, false, nil
}
