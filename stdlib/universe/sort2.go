package universe

import (
	"container/heap"
	"context"
	"sort"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/plan"
)

type sortTransformation2 struct {
	execute.ExecutionNode
	d       *execute.PassthroughDataset
	mem     memory.Allocator
	cols    []string
	compare arrowutil.CompareFunc
}

func newSortTransformation2(id execute.DatasetID, spec *SortProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	t := &sortTransformation2{
		d:       execute.NewPassthroughDataset(id),
		mem:     a.Allocator(),
		cols:    spec.Columns,
		compare: arrowutil.Compare,
	}
	if spec.Desc {
		// If descending, use the descending comparison.
		t.compare = arrowutil.CompareDesc
	}
	return t, t.d, nil
}

func (s *sortTransformation2) Process(id execute.DatasetID, tbl flux.Table) error {
	sortCols := make([]int, 0, len(s.cols))
	for _, col := range s.cols {
		if idx := execute.ColIdx(col, tbl.Cols()); idx >= 0 {
			// If the sort key is part of the group key, skip it anyway.
			// They are all sorted anyway.
			if tbl.Key().HasCol(col) {
				continue
			}
			sortCols = append(sortCols, idx)
		}
	}

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

	out, err := mh.Table(s.mem)
	if err != nil {
		return err
	}
	return s.d.Process(out)
}

func (s *sortTransformation2) processView(mh *sortTableMergeHeap, cr flux.ColReader) error {
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

func (s *sortTransformation2) isSorted(cr flux.ColReader, cols []int) bool {
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

func (s *sortTransformation2) sort(cr flux.ColReader, cols []int) *array.Int64 {
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

func (s *sortTransformation2) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return s.d.RetractTable(key)
}
func (s *sortTransformation2) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return s.d.UpdateWatermark(t)
}
func (s *sortTransformation2) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return s.d.UpdateProcessingTime(t)
}
func (s *sortTransformation2) Finish(id execute.DatasetID, err error) {
	s.d.Finish(err)
}

type sortTableMergeHeapItem struct {
	cr        flux.ColReader
	indices   *array.Int64
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

func (s *sortTableMergeHeap) Table(mem memory.Allocator) (flux.Table, error) {
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
	keys := make([]array.Interface, len(s.key.Cols()))
	defer func() {
		for _, key := range keys {
			if key != nil {
				key.Release()
			}
		}
	}()

	// Continue merging the tables until there are none.
	for len(s.items) > 0 {
		n := s.ValueLen()
		if n > table.BufferSize {
			n = table.BufferSize
		}

		buffer := s.NextBuffer(builders, keys, n, mem)
		if err := builder.AppendBuffer(&buffer); err != nil {
			buffer.Release()
			return nil, err
		}
		buffer.Release()
	}

	// Determine the next buffer size.
	return builder.Table()
}

func (s *sortTableMergeHeap) NextBuffer(builders []array.Builder, keys []array.Interface, n int, mem memory.Allocator) arrow.TableBuffer {
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
			keys[i] = arrow.Repeat(s.key.Value(i), n, mem)
		}
	}

	// Create the table buffer by merging our builders with the keys.
	buffer := arrow.TableBuffer{
		GroupKey: s.key,
		Columns:  s.cols,
		Values:   make([]array.Interface, len(s.cols)),
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

type OptimizeSortRule struct{}

func (r OptimizeSortRule) Name() string {
	return "OptimizeSortRule"
}

func (r OptimizeSortRule) Pattern() plan.Pattern {
	return plan.Pat(SortKind, plan.Any())
}

func (r OptimizeSortRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	sortSpec := node.ProcedureSpec().(*SortProcedureSpec)
	if sortSpec.Optimize {
		return node, false, nil
	}
	sortSpec = sortSpec.Copy().(*SortProcedureSpec)
	sortSpec.Optimize = true
	if err := node.ReplaceSpec(sortSpec); err != nil {
		return node, false, err
	}
	return node, true, nil
}
