package universe

import (
	"math"
	"sort"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
)

const WindowKind = "window"

type WindowOpSpec struct {
	Every       flux.Duration `json:"every"`
	Period      flux.Duration `json:"period"`
	Offset      flux.Duration `json:"offset"`
	TimeColumn  string        `json:"timeColumn"`
	StopColumn  string        `json:"stopColumn"`
	StartColumn string        `json:"startColumn"`
	CreateEmpty bool          `json:"createEmpty"`
}

var infinityVar = values.NewDuration(values.ConvertDuration(math.MaxInt64))

func init() {
	windowSignature := flux.LookupBuiltInType("universe", "window")

	flux.RegisterPackageValue("universe", WindowKind, flux.MustValue(flux.FunctionValue(WindowKind, createWindowOpSpec, windowSignature)))
	flux.RegisterOpSpec(WindowKind, newWindowOp)
	flux.RegisterPackageValue("universe", "inf", infinityVar)
	plan.RegisterProcedureSpec(WindowKind, newWindowProcedure, WindowKind)
	plan.RegisterPhysicalRules(WindowTriggerPhysicalRule{})
	execute.RegisterTransformation(WindowKind, createWindowTransformation)
}

func createWindowOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(WindowOpSpec)
	every, everySet, err := args.GetDuration("every")
	if err != nil {
		return nil, err
	}
	if everySet {
		spec.Every = every
	}
	period, periodSet, err := args.GetDuration("period")
	if err != nil {
		return nil, err
	}
	if periodSet {
		spec.Period = period
	}
	if offset, ok, err := args.GetDuration("offset"); err != nil {
		return nil, err
	} else if ok {
		spec.Offset = offset
	}

	if !everySet && !periodSet {
		return nil, errors.New(codes.Invalid, `window function requires at least one of "every" or "period" to be set`)
	}

	if label, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = label
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}
	if label, ok, err := args.GetString("startColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.StartColumn = label
	} else {
		spec.StartColumn = execute.DefaultStartColLabel
	}
	if label, ok, err := args.GetString("stopColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.StopColumn = label
	} else {
		spec.StopColumn = execute.DefaultStopColLabel
	}
	if createEmpty, ok, err := args.GetBool("createEmpty"); err != nil {
		return nil, err
	} else if ok {
		spec.CreateEmpty = createEmpty
	} else {
		spec.CreateEmpty = false
	}

	// Apply defaults
	if !everySet {
		spec.Every = spec.Period
	}
	if !periodSet {
		spec.Period = spec.Every
	}
	return spec, nil
}

func newWindowOp() flux.OperationSpec {
	return new(WindowOpSpec)
}

func (s *WindowOpSpec) Kind() flux.OperationKind {
	return WindowKind
}

type WindowProcedureSpec struct {
	plan.DefaultCost
	Window plan.WindowSpec
	TimeColumn,
	StartColumn,
	StopColumn string
	CreateEmpty bool
}

func newWindowProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*WindowOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	p := &WindowProcedureSpec{
		Window: plan.WindowSpec{
			Every:  s.Every,
			Period: s.Period,
			Offset: s.Offset,
		},
		TimeColumn:  s.TimeColumn,
		StartColumn: s.StartColumn,
		StopColumn:  s.StopColumn,
		CreateEmpty: s.CreateEmpty,
	}
	return p, nil
}

func (s *WindowProcedureSpec) Kind() plan.ProcedureKind {
	return WindowKind
}
func (s *WindowProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WindowProcedureSpec)
	ns.Window = s.Window
	return ns
}

func createWindowTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WindowProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	bounds := a.StreamContext().Bounds()
	if bounds == nil {
		return nil, nil, errors.New(codes.Invalid, "nil bounds passed to window")
	}

	w, err := execute.NewWindow(
		s.Window.Every,
		s.Window.Period,
		s.Window.Offset,
	)
	if err != nil {
		return nil, nil, err
	}
	t, d := NewFixedWindowTransformation(
		id,
		*bounds,
		w,
		s.TimeColumn,
		s.StartColumn,
		s.StopColumn,
		s.CreateEmpty,
		a.Allocator(),
	)
	return t, d, nil
}

func NewFixedWindowTransformation(
	id execute.DatasetID,
	bounds execute.Bounds,
	w execute.Window,
	timeCol,
	startCol,
	stopCol string,
	createEmpty bool,
	mem *memory.Allocator,
) (execute.Transformation, execute.Dataset) {
	if w.Every == infinityVar.Duration() {
		return newInfinityWindowTransformation(id, bounds, timeCol, startCol, stopCol, mem)
	}
	return newFixedWindowTransformation(id, bounds, w, timeCol, startCol, stopCol, createEmpty, mem)
}

// WindowTriggerPhysicalRule rewrites a physical window operation
// to use a narrow trigger if certain conditions are met.
type WindowTriggerPhysicalRule struct{}

func (WindowTriggerPhysicalRule) Name() string {
	return "WindowTriggerPhysicalRule"
}

// Pattern matches the physical operator pattern consisting of a window
// operator with a single predecessor of any kind.
func (WindowTriggerPhysicalRule) Pattern() plan.Pattern {
	return plan.PhysPat(WindowKind, plan.Any())
}

// Rewrite modifies a window's trigger spec so long as it doesn't have any
// window descendents that occur earlier in the plan and as long as none
// of its descendents merge multiple streams together like union and join.
func (WindowTriggerPhysicalRule) Rewrite(window plan.Node) (plan.Node, bool, error) {
	// This rule's pattern ensures us only one predecessor
	if !hasValidPredecessors(window.Predecessors()[0]) {
		return window, false, nil
	}
	// This rule's pattern ensures us a physical operator
	ppn := window.(*plan.PhysicalPlanNode)
	if ppn.TriggerSpec != nil {
		return ppn, false, nil
	}
	ppn.TriggerSpec = plan.NarrowTransformationTriggerSpec{}
	return ppn, true, nil
}

func hasValidPredecessors(node plan.Node) bool {
	pred := node.Predecessors()
	// Source nodes might not produce uniform time bounds for all
	// tables in which case we can't optimize window. However if a
	// source is a BoundsAwareProcedureSpec then it must produce
	// bounded data in which case we can perform the optimization.
	if len(pred) == 0 {
		s := node.ProcedureSpec()
		n, ok := s.(plan.BoundsAwareProcedureSpec)
		return ok && n.TimeBounds(nil) != nil
	}
	kind := node.Kind()
	switch kind {
	// Range gives the same static time bounds to the entire stream,
	// so no need to recurse.
	case RangeKind:
		return true
	case ColumnsKind,
		CumulativeSumKind,
		DerivativeKind,
		DifferenceKind,
		DistinctKind,
		FilterKind,
		FirstKind,
		GroupKind,
		KeyValuesKind,
		KeysKind,
		LastKind,
		LimitKind,
		MaxKind,
		MinKind,
		ExactQuantileSelectKind,
		SampleKind,
		DropKind,
		KeepKind,
		DuplicateKind,
		RenameKind,
		ShiftKind,
		SortKind,
		StateTrackingKind,
		UniqueKind:
	default:
		return false
	}
	if len(pred) == 1 {
		return hasValidPredecessors(pred[0])
	}
	return false
}

type windowTransformationBase struct {
	d      execute.Dataset
	cache  table.BuilderCache
	bounds execute.Bounds
	mem    *memory.Allocator

	timeCol,
	startCol,
	stopCol string
}

// newWindowGroupKey will return a new group key that either adds or modifies the existing start
// and stop column to match the bounds of the window.
func (w *windowTransformationBase) newWindowGroupKey(key flux.GroupKey, bnds execute.Bounds) flux.GroupKey {
	// Construct the group key schema.
	cols, startIdx, stopIdx := w.createSchema(key.Cols())

	// Make a copy of the values and replace the start and stop.
	vs := make([]values.Value, len(cols))
	for j := range cols {
		if j == startIdx {
			vs[j] = values.NewTime(bnds.Start)
		} else if j == stopIdx {
			vs[j] = values.NewTime(bnds.Stop)
		} else {
			// This will always be the same index as the old location
			// since the key is always additive.
			vs[j] = key.Value(j)
		}
	}
	return execute.NewGroupKey(cols, vs)
}

// createSchema will create a compatible schema for the given column metadata.
func (w *windowTransformationBase) createSchema(cols []flux.ColMeta) ([]flux.ColMeta, int, int) {
	startIdx := execute.ColIdx(w.startCol, cols)
	stopIdx := execute.ColIdx(w.stopCol, cols)
	if (startIdx < 0 || cols[startIdx].Type != flux.TTime) ||
		(stopIdx < 0 || cols[stopIdx].Type != flux.TTime) {
		newCols := make([]flux.ColMeta, len(cols), len(cols)+2)
		copy(newCols, cols)
		cols = newCols
		if startIdx < 0 {
			cols = append(cols, flux.ColMeta{
				Label: w.startCol,
				Type:  flux.TTime,
			})
			startIdx = len(cols) - 1
		} else {
			cols[startIdx].Type = flux.TTime
		}
		if stopIdx < 0 {
			cols = append(cols, flux.ColMeta{
				Label: w.stopCol,
				Type:  flux.TTime,
			})
			stopIdx = len(cols) - 1
		} else {
			cols[stopIdx].Type = flux.TTime
		}
	}
	return cols, startIdx, stopIdx
}

func (w *windowTransformationBase) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return w.d.RetractTable(key)
}

func (w *windowTransformationBase) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return w.d.UpdateWatermark(t)
}

func (w *windowTransformationBase) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return w.d.UpdateProcessingTime(t)
}

func (w *windowTransformationBase) Finish(id execute.DatasetID, err error) {
	w.d.Finish(err)
}

type infinityWindowTransformation struct {
	windowTransformationBase
}

func newInfinityWindowTransformation(id execute.DatasetID, bounds execute.Bounds, timeCol, startCol, stopCol string, mem *memory.Allocator) (execute.Transformation, execute.Dataset) {
	t := &infinityWindowTransformation{
		windowTransformationBase: windowTransformationBase{
			cache: table.BuilderCache{
				New: func(key flux.GroupKey) table.Builder {
					return table.NewBufferedBuilder(key, mem)
				},
			},
			bounds:   bounds,
			mem:      mem,
			timeCol:  timeCol,
			startCol: startCol,
			stopCol:  stopCol,
		},
	}
	t.d = table.NewDataset(id, &t.cache)
	return t, t.d
}

func (w *infinityWindowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key := w.newWindowGroupKey(tbl.Key(), w.bounds)
	b, _ := table.GetBufferedBuilder(key, &w.cache)

	cols, startIdx, stopIdx := w.createSchema(tbl.Cols())
	return tbl.Do(func(cr flux.ColReader) error {
		buffer := &arrow.TableBuffer{
			GroupKey: key,
			Columns:  cols,
			Values:   make([]array.Interface, len(cols)),
		}

		l := cr.Len()
		for j := range cols {
			var vs array.Interface
			// TODO(jsternberg): These created arrays can be cached and shared.
			if j == startIdx {
				ts := values.NewTime(w.bounds.Start)
				vs = arrow.Repeat(ts, l, w.mem)
			} else if j == stopIdx {
				ts := values.NewTime(w.bounds.Stop)
				vs = arrow.Repeat(ts, l, w.mem)
			} else {
				vs = table.Values(cr, j)
				vs.Retain()
			}
			buffer.Values[j] = vs
		}
		return b.AppendBuffer(buffer)
	})
}

type fixedWindowTransformation struct {
	windowTransformationBase
	w           execute.Window
	createEmpty bool
}

func newFixedWindowTransformation(id execute.DatasetID, bounds execute.Bounds, w execute.Window, timeCol, startCol, stopCol string, createEmpty bool, mem *memory.Allocator) (execute.Transformation, execute.Dataset) {
	t := &fixedWindowTransformation{
		windowTransformationBase: windowTransformationBase{
			cache: table.BuilderCache{
				New: func(key flux.GroupKey) table.Builder {
					return table.NewBufferedBuilder(key, mem)
				},
				Tables: execute.NewRandomAccessGroupLookup(),
			},
			bounds:   bounds,
			mem:      mem,
			timeCol:  timeCol,
			startCol: startCol,
			stopCol:  stopCol,
		},
		w:           w,
		createEmpty: createEmpty,
	}
	t.d = table.NewDataset(id, &t.cache)
	return t, t.d
}

func (w *fixedWindowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	timeIdx := execute.ColIdx(w.timeCol, tbl.Cols())
	if timeIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "missing time column %q", w.timeCol)
	} else if col := tbl.Cols()[timeIdx]; col.Type != flux.TTime {
		return errors.Newf(codes.FailedPrecondition, "time column %q must be of type time, is %s", col.Label, col.Type)
	}

	// Generate the window boundaries if we have been told to create
	// empty windows.
	var bounds []execute.Bounds
	if w.createEmpty {
		bounds = w.getAllWindowBounds()
	}

	// Precreate the tables if create empty is true.
	for _, bnds := range bounds {
		key := w.newWindowGroupKey(tbl.Key(), bnds)
		if b, created := table.GetBufferedBuilder(key, &w.cache); created {
			b.Columns, _, _ = w.createSchema(tbl.Cols())
		}
	}

	// Process the table and insert rows into the appropriate boundaries.
	return tbl.Do(func(cr flux.ColReader) error {
		// If the buffer is empty, skip it.
		if cr.Len() == 0 {
			return nil
		}

		// Retrieve the time index and sort the time input by the indices.
		ts := cr.Times(timeIdx)
		indices := w.getSortedIndices(ts)

		// Create the window boundaries encountered in this buffer
		// if we are not creating empty boundaries.
		if !w.createEmpty {
			bounds = w.getWindowBoundsFromTimes(ts, indices)
		}

		// Iterate through each of the window boundaries
		// and create a new table for each one.
		// The algorithm will create one boundary at a time to
		// reduce the number of table lookups required to be created.
		// The start location is where we believe the first element
		// for the boundary starts. We iterate until we find the first
		// and final element that fit into the boundary. We then
		// record the start so that we can speed up the next iteration.
		// This is possible because both the times and the boundaries
		// are sorted.
		start := 0
		for _, b := range bounds {
			// If the current start index is after the first
			// boundary, we will never find a time that is within
			// the current boundary so abandon this boundary.
			if tm := execute.Time(ts.Value(indices[start])); tm >= b.Stop {
				continue
			}

			// Find the start index.
			for l := len(indices); start < l; start++ {
				tm := execute.Time(ts.Value(indices[start]))
				if b.Contains(tm) {
					break
				}
			}

			// If the start index is after the index, exit here.
			// This shouldn't be possible.
			if start >= len(indices) {
				break
			}

			// Determine the first index that is not in the
			// boundary.
			end := start + 1
			for l := len(indices); end < l; end++ {
				tm := execute.Time(ts.Value(indices[end]))
				if !b.Contains(tm) {
					break
				}
			}

			// Append the rows to the appropriate table buffer.
			if err := w.appendRows(cr, b, indices, start, end); err != nil {
				return err
			}
		}
		return nil
	})
}

// getSortedIndices will return a list of sorted indices for the array.
func (w *fixedWindowTransformation) getSortedIndices(vs *array.Int64) []int {
	indices := make([]int, vs.Len())

	// Add all of the valid indices to the array in order.
	// While we are doing this, check to see if we are sorted.
	sorted := true
	for i, l := 0, vs.Len(); i < l; i++ {
		// TODO(jsternberg): The original window function didn't differentiate
		// between null values so we're keeping that same behavior, but that
		// is obviously not correct.
		indices[i] = i
		if i >= 1 && sorted {
			if vs.Value(i-1) > vs.Value(i) {
				sorted = false
			}
		}
	}

	// If we are not sorted, then sort the indices by their value.
	if !sorted {
		sort.SliceStable(indices, func(i, j int) bool {
			idx, jdx := indices[i], indices[j]
			return vs.Value(idx) < vs.Value(jdx)
		})
	}

	// Return the indices in sorted order.
	return indices
}

func (w *fixedWindowTransformation) getAllWindowBounds() (bounds []execute.Bounds) {
	bs := w.w.GetOverlappingBounds(w.bounds)
	w.clipBounds(bs)
	return bs
}

func (w *fixedWindowTransformation) getWindowBoundsFromTimes(ts *array.Int64, indices []int) (bounds []execute.Bounds) {
	// TODO(jsternberg): Implement create empty.
	// Iterate through each time and generate window boundaries.
	seen := make(map[execute.Bounds]bool)
	for _, i := range indices {
		tm := execute.Time(ts.Value(i))
		for _, b := range w.getWindowBounds(tm) {
			if seen[b] {
				continue
			}
			bounds = append(bounds, b)
			seen[b] = true
		}
	}
	return bounds
}

func (w *fixedWindowTransformation) clipBounds(bs []execute.Bounds) {
	for i := range bs {
		bs[i] = w.bounds.Intersect(bs[i])
	}
}

func (w *fixedWindowTransformation) getWindowBounds(tm execute.Time) []execute.Bounds {
	bs := w.w.GetOverlappingBounds(execute.Bounds{Start: tm, Stop: tm + 1})
	w.clipBounds(bs)
	return bs
}

func (w *fixedWindowTransformation) appendRows(cr flux.ColReader, bnds execute.Bounds, indices []int, start, stop int) error {
	// Construct the group key and the columns.
	key := w.newWindowGroupKey(cr.Key(), bnds)
	cols, startIdx, stopIdx := w.createSchema(cr.Cols())

	buffer := &arrow.TableBuffer{
		GroupKey: key,
		Columns:  cols,
		Values:   make([]array.Interface, len(cols)),
	}

	l := stop - start
	for j, c := range buffer.Columns {
		var vs array.Interface
		if l == 0 {
			vs = arrow.NewBuilder(c.Type, w.mem).NewArray()
		} else if j == startIdx {
			ts := values.NewTime(bnds.Start)
			vs = arrow.Repeat(ts, l, w.mem)
		} else if j == stopIdx {
			ts := values.NewTime(bnds.Stop)
			vs = arrow.Repeat(ts, l, w.mem)
		} else {
			vs = w.appendRowsForColumn(cr, j, indices, start, stop)
		}
		buffer.Values[j] = vs
	}

	// Find the buffered table builder in the cache and append this buffer.
	b, _ := table.GetBufferedBuilder(key, &w.cache)
	return b.AppendBuffer(buffer)
}

func (w *fixedWindowTransformation) appendRowsForColumn(cr flux.ColReader, j int, indices []int, start, stop int) array.Interface {
	col := cr.Cols()[j]
	if idx := execute.ColIdx(col.Label, cr.Key().Cols()); idx >= 0 {
		// If this is part of the group key, all of the values are
		// the same and the indices don't matter. Just create a slice
		// because this doesn't matter.
		return arrow.Slice(table.Values(cr, j), int64(start), int64(stop))
	}

	l := stop - start
	switch col.Type {
	case flux.TInt:
		b := array.NewInt64Builder(w.mem)
		b.Resize(l)

		vs := cr.Ints(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	case flux.TUInt:
		b := array.NewUint64Builder(w.mem)
		b.Resize(l)

		vs := cr.UInts(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	case flux.TFloat:
		b := array.NewFloat64Builder(w.mem)
		b.Resize(l)

		vs := cr.Floats(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	case flux.TString:
		b := arrow.NewStringBuilder(w.mem)
		b.Resize(l)

		vs := cr.Strings(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	case flux.TBool:
		b := array.NewBooleanBuilder(w.mem)
		b.Resize(l)

		vs := cr.Bools(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	case flux.TTime:
		b := array.NewInt64Builder(w.mem)
		b.Resize(l)

		vs := cr.Times(j)
		for i := start; i < stop; i++ {
			idx := indices[i]
			if vs.IsNull(idx) {
				b.AppendNull()
				continue
			}
			b.Append(vs.Value(idx))
		}
		return b.NewArray()
	default:
		panic(errors.Newf(codes.Internal, "unknown column type: %s", col.Type))
	}
}
