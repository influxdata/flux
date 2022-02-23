package universe

import (
	"context"
	"math"
	"sort"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/plan"
	experimentaltable "github.com/influxdata/flux/stdlib/experimental/table"
	"github.com/influxdata/flux/values"
)

func init() {
	plan.RegisterPhysicalRules(
		AggregateWindowRule{},
		AggregateWindowCreateEmptyRule{},
	)
	execute.RegisterTransformation(AggregateWindowKind, createAggregateWindowTransformation)
}

const AggregateWindowKind = "aggregateWindow"

type AggregateWindowProcedureSpec struct {
	plan.DefaultCost
	spec      *WindowProcedureSpec
	aggregate aggregateWindow
	valueCol  string
	useStart  bool
}

func (s *AggregateWindowProcedureSpec) Kind() plan.ProcedureKind {
	return AggregateWindowKind
}

func (s *AggregateWindowProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.spec = ns.spec.Copy().(*WindowProcedureSpec)
	return &ns
}

type aggregateWindowState struct {
	ts      *array.Int
	buffers []array.Interface
}

type aggregateWindow interface {
	Initialize(valueType flux.ColType, mem memory.Allocator) ([]array.Builder, error)
	Aggregate(ts, indices, start, stop *array.Int, values array.Interface, builders []array.Builder)
	Merge(prevT, nextT *array.Int, prev, next []array.Interface, mem memory.Allocator) (*array.Int, []array.Interface)
	Compute(buffers []array.Interface) (flux.ColType, array.Interface)
	AppendEmpty(b array.Builder)
}

type aggregateWindowTransformation struct {
	w           interval.Window
	bounds      *execute.Bounds
	createEmpty bool
	timeCol     string
	valueCol    string
	useStart    bool
	aggregate   aggregateWindow
}

func createAggregateWindowTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*AggregateWindowProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	bounds := a.StreamContext().Bounds()
	if bounds == nil {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#nil-bounds-passed-to-window"
		return nil, nil, errors.New(codes.Invalid, "nil bounds passed to window; use range to set the window range").
			WithDocURL(docURL)
	}

	loc, err := s.spec.Window.LoadLocation()
	if err != nil {
		return nil, nil, err
	}

	w, err := interval.NewWindowInLocation(
		s.spec.Window.Every,
		s.spec.Window.Period,
		s.spec.Window.Offset,
		loc,
	)
	if err != nil {
		return nil, nil, err
	}

	tr := &aggregateWindowTransformation{
		w:           w,
		bounds:      bounds,
		createEmpty: s.spec.CreateEmpty,
		timeCol:     s.spec.TimeColumn,
		valueCol:    s.valueCol,
		useStart:    s.useStart,
		aggregate:   s.aggregate,
	}
	return execute.NewAggregateTransformation(id, tr, a.Allocator())
}

func (a *aggregateWindowTransformation) Aggregate(chunk table.Chunk, state interface{}, mem memory.Allocator) (interface{}, bool, error) {
	ws, _ := state.(*aggregateWindowState)
	newState, err := a.processChunk(chunk, mem)
	if err != nil {
		return nil, false, err
	}

	s, err := a.mergeWindows(ws, newState, mem)
	if err != nil {
		return nil, false, err
	}
	return s, true, nil
}

func (a *aggregateWindowTransformation) Compute(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
	ws := state.(*aggregateWindowState)
	key = a.recomputeKey(key)
	buffer := a.computeFromState(key, ws, mem)
	if err := buffer.Validate(); err != nil {
		return err
	}

	chunk := table.ChunkFromBuffer(buffer)
	return d.Process(chunk)
}

func (a *aggregateWindowTransformation) processChunk(chunk table.Chunk, mem memory.Allocator) (*aggregateWindowState, error) {
	// Find the time column for this table chunk.
	ts, err := a.getTimeColumn(chunk)
	if err != nil {
		return nil, err
	}

	// Find the value column for this table chunk.
	vt, vs, err := a.getValueColumn(chunk)
	if err != nil {
		return nil, err
	}

	// Sort the timestamps and return the
	// offsets of the sorted timestamps.
	indices := a.sort(ts, mem)
	defer indices.Release()

	// Scan the timestamps and construct the window boundaries.
	start, stop := a.scanWindows(ts, indices, mem)
	buffers, err := a.computeWindows(ts, indices, start, stop, vt, vs, mem)
	if err != nil {
		start.Release()
		stop.Release()
		return nil, err
	}

	if a.useStart {
		stop.Release()
		stop = start
	} else {
		start.Release()
	}
	return &aggregateWindowState{
		ts:      stop,
		buffers: buffers,
	}, nil
}

func (a *aggregateWindowTransformation) getTimeColumn(chunk table.Chunk) (*array.Int, error) {
	idx := chunk.Index(a.timeCol)
	if idx < 0 {
		return nil, errors.Newf(codes.FailedPrecondition, "no time column: %s", a.timeCol)
	}

	if colType := chunk.Col(idx).Type; colType != flux.TTime {
		return nil, errors.Newf(codes.FailedPrecondition, "time column is not a time value: %s", colType)
	}
	return chunk.Ints(idx), nil
}

func (a *aggregateWindowTransformation) getValueColumn(chunk table.Chunk) (flux.ColType, array.Interface, error) {
	idx := chunk.Index(a.valueCol)
	if idx < 0 {
		return flux.TInvalid, nil, errors.Newf(codes.FailedPrecondition, "column %q does not exist", a.valueCol)
	}

	if chunk.Key().HasCol(a.valueCol) {
		return flux.TInvalid, nil, errors.New(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
	}
	return chunk.Col(idx).Type, chunk.Values(idx), nil
}

// sort will return the indexes of the array as if it were sorted.
// It does not modify the array and the array returned are the indexes of the
// sorted values.
func (a *aggregateWindowTransformation) sort(ts *array.Int, mem memory.Allocator) *array.Int {
	// Construct a mutable array builder so that we can modify the buffer in-place
	// while still using memory accounting.
	indexes := mutable.NewInt64Array(mem)
	indexes.Resize(ts.Len())

	// Retrieve the raw slice.
	offsets := indexes.Int64Values()
	for i := range offsets {
		offsets[i] = int64(i)
	}

	// Sort the offsets by using the values in the timestamp array.
	sort.SliceStable(offsets, func(i, j int) bool {
		i, j = int(offsets[i]), int(offsets[j])
		// Maybe we should error if we see a null timestamp?
		// The original window doesn't seem to even check this case.
		// Nulls are considered greater than everything.
		if ts.IsNull(j) {
			return ts.IsValid(i)
		} else if ts.IsNull(i) {
			return false
		}
		return ts.Value(i) < ts.Value(j)
	})

	// Slice of null values from the index.
	arr := indexes.NewInt64Array()
	if nulls := ts.NullN(); nulls > 0 {
		narr := arrow.IntSlice(arr, 0, ts.Len()-nulls)
		arr.Release()
		arr = narr
	}
	return arr
}

// scanWindows scans the timestamps and returns the appropriate boundaries.
// Not all timestamps may be associated with a boundary and some timestamps may
// be associated with multiple boundaries.
func (a *aggregateWindowTransformation) scanWindows(ts, indices *array.Int, mem memory.Allocator) (start, stop *array.Int) {
	startB := array.NewIntBuilder(mem)
	stopB := array.NewIntBuilder(mem)
	latest := int64(math.MinInt64)

	var bounds []execute.Bounds
	for i, n := 0, indices.Len(); i < n; i++ {
		t := ts.Value(int(indices.Value(i)))

		bounds = bounds[:0]

		bound := a.w.GetLatestBounds(values.Time(t))
		for bound.Contains(values.Time(t)) {
			b := execute.Bounds{
				Start: bound.Start(),
				Stop:  bound.Stop(),
			}
			if int64(b.Start) <= latest {
				break
			}
			bounds = append(bounds, b)

			// Look at the previous boundary.
			bound = a.w.PrevBounds(bound)
		}

		if len(bounds) == 0 {
			continue
		}

		// Append the times going backwards.
		for j := len(bounds) - 1; j >= 0; j-- {
			b := a.bounds.Intersect(bounds[j])
			startB.Append(int64(b.Start))
			stopB.Append(int64(b.Stop))
		}
		latest = int64(bounds[0].Start)
	}
	return startB.NewIntArray(), stopB.NewIntArray()
}

func (a *aggregateWindowTransformation) computeWindows(ts, indices, start, stop *array.Int, valueType flux.ColType, values array.Interface, mem memory.Allocator) ([]array.Interface, error) {
	builders, err := a.aggregate.Initialize(valueType, mem)
	if err != nil {
		return nil, err
	}
	for _, b := range builders {
		b.Resize(start.Len())
	}
	a.aggregate.Aggregate(ts, indices, start, stop, values, builders)

	// Compute the windows for the current chunk.
	// This isn't the final computation that produces the column,
	// but an interim result that will be merged and computed later.
	results := make([]array.Interface, len(builders))
	for i, b := range builders {
		results[i] = b.NewArray()
	}
	return results, nil
}

func (a *aggregateWindowTransformation) mergeWindows(s, ns *aggregateWindowState, mem memory.Allocator) (*aggregateWindowState, error) {
	if s == nil {
		return ns, nil
	}
	ts, buffers := a.aggregate.Merge(s.ts, ns.ts, s.buffers, ns.buffers, mem)
	s.ts.Release()
	ns.ts.Release()
	ns.ts, ns.buffers = ts, buffers
	return ns, nil
}

func (a *aggregateWindowTransformation) keyMatches(key flux.GroupKey) bool {
	startIdx := execute.ColIdx(execute.DefaultStartColLabel, key.Cols())
	if startIdx < 0 || key.Cols()[startIdx].Type != flux.TTime {
		return false
	}

	if key.ValueTime(startIdx) != a.bounds.Start {
		return false
	}

	stopIdx := execute.ColIdx(execute.DefaultStopColLabel, key.Cols())
	if stopIdx < 0 || key.Cols()[stopIdx].Type != flux.TTime {
		return false
	}

	if key.ValueTime(stopIdx) != a.bounds.Stop {
		return false
	}
	return true
}

func (a *aggregateWindowTransformation) recomputeKey(key flux.GroupKey) flux.GroupKey {
	if a.keyMatches(key) {
		return key
	}

	// We have to recompute the key because it is different than the input.
	startIdx := execute.ColIdx(execute.DefaultStartColLabel, key.Cols())
	stopIdx := execute.ColIdx(execute.DefaultStopColLabel, key.Cols())
	ncols := len(key.Cols())
	if startIdx < 0 {
		ncols++
	}
	if stopIdx < 0 {
		ncols++
	}

	// Copy over the existing values before modifying the parts we need to modify.
	cols := make([]flux.ColMeta, len(key.Cols()), ncols)
	vs := make([]values.Value, len(cols), ncols)
	copy(cols, key.Cols())
	copy(vs, key.Values())

	if startIdx >= 0 {
		cols[startIdx].Type = flux.TTime
		vs[startIdx] = values.NewTime(a.bounds.Start)
	} else {
		cols = append(cols, flux.ColMeta{
			Label: execute.DefaultStartColLabel,
			Type:  flux.TTime,
		})
		vs = append(vs, values.NewTime(a.bounds.Start))
	}

	if stopIdx >= 0 {
		cols[stopIdx].Type = flux.TTime
		vs[stopIdx] = values.NewTime(a.bounds.Stop)
	} else {
		cols = append(cols, flux.ColMeta{
			Label: execute.DefaultStopColLabel,
			Type:  flux.TTime,
		})
		vs = append(vs, values.NewTime(a.bounds.Stop))
	}
	return execute.NewGroupKey(cols, vs)
}

func (a *aggregateWindowTransformation) computeFromState(key flux.GroupKey, s *aggregateWindowState, mem memory.Allocator) arrow.TableBuffer {
	ts := s.ts
	vt, vs := a.aggregate.Compute(s.buffers)

	if a.createEmpty {
		ts, vs = a.createEmptyWindows(ts, vs, mem)
	}
	n := ts.Len()

	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  make([]flux.ColMeta, 0, len(key.Cols())+2),
	}
	buffer.Values = make([]array.Interface, 0, cap(buffer.Columns))

	buffer.Columns = append(buffer.Columns, flux.ColMeta{
		Label: execute.DefaultTimeColLabel,
		Type:  flux.TTime,
	})
	buffer.Values = append(buffer.Values, ts)

	for j, col := range key.Cols() {
		buffer.Columns = append(buffer.Columns, col)
		buffer.Values = append(buffer.Values, arrow.Repeat(col.Type, key.Value(j), n, mem))
	}

	buffer.Columns = append(buffer.Columns, flux.ColMeta{
		Label: a.valueCol,
		Type:  vt,
	})
	buffer.Values = append(buffer.Values, vs)
	return buffer
}

func (a *aggregateWindowTransformation) createEmptyWindows(ts *array.Int, vs array.Interface, mem memory.Allocator) (*array.Int, array.Interface) {
	switch vs := vs.(type) {
	case *array.Int:
		vb := array.NewIntBuilder(mem)
		ts = a.createEmptyWindowsFunc(ts, vb, mem, func(i int) {
			vb.Append(vs.Value(i))
		})
		vs.Release()
		return ts, vb.NewArray()
	case *array.Float:
		vb := array.NewFloatBuilder(mem)
		ts = a.createEmptyWindowsFunc(ts, vb, mem, func(i int) {
			vb.Append(vs.Value(i))
		})
		vs.Release()
		return ts, vb.NewArray()
	case *array.Uint:
		vb := array.NewUintBuilder(mem)
		ts = a.createEmptyWindowsFunc(ts, vb, mem, func(i int) {
			vb.Append(vs.Value(i))
		})
		vs.Release()
		return ts, vb.NewArray()
	default:
		panic("unimplemented")
	}
}

func (a *aggregateWindowTransformation) createEmptyWindowsFunc(ts *array.Int, vb array.Builder, mem memory.Allocator, fn func(i int)) *array.Int {
	// We will now use the bounds to iterate over each window.
	// We'll match the windows to the input and append nulls when no match is found.
	bound := a.w.GetLatestBounds(a.bounds.Start)
	for ; bound.Stop() > a.bounds.Start; bound = a.w.PrevBounds(bound) {
		// Do nothing.
	}

	// We found the boundary right before the first window.
	// Move to the first window.
	bound = a.w.NextBounds(bound)

	// Iterate through each window. If the boundary matches the current
	// timestamp, invoke the function. Otherwise, use the builder to append null.
	i, n := 0, ts.Len()
	tb := array.NewIntBuilder(mem)
	for ; bound.Start() < a.bounds.Stop; bound = a.w.NextBounds(bound) {
		b := a.bounds.Intersect(execute.Bounds{
			Start: bound.Start(),
			Stop:  bound.Stop(),
		})

		tv := int64(b.Stop)
		if a.useStart {
			tv = int64(b.Start)
		}

		tb.Append(tv)
		if i < n && tv == ts.Value(i) {
			fn(i)
			i++
			continue
		}
		a.aggregate.AppendEmpty(vb)
	}

	ts.Release()
	return tb.NewIntArray()
}

func (a *aggregateWindowTransformation) Close() error {
	return nil
}

func aggregateWindows(ts, indices, start, stop *array.Int, fn func(i, j int)) {
	l := indices.Len()
	for i, n := 0, start.Len(); i < n; i++ {
		startT := start.Value(i)
		stopT := stop.Value(i)

		startI := 0
		for ; startI < l; startI++ {
			t := ts.Value(int(indices.Value(startI)))
			if t >= startT {
				break
			}
		}

		stopI := startI
		for ; stopI < l; stopI++ {
			t := ts.Value(int(indices.Value(stopI)))
			if t >= stopT {
				break
			}
		}
		fn(startI, stopI)
	}
}

func mergeWindows(prevT, nextT *array.Int, mem memory.Allocator, fn func(i, j int)) *array.Int {
	b := array.NewIntBuilder(mem)
	b.Resize(prevT.Len())

	i, j := 0, 0
	for i < prevT.Len() && j < nextT.Len() {
		l, r := prevT.Value(i), nextT.Value(j)
		if l == r {
			b.Append(l)
			fn(i, j)
			i++
			j++
		} else if l < r {
			b.Append(l)
			fn(i, -1)
			i++
		} else {
			b.Append(r)
			fn(-1, j)
			j++
		}
	}

	if i < prevT.Len() {
		for ; i < prevT.Len(); i++ {
			b.Append(prevT.Value(i))
			fn(i, -1)
		}
	}

	if j < nextT.Len() {
		for ; j < nextT.Len(); j++ {
			b.Append(nextT.Value(j))
			fn(-1, j)
		}
	}
	return b.NewIntArray()
}

type aggregateWindowCount struct{}

func (a aggregateWindowCount) Initialize(valueType flux.ColType, mem memory.Allocator) ([]array.Builder, error) {
	return []array.Builder{array.NewIntBuilder(mem)}, nil
}

func (a aggregateWindowCount) Aggregate(ts, indices, start, stop *array.Int, values array.Interface, builders []array.Builder) {
	b := builders[0].(*array.IntBuilder)
	aggregateWindows(ts, indices, start, stop, func(i, j int) {
		b.Append(int64(j - i))
	})
}

func (a aggregateWindowCount) Merge(prevT, nextT *array.Int, prev, next []array.Interface, mem memory.Allocator) (*array.Int, []array.Interface) {
	first := prev[0].(*array.Int)
	second := next[0].(*array.Int)
	b := array.NewIntBuilder(mem)
	ts := mergeWindows(prevT, nextT, mem, func(i, j int) {
		if i >= 0 && j >= 0 {
			b.Append(first.Value(i) + second.Value(j))
		} else if i >= 0 {
			b.Append(first.Value(i))
		} else {
			b.Append(second.Value(j))
		}
	})
	return ts, []array.Interface{b.NewArray()}
}

func (a aggregateWindowCount) Compute(buffers []array.Interface) (flux.ColType, array.Interface) {
	return flux.TInt, buffers[0]
}

func (a aggregateWindowCount) AppendEmpty(b array.Builder) {
	b.(*array.IntBuilder).Append(0)
}

type aggregateWindowSum struct{}

func (a aggregateWindowSum) Initialize(valueType flux.ColType, mem memory.Allocator) ([]array.Builder, error) {
	var b array.Builder
	switch valueType {
	case flux.TFloat:
		b = array.NewFloatBuilder(mem)
	case flux.TInt:
		b = array.NewIntBuilder(mem)
	case flux.TUInt:
		b = array.NewUintBuilder(mem)
	default:
		return nil, errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", valueType)
	}
	return []array.Builder{b}, nil
}

func (a aggregateWindowSum) Aggregate(ts, indices, start, stop *array.Int, values array.Interface, builders []array.Builder) {
	switch b := builders[0].(type) {
	case *array.FloatBuilder:
		vs := values.(*array.Float)
		aggregateWindows(ts, indices, start, stop, func(i, j int) {
			var sum float64
			for ; i < j; i++ {
				sum += vs.Value(int(indices.Value(i)))
			}
			b.Append(sum)
		})
	case *array.IntBuilder:
		vs := values.(*array.Int)
		aggregateWindows(ts, indices, start, stop, func(i, j int) {
			var sum int64
			for ; i < j; i++ {
				sum += vs.Value(int(indices.Value(i)))
			}
			b.Append(sum)
		})
	case *array.UintBuilder:
		vs := values.(*array.Uint)
		aggregateWindows(ts, indices, start, stop, func(i, j int) {
			var sum uint64
			for ; i < j; i++ {
				sum += vs.Value(int(indices.Value(i)))
			}
			b.Append(sum)
		})
	default:
		panic("unreachable")
	}
}

func (a aggregateWindowSum) Merge(prevT, nextT *array.Int, prev, next []array.Interface, mem memory.Allocator) (*array.Int, []array.Interface) {
	switch prev := prev[0].(type) {
	case *array.Float:
		next := next[0].(*array.Float)
		b := array.NewFloatBuilder(mem)
		ts := mergeWindows(prevT, nextT, mem, func(i, j int) {
			if i >= 0 && j >= 0 {
				b.Append(prev.Value(i) + next.Value(j))
			} else if i >= 0 {
				b.Append(prev.Value(i))
			} else {
				b.Append(next.Value(j))
			}
		})
		return ts, []array.Interface{b.NewArray()}
	case *array.Int:
		next := next[0].(*array.Int)
		b := array.NewIntBuilder(mem)
		ts := mergeWindows(prevT, nextT, mem, func(i, j int) {
			if i >= 0 && j >= 0 {
				b.Append(prev.Value(i) + next.Value(j))
			} else if i >= 0 {
				b.Append(prev.Value(i))
			} else {
				b.Append(next.Value(j))
			}
		})
		return ts, []array.Interface{b.NewArray()}
	case *array.Uint:
		next := next[0].(*array.Uint)
		b := array.NewUintBuilder(mem)
		ts := mergeWindows(prevT, nextT, mem, func(i, j int) {
			if i >= 0 && j >= 0 {
				b.Append(prev.Value(i) + next.Value(j))
			} else if i >= 0 {
				b.Append(prev.Value(i))
			} else {
				b.Append(next.Value(j))
			}
		})
		return ts, []array.Interface{b.NewArray()}
	default:
		panic("unreachable")
	}
}

func (a aggregateWindowSum) Compute(buffers []array.Interface) (flux.ColType, array.Interface) {
	var valueType flux.ColType
	switch buffers[0].(type) {
	case *array.Float:
		valueType = flux.TFloat
	case *array.Int:
		valueType = flux.TInt
	case *array.Uint:
		valueType = flux.TUInt
	default:
		panic("unreachable")
	}
	return valueType, buffers[0]
}

func (a aggregateWindowSum) AppendEmpty(b array.Builder) {
	b.AppendNull()
}

// type aggregateWindowMean struct{}
//
// func (a aggregateWindowMean) Initialize(valueType flux.ColType, mem memory.Allocator) ([]array.Builder, error) {
// 	// TODO implement me
// 	panic("implement me")
// }
//
// func (a aggregateWindowMean) Aggregate(ts, indices, start, stop *array.Int, values array.Interface, builders []array.Builder) {
// 	// TODO implement me
// 	panic("implement me")
// }
//
// func (a aggregateWindowMean) Merge(prevT, nextT *array.Int, prev, next []array.Interface, mem memory.Allocator) (*array.Int, []array.Interface) {
// 	// TODO implement me
// 	panic("implement me")
// }
//
// func (a aggregateWindowMean) Compute(buffers []array.Interface) (flux.ColType, array.Interface) {
// 	// TODO implement me
// 	panic("implement me")
// }

type AggregateWindowRule struct{}

func (a AggregateWindowRule) Name() string {
	return "AggregateWindowRule"
}

func (a AggregateWindowRule) Pattern() plan.Pattern {
	return plan.Pat(WindowKind,
		plan.Pat(SchemaMutationKind,
			plan.OneOf([]plan.ProcedureKind{MeanKind, SumKind, CountKind},
				plan.Pat(WindowKind, plan.Any()))))
}

func (a AggregateWindowRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	if !feature.OptimizeAggregateWindow().Enabled(ctx) {
		return node, false, nil
	}

	windowInfSpec := node.ProcedureSpec().(*WindowProcedureSpec)
	if !a.isValidWindowInfSpec(windowInfSpec) {
		return node, false, nil
	}

	duplicateNode := node.Predecessors()[0]
	duplicateSpec := duplicateNode.ProcedureSpec().(*SchemaMutationProcedureSpec)

	useStart, ok := a.isValidDuplicateSpec(duplicateSpec)
	if !ok {
		return node, false, nil
	}

	aggregateNode := duplicateNode.Predecessors()[0]
	aggregate, valueCol, ok := a.isValidAggregateSpec(aggregateNode.ProcedureSpec())
	if !ok {
		return node, false, nil
	}

	windowNode := aggregateNode.Predecessors()[0]
	windowSpec := windowNode.ProcedureSpec().(*WindowProcedureSpec)
	if !a.isValidWindowSpec(windowSpec) {
		return node, false, nil
	}
	parentNode := windowNode.Predecessors()[0]

	parentNode.ClearSuccessors()
	newNode := plan.CreateUniquePhysicalNode(ctx, "aggregateWindow", &AggregateWindowProcedureSpec{
		spec:      windowSpec,
		aggregate: aggregate,
		valueCol:  valueCol,
		useStart:  useStart,
	})
	parentNode.AddSuccessors(newNode)
	newNode.AddPredecessors(parentNode)
	return newNode, true, nil
}

func (a AggregateWindowRule) isValidWindowInfSpec(spec *WindowProcedureSpec) bool {
	return spec.TimeColumn == execute.DefaultTimeColLabel &&
		spec.StartColumn == execute.DefaultStartColLabel &&
		spec.StopColumn == execute.DefaultStopColLabel &&
		!spec.CreateEmpty ||
		spec.Window.Every == infinityVar.Duration()
}

func (a AggregateWindowRule) isValidDuplicateSpec(spec *SchemaMutationProcedureSpec) (useStart, ok bool) {
	if len(spec.Mutations) != 1 {
		return false, false
	} else if s, ok := spec.Mutations[0].(*DuplicateOpSpec); !ok {
		return false, false
	} else {
		if s.As != execute.DefaultTimeColLabel {
			return false, false
		}

		switch s.Column {
		case execute.DefaultStartColLabel:
			useStart = true
		case execute.DefaultStopColLabel:
			useStart = false
		default:
			return false, false
		}
	}
	return useStart, true
}

func (a AggregateWindowRule) isValidAggregateSpec(spec plan.ProcedureSpec) (aggregateWindow, string, bool) {
	switch spec.Kind() {
	case CountKind:
		aggregateSpec := spec.(*CountProcedureSpec)
		if len(aggregateSpec.Columns) != 1 {
			return nil, "", false
		}
		return aggregateWindowCount{}, aggregateSpec.Columns[0], true
	case SumKind:
		aggregateSpec := spec.(*SumProcedureSpec)
		if len(aggregateSpec.Columns) != 1 {
			return nil, "", false
		}
		return aggregateWindowSum{}, aggregateSpec.Columns[0], true
	// case MeanKind:
	// 	aggregateSpec := spec.(*MeanProcedureSpec)
	// 	if len(aggregateSpec.Columns) != 1 {
	// 		return nil, "", false
	// 	}
	// 	return aggregateWindowMean{}, aggregateSpec.Columns[0], true
	default:
		return nil, "", false
	}
}

func (a AggregateWindowRule) isValidWindowSpec(spec *WindowProcedureSpec) bool {
	return spec.TimeColumn == execute.DefaultTimeColLabel &&
		spec.StartColumn == execute.DefaultStartColLabel &&
		spec.StopColumn == execute.DefaultStopColLabel &&
		spec.Window.Every != infinityVar.Duration()
}

type AggregateWindowCreateEmptyRule struct {
	AggregateWindowRule
}

func (a AggregateWindowCreateEmptyRule) Name() string {
	return "AggregateWindowCreateEmptyRule"
}

func (a AggregateWindowCreateEmptyRule) Pattern() plan.Pattern {
	return plan.Pat(WindowKind,
		plan.Pat(SchemaMutationKind,
			plan.Pat(experimentaltable.FillKind,
				plan.OneOf([]plan.ProcedureKind{MeanKind, SumKind, CountKind},
					plan.Pat(WindowKind, plan.Any())))))
}

func (a AggregateWindowCreateEmptyRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	if !feature.OptimizeAggregateWindow().Enabled(ctx) {
		return node, false, nil
	}

	windowInfSpec := node.ProcedureSpec().(*WindowProcedureSpec)
	if !a.isValidWindowInfSpec(windowInfSpec) {
		return node, false, nil
	}

	duplicateNode := node.Predecessors()[0]
	duplicateSpec := duplicateNode.ProcedureSpec().(*SchemaMutationProcedureSpec)

	useStart, ok := a.isValidDuplicateSpec(duplicateSpec)
	if !ok {
		return node, false, nil
	}

	fillNode := duplicateNode.Predecessors()[0]
	aggregateNode := fillNode.Predecessors()[0]
	aggregate, valueCol, ok := a.isValidAggregateSpec(aggregateNode.ProcedureSpec())
	if !ok {
		return node, false, nil
	}

	windowNode := aggregateNode.Predecessors()[0]
	windowSpec := windowNode.ProcedureSpec().(*WindowProcedureSpec)
	if !a.isValidWindowSpec(windowSpec) {
		return node, false, nil
	}
	parentNode := windowNode.Predecessors()[0]

	parentNode.ClearSuccessors()
	newNode := plan.CreateUniquePhysicalNode(ctx, "aggregateWindow", &AggregateWindowProcedureSpec{
		spec:      windowSpec,
		aggregate: aggregate,
		valueCol:  valueCol,
		useStart:  useStart,
	})
	parentNode.AddSuccessors(newNode)
	newNode.AddPredecessors(parentNode)
	return newNode, true, nil
}
