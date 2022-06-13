package universe

import (
	"context"
	"math"
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
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/plan"
	experimentaltable "github.com/influxdata/flux/stdlib/experimental/table"
	"github.com/influxdata/flux/values"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@../../internal/types.tmpldata -o aggregate_window.gen.go aggregate_window.gen.go.tmpl

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
	WindowSpec     *WindowProcedureSpec
	AggregateKind  plan.ProcedureKind
	ValueCol       string
	UseStart       bool
	ForceAggregate bool
}

func (s *AggregateWindowProcedureSpec) Kind() plan.ProcedureKind {
	return AggregateWindowKind
}

func (s *AggregateWindowProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.WindowSpec = ns.WindowSpec.Copy().(*WindowProcedureSpec)
	return &ns
}

type aggregateWindowInitializer func(a *aggregateWindowTransformation, valueType flux.ColType) (aggregateWindow, error)

type aggregateWindowState struct {
	inType flux.ColType
	state  aggregateWindow
}

type aggregateWindow interface {
	// Aggregate will aggregate the values into the buckets denoted by the start/stop
	// arrays. The ts and vs arrays must be the same size while start/stop
	// are the buckets the values will be grouped into.
	Aggregate(ts *array.Int, vs array.Array, start, stop *array.Int, mem memory.Allocator)

	// Merge will take an aggregateWindow of the same type and merge the
	// values from to the into state.
	Merge(from aggregateWindow, mem memory.Allocator)

	// Compute will compute the final values for the aggregated windows.
	Compute(mem memory.Allocator) (*array.Int, flux.ColType, array.Array)

	// Close will release resources associated with this aggregate window state.
	Close() error
}

type aggregateWindowTransformation struct {
	w           interval.Window
	bounds      *execute.Bounds
	createEmpty bool
	timeCol     string
	valueCol    string
	useStart    bool
	initialize  aggregateWindowInitializer
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
	return newAggregateWindowTransformation(id, a.Parents(), s, bounds, a.Allocator())
}

func newAggregateWindowTransformation(id execute.DatasetID, parents []execute.DatasetID, s *AggregateWindowProcedureSpec, bounds *execute.Bounds, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	loc, err := s.WindowSpec.Window.LoadLocation()
	if err != nil {
		return nil, nil, err
	}

	w, err := interval.NewWindowInLocation(
		s.WindowSpec.Window.Every,
		s.WindowSpec.Window.Period,
		s.WindowSpec.Window.Offset,
		loc,
	)
	if err != nil {
		return nil, nil, err
	}

	tr := &aggregateWindowTransformation{
		w:           w,
		bounds:      bounds,
		createEmpty: s.WindowSpec.CreateEmpty,
		timeCol:     s.WindowSpec.TimeColumn,
		valueCol:    s.ValueCol,
		useStart:    s.UseStart,
	}

	switch s.AggregateKind {
	case CountKind:
		tr.initialize = newAggregateWindowCount
	case SumKind:
		tr.initialize = newAggregateWindowSum
	case MeanKind:
		tr.initialize = newAggregateWindowMean
	default:
		return nil, nil, errors.Newf(codes.Internal, "cannot use %q for aggregate window", s.AggregateKind)
	}
	return execute.NewAggregateParallelTransformation(id, parents, tr, mem)
}

func (a *aggregateWindowTransformation) Aggregate(chunk table.Chunk, state interface{}, mem memory.Allocator) (interface{}, bool, error) {
	ws, _ := state.(*aggregateWindowState)
	newState, err := a.processChunk(chunk, ws, mem)
	if err != nil {
		return nil, false, err
	}
	return newState, true, nil
}

func (a *aggregateWindowTransformation) Merge(into, from interface{}, mem memory.Allocator) (interface{}, error) {
	intoState := into.(*aggregateWindowState)
	fromState := from.(*aggregateWindowState)
	if intoState.inType != fromState.inType {
		return nil, errors.Newf(codes.FailedPrecondition, "schema collision detected: column %q is both of type %s and %s", a.valueCol, intoState.inType, fromState.inType)
	}
	intoState.state.Merge(fromState.state, mem)
	return into, nil
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

func (a *aggregateWindowTransformation) processChunk(chunk table.Chunk, ws *aggregateWindowState, mem memory.Allocator) (*aggregateWindowState, error) {
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

	// Verify the input column is still the same.
	if ws != nil {
		if ws.inType != vt {
			return nil, errors.Newf(codes.FailedPrecondition, "schema collision detected: column %q is both of type %s and %s", a.valueCol, ws.inType, vt)
		}
	} else {
		state, err := a.initialize(a, vt)
		if err != nil {
			return nil, err
		}
		ws = &aggregateWindowState{
			inType: vt,
			state:  state,
		}
	}

	// Sort the timestamps and return the
	// offsets of the sorted timestamps.
	ts, vs = a.sort(ts, vs, mem)
	defer ts.Release()
	defer vs.Release()

	// Scan the timestamps and construct the window boundaries.
	start, stop := a.scanWindows(ts, mem)
	defer start.Release()
	defer stop.Release()

	// Send these to the aggregation method.
	ws.state.Aggregate(ts, vs, start, stop, mem)
	return ws, nil
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

func (a *aggregateWindowTransformation) getValueColumn(chunk table.Chunk) (flux.ColType, array.Array, error) {
	idx := chunk.Index(a.valueCol)
	if idx < 0 {
		return flux.TInvalid, nil, errors.Newf(codes.FailedPrecondition, "column %q does not exist", a.valueCol)
	}

	if chunk.Key().HasCol(a.valueCol) {
		return flux.TInvalid, nil, errors.New(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
	}
	return chunk.Col(idx).Type, chunk.Values(idx), nil
}

func (a *aggregateWindowTransformation) isSorted(ts *array.Int) bool {
	arr := ts.Int64Values()
	return sort.SliceIsSorted(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
}

// sort will return the indexes of the array as if it were sorted.
// It does not modify the array and the array returned are the indexes of the
// sorted values.
func (a *aggregateWindowTransformation) sort(ts *array.Int, vs array.Array, mem memory.Allocator) (*array.Int, array.Array) {
	// Check if the timestamps are already sorted.
	if a.isSorted(ts) {
		ts.Retain()
		vs.Retain()
		return ts, vs
	}

	// Construct a mutable array builder so that we can modify the buffer in-place
	// while still using memory accounting.
	indices := mutable.NewInt64Array(mem)
	indices.Resize(ts.Len())

	// Retrieve the raw slice.
	offsets := indices.Int64Values()
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

	// Construct the indices so we can use them.
	arr := indices.NewInt64Array()
	defer arr.Release()

	// Slice of null values from the index.
	if nulls := ts.NullN(); nulls > 0 {
		narr := arrow.IntSlice(arr, 0, ts.Len()-nulls)
		arr.Release()
		arr = narr
	}

	// Copy the arrays using the computed indices.
	ts = arrowutil.CopyIntsByIndex(ts, arr, mem)
	vs = arrowutil.CopyByIndex(vs, arr, mem)
	return ts, vs
}

// scanWindows scans the timestamps and returns the appropriate boundaries.
// Not all timestamps may be associated with a boundary and some timestamps may
// be associated with multiple boundaries.
func (a *aggregateWindowTransformation) scanWindows(ts *array.Int, mem memory.Allocator) (start, stop *array.Int) {
	startB := array.NewIntBuilder(mem)
	stopB := array.NewIntBuilder(mem)
	latest := int64(math.MinInt64)

	// Determine a size hint based on the minimum and maximum times.
	size := 0
	if ts.Len() > 0 {
		startT := ts.Value(0)
		stopT := ts.Value(ts.Len() - 1)
		size = a.sizeHint(startT, stopT)
	}

	// If the size hint was greater than the number of points
	// in this array, then we probably have a sparse array
	// and should just expect one point per window.
	if size > ts.Len() {
		size = ts.Len()
	}

	// Preallocate the array size.
	if size > 0 {
		startB.Resize(size)
		stopB.Resize(size)
	}

	var bounds []execute.Bounds
	for i, n := 0, ts.Len(); i < n; i++ {
		t := ts.Value(i)

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

func (a *aggregateWindowTransformation) computeFromState(key flux.GroupKey, ws *aggregateWindowState, mem memory.Allocator) arrow.TableBuffer {
	ts, vt, vs := ws.state.Compute(mem)
	n := ts.Len()

	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  make([]flux.ColMeta, 0, len(key.Cols())+2),
	}
	buffer.Values = make([]array.Array, 0, cap(buffer.Columns))

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

// sizeHint will return a hint for the number of intervals given the start and stop times.
// Both start and stop are inclusive. This is important as this method is used in multiple
// locations. When used in a context with the start and stop boundaries, we have to adjust
// the exclusive stop time to match the inclusive stop time for this method. When we are
// getting a size hint with actual data, then the time is already inclusive and no adjustment
// is necessary.
//
// The value here is a maximum and not necessarily the value that arrays should be resized to.
// For example, a sparse data set with only 10 points may be spread out over potentially hundreds
// of intervals. For this, it's enough to know that the data is probably sparse and we can just
// allocate for the 10 intervals as determined by the number of points.
func (a *aggregateWindowTransformation) sizeHint(start, stop int64) int {
	every := a.w.Every()
	window := every.Nanoseconds()
	if every.MonthsOnly() {
		// Determine the year/month for the stop and start.
		// Then the math is mostly the same as below.
		startYear, startMonth, _ := values.Time(start).Time().Date()
		start = int64(startYear*12 + int(startMonth) - 1)

		stopYear, stopMonth, _ := values.Time(start).Time().Date()
		stop = int64(stopYear*12 + int(stopMonth) - 1)

		// The window is now 1 because start and stop are now months
		// instead of nanoseconds.
		window = 1
	}
	// We determine the maximum number of intervals by finding
	// the difference between the stop and start times, dividing it
	// by the number of nanoseconds in the interval, then adjusting
	// to include the extra interval.
	//
	// A quick proof of this is a time range of [0, 29] with an interval
	// of 10. This should produce 3. (29-0)/10 = 2 because of integer division
	// and we add 1 to include the extra interval. For [0, 30], we should produce
	// 4 as we have one point in the [30, 40) interval. (30-0)/10 = 3 and adding 1
	// produces the fourth interval.
	return int((stop-start)/window) + 1
}

func (a *aggregateWindowTransformation) Close() error {
	return nil
}

func aggregateWindows(ts, start, stop *array.Int, fn func(i, j int)) {
	l := ts.Len()
	for i, n := 0, start.Len(); i < n; i++ {
		startT := start.Value(i)
		stopT := stop.Value(i)

		startI := 0
		for ; startI < l; startI++ {
			t := ts.Value(startI)
			if t >= startT {
				break
			}
		}

		stopI := startI
		for ; stopI < l; stopI++ {
			t := ts.Value(stopI)
			if t >= stopT {
				break
			}
		}
		fn(startI, stopI)
	}
}

func mergeWindowTimes(prevT, nextT *array.Int, mem memory.Allocator) *array.Int {
	if prevT == nil {
		nextT.Retain()
		return nextT
	}

	b := array.NewIntBuilder(mem)
	b.Resize(prevT.Len())

	i, j := 0, 0
	for i < prevT.Len() && j < nextT.Len() {
		l, r := prevT.Value(i), nextT.Value(j)
		if l == r {
			b.Append(l)
			i++
			j++
		} else if l < r {
			b.Append(l)
			i++
		} else {
			b.Append(r)
			j++
		}
	}

	if i < prevT.Len() {
		for ; i < prevT.Len(); i++ {
			b.Append(prevT.Value(i))
		}
	}

	if j < nextT.Len() {
		for ; j < nextT.Len(); j++ {
			b.Append(nextT.Value(j))
		}
	}
	return b.NewIntArray()
}

func mergeWindowValues(ts, prevT, nextT *array.Int, fn func(i, j int)) {
	prev, next := 0, 0
	for i, n := 0, ts.Len(); i < n; i++ {
		prevM := prev < prevT.Len() && ts.Value(i) == prevT.Value(prev)
		nextM := next < nextT.Len() && ts.Value(i) == nextT.Value(next)
		if prevM && nextM {
			fn(prev, next)
			prev++
			next++
		} else if prevM {
			fn(prev, -1)
			prev++
		} else if nextM {
			fn(-1, next)
			next++
		}
	}
}

type aggregateWindowBase struct {
	a  *aggregateWindowTransformation
	ts *array.Int
}

func (a *aggregateWindowBase) mergeWindows(start, stop *array.Int, mem memory.Allocator, fn func(ts, prev, next *array.Int)) {
	prev := a.ts
	if a.a.useStart {
		stop = start
	}
	a.ts = mergeWindowTimes(prev, stop, mem)
	fn(a.ts, prev, stop)
	if prev != nil {
		prev.Release()
	}
}

func (a *aggregateWindowBase) createEmptyWindows(mem memory.Allocator, fn func(n int) (append func(i int), done func())) {
	if !a.a.createEmpty {
		return
	}

	// We will now use the bounds to iterate over each window.
	// We'll match the windows to the input and append nulls when no match is found.
	bound := a.a.w.GetLatestBounds(a.a.bounds.Start)
	for ; bound.Stop() > a.a.bounds.Start; bound = a.a.w.PrevBounds(bound) {
		// Do nothing.
	}

	// We found the boundary right before the first window.
	// Move to the first window.
	bound = a.a.w.NextBounds(bound)

	// Determine an approximate size for the array.
	// We adjust the stop time because the boundary has an exclusive stop time
	// and sizeHint takes an inclusive stop time.
	size := a.a.sizeHint(int64(a.a.bounds.Start), int64(a.a.bounds.Stop)-1)

	// Iterate through each window and construct the time column.
	tb := array.NewIntBuilder(mem)
	tb.Resize(size)
	for ; bound.Start() < a.a.bounds.Stop; bound = a.a.w.NextBounds(bound) {
		b := a.a.bounds.Intersect(execute.Bounds{
			Start: bound.Start(),
			Stop:  bound.Stop(),
		})

		tv := int64(b.Stop)
		if a.a.useStart {
			tv = int64(b.Start)
		}
		tb.Append(tv)
	}

	// Construct the time column.
	ts := tb.NewIntArray()

	// Use the function to get an append and done function we will use
	// for merging the values.
	append, done := fn(ts.Len())
	i, n := 0, a.ts.Len()

	// Iterate through the timestamps. If there is a match, append the
	// value at that index. If there is no match, pass -1 to signal that
	// the null value for that specific aggregate should be used.
	for _, tv := range ts.Int64Values() {
		if i < n && tv == a.ts.Value(i) {
			append(i)
			i++
		} else {
			append(-1)
		}
	}

	// Mark our iteration as done.
	done()
	a.ts.Release()
	a.ts = ts
}

func (a *aggregateWindowBase) release() {
	if a.ts != nil {
		a.ts.Release()
		a.ts = nil
	}
}

type aggregateWindowCount struct {
	aggregateWindowBase
	vs *array.Int
}

func newAggregateWindowCount(a *aggregateWindowTransformation, valueType flux.ColType) (aggregateWindow, error) {
	return &aggregateWindowCount{
		aggregateWindowBase: aggregateWindowBase{a: a},
	}, nil
}

func (a *aggregateWindowCount) Aggregate(ts *array.Int, vs array.Array, start, stop *array.Int, mem memory.Allocator) {
	b := array.NewIntBuilder(mem)
	b.Resize(stop.Len())
	aggregateWindows(ts, start, stop, func(i, j int) {
		b.Append(int64(j - i))
	})

	result := b.NewIntArray()
	a.merge(start, stop, result, mem)
}

func (a *aggregateWindowCount) merge(start, stop, result *array.Int, mem memory.Allocator) {
	a.mergeWindows(start, stop, mem, func(ts, prev, next *array.Int) {
		if a.vs == nil {
			a.vs = result
			return
		}
		defer result.Release()

		merged := array.NewIntBuilder(mem)
		merged.Resize(ts.Len())
		mergeWindowValues(ts, prev, next, func(i, j int) {
			if i >= 0 && j >= 0 {
				merged.Append(a.vs.Value(i) + result.Value(j))
			} else if i >= 0 {
				merged.Append(a.vs.Value(i))
			} else {
				merged.Append(result.Value(j))
			}
		})
		a.vs.Release()
		a.vs = merged.NewIntArray()
	})
}

func (a *aggregateWindowCount) Merge(from aggregateWindow, mem memory.Allocator) {
	other := from.(*aggregateWindowCount)
	other.vs.Retain()
	a.merge(other.ts, other.ts, other.vs, mem)
}

func (a *aggregateWindowCount) Compute(mem memory.Allocator) (*array.Int, flux.ColType, array.Array) {
	a.createEmptyWindows(mem, func(n int) (append func(i int), done func()) {
		b := array.NewIntBuilder(mem)
		b.Resize(n)

		append = func(i int) {
			if i < 0 {
				b.Append(0)
			} else {
				b.Append(a.vs.Value(i))
			}
		}

		done = func() {
			a.vs.Release()
			a.vs = b.NewIntArray()
		}
		return append, done
	})

	a.ts.Retain()
	a.vs.Retain()
	return a.ts, flux.TInt, a.vs
}

func (a *aggregateWindowCount) Close() error {
	a.release()
	if a.vs != nil {
		a.vs.Release()
		a.vs = nil
	}
	return nil
}

func newAggregateWindowSum(a *aggregateWindowTransformation, valueType flux.ColType) (aggregateWindow, error) {
	base := aggregateWindowBase{a: a}
	switch valueType {
	case flux.TInt:
		return &aggregateWindowSumInt{
			aggregateWindowBase: base,
		}, nil
	case flux.TUInt:
		return &aggregateWindowSumUint{
			aggregateWindowBase: base,
		}, nil
	case flux.TFloat:
		return &aggregateWindowSumFloat{
			aggregateWindowBase: base,
		}, nil
	default:
		return nil, errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", valueType)
	}
}

func newAggregateWindowMean(a *aggregateWindowTransformation, valueType flux.ColType) (aggregateWindow, error) {
	base := aggregateWindowBase{a: a}
	switch valueType {
	case flux.TInt:
		return &aggregateWindowMeanInt{
			aggregateWindowBase: base,
		}, nil
	case flux.TUInt:
		return &aggregateWindowMeanUint{
			aggregateWindowBase: base,
		}, nil
	case flux.TFloat:
		return &aggregateWindowMeanFloat{
			aggregateWindowBase: base,
		}, nil
	default:
		return nil, errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", valueType)
	}
}

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
	valueCol, ok := a.isValidAggregateSpec(aggregateNode.ProcedureSpec())
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
		WindowSpec:     windowSpec,
		AggregateKind:  aggregateNode.Kind(),
		ValueCol:       valueCol,
		UseStart:       useStart,
		ForceAggregate: false,
	})
	parentNode.AddSuccessors(newNode)
	newNode.AddPredecessors(parentNode)
	return newNode, true, nil
}

func (a AggregateWindowRule) isValidWindowInfSpec(spec *WindowProcedureSpec) bool {
	return spec.TimeColumn == execute.DefaultTimeColLabel &&
		spec.StartColumn == execute.DefaultStartColLabel &&
		spec.StopColumn == execute.DefaultStopColLabel &&
		!spec.CreateEmpty &&
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

func (a AggregateWindowRule) isValidAggregateSpec(spec plan.ProcedureSpec) (string, bool) {
	switch spec.Kind() {
	case CountKind:
		aggregateSpec := spec.(*CountProcedureSpec)
		if len(aggregateSpec.Columns) != 1 {
			return "", false
		}
		return aggregateSpec.Columns[0], true
	case SumKind:
		aggregateSpec := spec.(*SumProcedureSpec)
		if len(aggregateSpec.Columns) != 1 {
			return "", false
		}
		return aggregateSpec.Columns[0], true
	case MeanKind:
		aggregateSpec := spec.(*MeanProcedureSpec)
		if len(aggregateSpec.Columns) != 1 {
			return "", false
		}
		return aggregateSpec.Columns[0], true
	default:
		return "", false
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
	valueCol, ok := a.isValidAggregateSpec(aggregateNode.ProcedureSpec())
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
		WindowSpec:     windowSpec,
		AggregateKind:  aggregateNode.Kind(),
		ValueCol:       valueCol,
		UseStart:       useStart,
		ForceAggregate: true,
	})
	parentNode.AddSuccessors(newNode)
	newNode.AddPredecessors(parentNode)
	return newNode, true, nil
}
