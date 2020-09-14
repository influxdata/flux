package aggregate

import (
	"container/list"
	"context"
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const WindowKind = pkgpath + ".window"

type WindowOpSpec struct {
	TimeSrc string
	Window  execute.Window
	Columns []TableColumn
}

func init() {
	runtime.RegisterPackageValue(pkgpath, "window", flux.MustValue(flux.FunctionValue(
		"window",
		createWindowOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "window"),
	)))
	plan.RegisterProcedureSpec(WindowKind, newWindowProcedure, WindowKind)
	execute.RegisterTransformation(WindowKind, createWindowTransformation)
}

func createWindowOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(WindowOpSpec)

	columnsArg, err := args.GetRequired("columns")
	if err != nil {
		return nil, err
	}

	columns, err := tableColumnsFromObject(columnsArg.Object())
	if err != nil {
		return nil, err
	}
	spec.Columns = columns

	// Verify that the output columns don't contain start or stop.
	// This is allowed for the normal table transformation, but
	// not allowed for windowing since we add these columns
	// to the output.
	for _, c := range spec.Columns {
		if c.As == "start" || c.As == "stop" {
			return nil, errors.Newf(codes.Invalid, "window does not allow the output column to be named %q", c.As)
		}
	}

	var every flux.Duration
	if every, err = args.GetRequiredDuration("every"); err != nil {
		return nil, err
	}

	if every.IsNegative() {
		return nil, errors.New(codes.Invalid, "parameter \"every\" must be a positive duration")
	} else if every.IsZero() {
		return nil, errors.New(codes.Invalid, "parameter \"every\" must be a non-zero duration")
	}

	period := every
	if p, ok, err := args.GetDuration("period"); err != nil {
		return nil, err
	} else if ok {
		if p.IsZero() {
			return nil, errors.New(codes.Invalid, "parameter \"period\" must be a positive duration")
		} else if p.IsZero() {
			return nil, errors.New(codes.Invalid, "parameter \"period\" must be a non-zero duration")
		}
		period = p
	}

	offset := flux.Duration{}
	window, err := execute.NewWindow(every, period, offset)
	if err != nil {
		return nil, errors.Newf(codes.Invalid, "invalid window: %s", err)
	}
	spec.Window = window

	if timeSrc, ok, err := args.GetString("time"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeSrc = timeSrc
	}

	return spec, nil
}

func (a *WindowOpSpec) Kind() flux.OperationKind {
	return WindowKind
}

type WindowProcedureSpec struct {
	plan.DefaultCost
	*TableProcedureSpec
	TimeSrc string
	Window  execute.Window
}

func newWindowProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*WindowOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &WindowProcedureSpec{
		TableProcedureSpec: &TableProcedureSpec{
			Columns: spec.Columns,
		},
		TimeSrc: spec.TimeSrc,
		Window:  spec.Window,
	}, nil
}

func (s *WindowProcedureSpec) Kind() plan.ProcedureKind {
	return WindowKind
}

func (s *WindowProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WindowProcedureSpec)
	*ns = *s
	ns.TableProcedureSpec = s.TableProcedureSpec.Copy().(*TableProcedureSpec)
	return ns
}

func createWindowTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WindowProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewWindowTransformation(a.Context(), s, id, a.Allocator())
}

type windowTransformation struct {
	*tableTransformation
	spec *WindowProcedureSpec
}

func NewWindowTransformation(ctx context.Context, spec *WindowProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	w := &windowTransformation{
		tableTransformation: &tableTransformation{
			d:    execute.NewPassthroughDataset(id),
			spec: spec.TableProcedureSpec,
			ctx:  ctx,
			mem:  mem,
		},
		spec: spec,
	}
	return w, w.d, nil
}

func (w *windowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if err := w.validateInputTable(tbl); err != nil {
		return err
	}

	// Retrieve the appropriate time column.
	// This accounts for the search for a default time column,
	// ensures that the column exists, and that it is of type time.
	timeColumn, err := w.getTimeColumn(tbl)
	if err != nil {
		return err
	}

	// Prepare each of the columns.
	columns, err := w.prepare(tbl.Cols(), 0)
	if err != nil {
		return err
	}

	// Construct a buffer for the start and stop columns.
	startB := array.NewInt64Builder(w.mem)
	stopB := array.NewInt64Builder(w.mem)

	// Maintain the state and process the table.
	state := windowTableState{
		windowTransformation: w,
		startB:               startB,
		stopB:                stopB,
		timeColumn:           timeColumn,
		columns:              columns,
		minTime:              int64(math.MinInt64),
	}
	if err := tbl.Do(state.Eval); err != nil {
		return err
	}

	// Flush any remaining windows.
	if err := state.Flush(); err != nil {
		return err
	}

	// Keep a list of active windows.
	// Build the table from the results.
	// TODO(jsternberg): Include start/stop columns here.
	cols := []flux.ColMeta{
		{Label: "start", Type: flux.TTime},
		{Label: "stop", Type: flux.TTime},
	}
	arrs := []array.Interface{startB.NewArray(), stopB.NewArray()}
	outTable, err := w.buildTable(tbl.Key(), cols, arrs, columns)
	if err != nil {
		return err
	}
	return w.d.Process(outTable)
}

func (w *windowTransformation) getTimeColumn(tbl flux.Table) (int, error) {
	var idx int
	if w.spec.TimeSrc != "" {
		if idx = execute.ColIdx(w.spec.TimeSrc, tbl.Cols()); idx < 0 {
			return -1, errors.Newf(codes.FailedPrecondition, "time column %q missing from series %s", w.spec.TimeSrc, tbl.Key())
		}
	} else {
		for i, c := range tbl.Cols() {
			if c.Label == "_time" || c.Label == "time" {
				idx = i
				break
			}
		}

		if idx < 0 {
			return -1, errors.Newf(codes.FailedPrecondition, "default time column \"_time\" or \"time\" missing from series %s", tbl.Key())
		}
	}

	// Verify that the column is the time type.
	if c := tbl.Cols()[idx]; c.Type != flux.TTime {
		return -1, errors.Newf(codes.FailedPrecondition, "time column %q is of type %s when %s is required", c.Label, c.Type, flux.TTime)
	}
	return idx, nil
}

type windowTableState struct {
	*windowTransformation
	startB, stopB *array.Int64Builder
	timeColumn    int
	windows       list.List
	columns       []*columnState
	minTime       int64
}

type windowState struct {
	execute.Bounds
	State []values.Value
}

func (w *windowTableState) Eval(cr flux.ColReader) error {
	// Sanity check for column reader size.
	if cr.Len() == 0 {
		return nil
	}

	// Retrieve the times from the column reader.
	ts := cr.Times(w.timeColumn)
	if ts.NullN() > 0 {
		return errors.New(codes.FailedPrecondition, "null values in the time column are not allowed")
	}

	if err := w.EvalPendingWindows(cr, ts); err != nil {
		return err
	}
	return w.ProcessTable(cr, ts)
}

func (w *windowTableState) Flush() error {
	for e := w.windows.Front(); e != nil; e = e.Next() {
		ws := e.Value.(*windowState)
		if err := w.Write(ws); err != nil {
			return err
		}
	}
	w.windows = list.List{}
	return nil
}

func (w *windowTableState) EvalPendingWindows(cr flux.ColReader, ts *array.Int64) error {
	// If there are any active windows from the last buffer,
	// process them here.
	for e := w.windows.Front(); e != nil; {
		ws := e.Value.(*windowState)

		// If the first time is included in the window interval,
		// then evaluate the window.
		t := ts.Value(0)
		if t < int64(ws.Stop) {
			if complete, err := w.EvalWindow(cr, ts, ws, 0); err != nil {
				return err
			} else if !complete {
				// The window is not complete so move on to the next one.
				e = e.Next()
				continue
			}
		}

		// This interval is complete either because there
		// were no points in the bounds or we evaluated all
		// of them. Write the results.
		if err := w.Write(ws); err != nil {
			return err
		}

		// Remove this window from the active windows.
		next := e.Next()
		w.windows.Remove(e)
		e = next
	}
	return nil
}

func (w *windowTableState) ProcessTable(cr flux.ColReader, ts *array.Int64) error {
	window := w.spec.Window

	// Iterate through each of the time values.
	for i, n := 0, ts.Len(); i < n; i++ {
		// If the time value is earlier than our
		// minimum time, then skip it as the windows
		// associated with this time value were already
		// created.
		t := ts.Value(i)
		if t < w.minTime {
			continue
		}

		// Determine the earliest bounds for the current time.
		bounds := window.GetEarliestBounds(values.Time(t))
		for {
			// The earliest bounds for this time may be in
			// the past so skip over already visited intervals.
			if int64(bounds.Start) >= w.minTime {
				ws := &windowState{
					Bounds: bounds,
					State:  make([]values.Value, len(w.columns)),
				}

				// Evaluate this window with the first time as the
				// offset so we don't process the entire array.
				if complete, err := w.EvalWindow(cr, ts, ws, i); err != nil {
					return err
				} else if complete {
					// The window is complete since there was a time present
					// that was not in the window.
					// We can write the window and then discard since this
					// window hasn't been saved.
					if err := w.Write(ws); err != nil {
						return err
					}
				} else {
					// The window was not completed so save it for later.
					w.windows.PushBack(ws)
				}
			}

			// Move to the next bounds.
			bounds.Start = bounds.Start.Add(window.Every)
			bounds.Stop = bounds.Stop.Add(window.Every)
			if int64(bounds.Start) > t {
				break
			}
		}
		// The next point must be at least in this boundary
		// or its starting bounds would have already been covered
		// by an already produced interval.
		w.minTime = int64(bounds.Start)
	}
	return nil
}

func (w *windowTableState) EvalWindow(cr flux.ColReader, ts *array.Int64, ws *windowState, offset int) (complete bool, err error) {
	// Find the span for the current bounds in this column reader.
	start, stop, n := offset, offset+1, ts.Len()
	for ; stop < n; stop++ {
		// If the time for this column is not
		// in the current boundary, then it is the
		// stop location.
		t := ts.Value(stop)
		if t >= int64(ws.Stop) {
			break
		}
	}

	// Create a slice for the column reader and
	// evaluate the window.
	wr := w.createWindow(cr, start, stop)
	defer wr.Release()

	for i, c := range w.columns {
		if err := c.Eval(w.ctx, &wr, &ws.State[i]); err != nil {
			return false, err
		}
	}
	return stop < n, nil
}

func (w *windowTableState) Write(ws *windowState) error {
	for i, c := range w.columns {
		if err := c.Write(w.ctx, ws.State[i]); err != nil {
			return err
		}
	}

	// Write the start and stop time to the time arrays.
	w.startB.Append(int64(ws.Start))
	w.stopB.Append(int64(ws.Stop))
	return nil
}

func (w *windowTableState) createWindow(cr flux.ColReader, start, stop int) (b arrow.TableBuffer) {
	b.GroupKey = cr.Key()
	b.Columns = cr.Cols()
	b.Values = make([]array.Interface, len(b.Columns))
	for j := range b.Values {
		arr := table.Values(cr, j)
		b.Values[j] = arrow.Slice(arr, int64(start), int64(stop))
	}
	return b
}
