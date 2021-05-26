package universe

import (
	"context"
	"sort"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/dataset"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
)

type windowTransformation2 struct {
	execute.ExecutionNode
	d           execute.Dataset
	cache       *table.BuilderCache
	bounds      *execute.Bounds
	window      interval.Window
	createEmpty bool
	mem         memory.Allocator

	timeCol, startCol, stopCol string
}

func newWindowTransformation2(id execute.DatasetID, spec *WindowProcedureSpec, bounds *execute.Bounds, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	window, err := interval.NewWindow(spec.Window.Every, spec.Window.Period, spec.Window.Offset)
	if err != nil {
		return nil, nil, err
	}

	if bounds == nil && spec.CreateEmpty {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#nil-bounds-passed-to-window"
		return nil, nil, errors.New(codes.Invalid, "nil bounds passed to window; use range to set the window range").
			WithDocURL(docURL)
	}

	mem := a.Allocator()
	cache := &table.BuilderCache{
		New: func(key flux.GroupKey) table.Builder {
			return table.NewArrowBuilder(key, mem)
		},
		Tables: execute.NewRandomAccessGroupLookup(),
	}
	t := &windowTransformation2{
		d:           dataset.New(id, cache),
		cache:       cache,
		bounds:      bounds,
		window:      window,
		timeCol:     spec.TimeColumn,
		startCol:    spec.StartColumn,
		stopCol:     spec.StopColumn,
		createEmpty: spec.CreateEmpty,
		mem:         mem,
	}
	return t, t.d, nil
}

func (w *windowTransformation2) Process(id execute.DatasetID, tbl flux.Table) error {
	t := w.determineSchemaTemplate(tbl)
	if err := tbl.Do(func(cr flux.ColReader) error {
		return w.processView(cr, &t)
	}); err != nil {
		return err
	}

	if w.createEmpty {
		w.createEmptyWindows(&t)
	}
	return nil
}

func (w *windowTransformation2) processView(cr flux.ColReader, t *windowSchemaTemplate) error {
	// Find the time column for this column reader.
	ts, err := w.getTimeColumn(cr)
	if err != nil {
		return err
	}

	// Sort the timestamps and return the
	// offsets of the sorted timestamps.
	indices := w.sort(ts, w.mem)
	defer indices.Release()

	// Scan the timestamps and construct the window boundaries.
	bounds := w.scanWindows(ts, indices)

	// Create the tables with the values for each window boundary.
	w.createWindows(ts, indices, t, bounds, cr)
	return nil
}

// getTimeColumn retrieves the time column for this flux.ColReader.
func (w *windowTransformation2) getTimeColumn(cr flux.ColReader) (*array.Int64, error) {
	idx := execute.ColIdx(w.timeCol, cr.Cols())
	if idx < 0 {
		return nil, errors.Newf(codes.FailedPrecondition, "no time column: %s", w.timeCol)
	}

	if colType := cr.Cols()[idx].Type; colType != flux.TTime {
		return nil, errors.Newf(codes.FailedPrecondition, "time column is not a time value: %s", colType)
	}
	return cr.Times(idx), nil
}

// sort will return the indexes of the array as if it were sorted.
// It does not modify the array and the array returned are the indexes of the
// sorted values.
func (w *windowTransformation2) sort(ts *array.Int64, mem memory.Allocator) *array.Int64 {
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
func (w *windowTransformation2) scanWindows(ts, indices *array.Int64) []execute.Bounds {
	if w.window.Every() == infinityVar.Duration() {
		bounds := []execute.Bounds{{Start: interval.MinTime, Stop: interval.MaxTime}}
		w.clipBounds(bounds)
		return bounds
	}

	// TODO(jsternberg): optimize further.
	boundsMap := make(map[execute.Bounds]struct{})
	for i, n := 0, indices.Len(); i < n; i++ {
		t := ts.Value(int(indices.Value(i)))

		bound := w.window.GetLatestBounds(values.Time(t))
		for bound.Contains(values.Time(t)) {
			b := execute.Bounds{
				Start: bound.Start(),
				Stop:  bound.Stop(),
			}
			boundsMap[b] = struct{}{}

			// Look at the previous boundary.
			bound = w.window.PrevBounds(bound)
		}
	}

	bounds := make([]execute.Bounds, 0, len(boundsMap))
	for b := range boundsMap {
		bounds = append(bounds, b)
	}
	sort.Slice(bounds, func(i, j int) bool {
		return bounds[i].Start < bounds[j].Start
	})
	w.clipBounds(bounds)
	return bounds
}

func (w *windowTransformation2) clipBounds(bs []execute.Bounds) {
	if w.bounds == nil {
		return
	}

	for i := range bs {
		bs[i] = w.bounds.Intersect(bs[i])
	}
}

func (w *windowTransformation2) determineSchemaTemplate(tbl flux.Table) windowSchemaTemplate {
	// Determine the shared key and column metadata.
	keyCols, keyValues := w.newGroupKeyTemplate(tbl.Key())
	cols := w.createSchema(tbl.Cols())
	return windowSchemaTemplate{
		keyCols:   keyCols,
		keyValues: keyValues,
		cols:      cols,
	}
}

// createWindows iterates over the windows and creates each window
// for the found boundaries.
func (w *windowTransformation2) createWindows(ts, indices *array.Int64, t *windowSchemaTemplate, bounds []execute.Bounds, cr flux.ColReader) {
	// Run through the boundaries and construct the table buffers.
	offset := 0
	for _, bound := range bounds {
		builder := w.getBuilder(t, bound)
		offset = w.appendWindow(ts, indices, bound, builder, offset, cr)
	}
}

// createEmptyWindows will create empty windows for bounds that haven't been created yet.
func (w *windowTransformation2) createEmptyWindows(t *windowSchemaTemplate) {
	if w.window.Every() == infinityVar.Duration() {
		bounds := []execute.Bounds{{Start: interval.MinTime, Stop: interval.MaxTime}}
		w.clipBounds(bounds)
		for _, b := range bounds {
			_ = w.getBuilder(t, b)
		}
		return
	}

	bound := w.window.GetLatestBounds(w.bounds.Start)
	for ; bound.Stop() > w.bounds.Start; bound = w.window.PrevBounds(bound) {
		// Do nothing.
	}

	// We found the boundary right before the first window.
	// Move to the first window.
	bound = w.window.NextBounds(bound)

	// Iterate through each window. Create the group key and then
	// attempt to construct the tables that don't exist yet.
	for ; bound.Start() < w.bounds.Stop; bound = w.window.NextBounds(bound) {
		b := execute.Bounds{
			Start: bound.Start(),
			Stop:  bound.Stop(),
		}
		_ = w.getBuilder(t, b)
	}
}

// newGroupKeyTemplate creates the template for the group key columns and values.
// The columns are consistent across all group keys and the values only
// need to be copied into a new array with the start and stop values set.
func (w *windowTransformation2) newGroupKeyTemplate(key flux.GroupKey) ([]flux.ColMeta, []values.Value) {
	cols := w.createSchema(key.Cols())
	vs := make([]values.Value, len(cols))
	for i, col := range cols {
		if col.Label == w.startCol || col.Label == w.stopCol {
			continue
		}
		vs[i] = key.LabelValue(col.Label)
	}
	return cols, vs
}

// newWindowGroupKey constructs a group key by combining the template
// with the boundary values.
func (w *windowTransformation2) newWindowGroupKey(cols []flux.ColMeta, vs []values.Value, bound execute.Bounds) flux.GroupKey {
	newValues := make([]values.Value, len(vs))
	for i, col := range cols {
		if col.Label == w.startCol {
			start := bound.Start
			if w.bounds != nil && start < w.bounds.Start {
				start = w.bounds.Start
			}
			newValues[i] = values.NewTime(start)
		} else if col.Label == w.stopCol {
			stop := bound.Stop
			if w.bounds != nil && stop > w.bounds.Stop {
				stop = w.bounds.Stop
			}
			newValues[i] = values.NewTime(stop)
		} else {
			newValues[i] = vs[i]
		}
	}
	return execute.NewGroupKey(cols, newValues)
}

// createSchema constructs the table schema for the new tables.
func (w *windowTransformation2) createSchema(cols []flux.ColMeta) []flux.ColMeta {
	ncols := len(cols)
	if execute.ColIdx(w.startCol, cols) < 0 {
		ncols++
	}
	if execute.ColIdx(w.stopCol, cols) < 0 {
		ncols++
	}

	newCols := make([]flux.ColMeta, 0, ncols)
	for _, col := range cols {
		if col.Label == w.startCol || col.Label == w.stopCol {
			col.Type = flux.TTime
		}
		newCols = append(newCols, col)
	}

	if execute.ColIdx(w.startCol, newCols) < 0 {
		newCols = append(newCols, flux.ColMeta{
			Label: w.startCol,
			Type:  flux.TTime,
		})
	}

	if execute.ColIdx(w.stopCol, newCols) < 0 {
		newCols = append(newCols, flux.ColMeta{
			Label: w.stopCol,
			Type:  flux.TTime,
		})
	}
	return newCols
}

// getBuilder returns the builder for the given bounds.
func (w *windowTransformation2) getBuilder(t *windowSchemaTemplate, bound execute.Bounds) *table.ArrowBuilder {
	key := w.newWindowGroupKey(t.keyCols, t.keyValues, bound)
	builder, created := table.GetArrowBuilder(key, w.cache)
	if created {
		// Establish the table schema and initialize the builders.
		builder.Columns = t.cols
		builder.Builders = make([]array.Builder, len(t.cols))
		for i, col := range builder.Columns {
			builder.Builders[i] = arrow.NewBuilder(col.Type, w.mem)
		}
	}
	return builder
}

// appendWinddow will append the values for the current window to the table.
// This takes a start offset to begin the search for the starting point and it
// returns the actual starting point.
func (w *windowTransformation2) appendWindow(ts, indices *array.Int64, bound execute.Bounds, b *table.ArrowBuilder, offset int, cr flux.ColReader) int {
	// Retrieve the span of offsets that are in this boundary.
	start, stop := w.getWindowSpan(ts, indices, bound, offset)

	// Construct a slice with this boundary. We do not worry about releasing
	// the list of indices here because they aren't ours anyway and we only
	// focus on releasing our slice.
	indices = arrow.IntSlice(indices, start, stop)
	defer indices.Release()

	// Copy the values from the column reader.
	for j, col := range b.Columns {
		builder := b.Builders[j]

		switch col.Label {
		case w.startCol:
			b := builder.(*array.Int64Builder)
			b.Reserve(indices.Len())
			for i, n := 0, indices.Len(); i < n; i++ {
				b.Append(int64(bound.Start))
			}
		case w.stopCol:
			b := builder.(*array.Int64Builder)
			b.Reserve(indices.Len())
			for i, n := 0, indices.Len(); i < n; i++ {
				b.Append(int64(bound.Stop))
			}
		default:
			idx := execute.ColIdx(col.Label, cr.Cols())
			arr := table.Values(cr, idx)
			arrowutil.CopyByIndexTo(builder, arr, indices)
		}
	}
	return start
}

// getWindowSpan retrieves the span of indexes that fit into this boundary.
// There will always be at least one point because we only invoke this method
// with boundaries that we have already determined to have at least one row.
func (w *windowTransformation2) getWindowSpan(ts, indexes *array.Int64, bound execute.Bounds, offset int) (start, stop int) {
	// Find the start offset that fits in this boundary.
	n := indexes.Len()
	for ; offset < n; offset++ {
		t := ts.Value(int(indexes.Value(offset)))
		if values.Time(t) >= bound.Start {
			break
		}
	}

	// Determine the stop offset.
	stop = offset + 1
	for ; stop < n; stop++ {
		t := ts.Value(int(indexes.Value(stop)))
		if !bound.Contains(values.Time(t)) {
			break
		}
	}
	return offset, stop
}

func (w *windowTransformation2) Finish(id execute.DatasetID, err error) {
	w.d.Finish(err)
}

func (w *windowTransformation2) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return w.d.RetractTable(key)
}
func (w *windowTransformation2) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return w.d.UpdateWatermark(t)
}
func (w *windowTransformation2) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return w.d.UpdateProcessingTime(t)
}

type windowSchemaTemplate struct {
	keyCols   []flux.ColMeta
	keyValues []values.Value
	cols      []flux.ColMeta
}

type OptimizeWindowRule struct{}

func (r OptimizeWindowRule) Name() string {
	return "OptimizeWindowRule"
}

func (r OptimizeWindowRule) Pattern() plan.Pattern {
	return plan.Pat(WindowKind, plan.Any())
}

func (r OptimizeWindowRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	windowSpec := node.ProcedureSpec().(*WindowProcedureSpec)
	if windowSpec.Optimize {
		return node, false, nil
	}
	windowSpec = windowSpec.Copy().(*WindowProcedureSpec)
	windowSpec.Optimize = true
	if err := node.ReplaceSpec(windowSpec); err != nil {
		return node, false, err
	}
	return node, true, nil
}
