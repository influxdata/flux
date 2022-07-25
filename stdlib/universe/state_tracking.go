package universe

import (
	"context"
	"time"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const StateTrackingKind = "stateTracking"

type StateTrackingOpSpec struct {
	Fn             interpreter.ResolvedFunction `json:"fn"`
	CountColumn    string                       `json:"countColumn"`
	DurationColumn string                       `json:"durationColumn"`
	DurationUnit   flux.Duration                `json:"durationUnit"`
	TimeColumn     string                       `json:"timeColumn"`
}

func init() {
	stateTrackingSignature := runtime.MustLookupBuiltinType("universe", "stateTracking")

	runtime.RegisterPackageValue("universe", StateTrackingKind, flux.MustValue(flux.FunctionValue(StateTrackingKind, createStateTrackingOpSpec, stateTrackingSignature)))
	flux.RegisterOpSpec(StateTrackingKind, newStateTrackingOp)
	plan.RegisterProcedureSpec(StateTrackingKind, newStateTrackingProcedure, StateTrackingKind)
	execute.RegisterTransformation(StateTrackingKind, createStateTrackingTransformation)
}

func createStateTrackingOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	fn, err := interpreter.ResolveFunction(f)
	if err != nil {
		return nil, err
	}

	spec := &StateTrackingOpSpec{
		Fn:           fn,
		DurationUnit: flux.ConvertDuration(time.Second),
	}

	if label, ok, err := args.GetString("countColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.CountColumn = label
	}
	if label, ok, err := args.GetString("durationColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.DurationColumn = label
	}
	if unit, ok, err := args.GetDuration("durationUnit"); err != nil {
		return nil, err
	} else if ok {
		spec.DurationUnit = unit
	}
	if label, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = label
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if spec.DurationColumn != "" && !values.Duration(spec.DurationUnit).IsPositive() {
		return nil, errors.New(codes.Invalid, "state tracking duration unit must be greater than zero")
	}
	return spec, nil
}

func newStateTrackingOp() flux.OperationSpec {
	return new(StateTrackingOpSpec)
}

func (s *StateTrackingOpSpec) Kind() flux.OperationKind {
	return StateTrackingKind
}

type StateTrackingProcedureSpec struct {
	plan.DefaultCost
	Fn interpreter.ResolvedFunction
	CountColumn,
	DurationColumn string
	DurationUnit flux.Duration
	TimeCol      string
}

func newStateTrackingProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*StateTrackingOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &StateTrackingProcedureSpec{
		Fn:             spec.Fn,
		CountColumn:    spec.CountColumn,
		DurationColumn: spec.DurationColumn,
		DurationUnit:   spec.DurationUnit,
		TimeCol:        spec.TimeColumn,
	}, nil
}

func (s *StateTrackingProcedureSpec) Kind() plan.ProcedureKind {
	return StateTrackingKind
}
func (s *StateTrackingProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(StateTrackingProcedureSpec)
	*ns = *s

	ns.Fn = s.Fn.Copy()

	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *StateTrackingProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createStateTrackingTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*StateTrackingProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewStateTrackingTransformation(a.Context(), s, id, a.Allocator())
}

func NewStateTrackingTransformation(ctx context.Context, spec *StateTrackingProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	t := &stateTrackingTransformation{
		ctx:      ctx,
		fn:       fn,
		timeCol:  spec.TimeCol,
		countCol: spec.CountColumn,
		durCol:   spec.DurationColumn,
		unit:     int64(spec.DurationUnit.Duration()),
	}
	return execute.NewNarrowStateTransformation(id, t, mem)
}

type stateTrackingTransformation struct {
	ctx context.Context
	fn  *execute.RowPredicateFn

	timeCol,
	countCol,
	durCol string

	unit int64
}

type trackedState struct {
	start,
	prevTime values.Time

	countInState,
	durationInState bool

	count,
	duration int64
}

func (n *stateTrackingTransformation) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	// Track whether or not the state has been modified
	mod := false

	// Initialize state
	if state == nil {
		state = trackedState{
			count:    -1,
			duration: -1,
		}
		mod = true
	}
	s := state.(trackedState)
	mod, err := n.processChunk(chunk, &s, d, mem, mod)
	return s, mod, err
}

func (n *stateTrackingTransformation) Close() error { return nil }

// Updates the state object for each row in the chunk, creates a new chunk with
// columns tracking counts and/or durations, and passes that chunk to the next
// transport node.
func (n *stateTrackingTransformation) processChunk(chunk table.Chunk, state *trackedState, d *execute.TransportDataset, mem memory.Allocator, mod bool) (bool, error) {
	fn, err := n.fn.Prepare(chunk.Cols())
	if err != nil {
		return mod, err
	}

	timeIdx := chunk.Index(n.timeCol)
	if timeIdx < 0 {
		return mod, errors.Newf(codes.FailedPrecondition, "column %q does not exist", n.timeCol)
	}

	buf := chunk.Buffer()
	times := buf.Times(timeIdx)
	if n.durCol != "" {
		if times.NullN() > 0 {
			return mod, errors.New(codes.FailedPrecondition, "got a null timestamp")
		}
	}

	// Create the new columns
	counts := array.NewIntBuilder(mem)
	counts.Resize(chunk.Len())

	durations := array.NewIntBuilder(mem)
	durations.Resize(chunk.Len())

	for i := 0; i < chunk.Len(); i++ {
		// Evaluate the predicate for the current row
		match, err := fn.EvalRow(n.ctx, i, &buf)
		if err != nil {
			return mod, err
		}

		if mod, err = n.updateState(state, times, match, i, mod); err != nil {
			return mod, err
		}

		counts.Append(state.count)
		durations.Append(state.duration)
	}

	return mod, d.Process(n.createChunk(chunk, counts, durations))
}

// Updates the state and returns `true` if the state has been modfied.
func (n *stateTrackingTransformation) updateState(state *trackedState, times *array.Int, match bool, i int, mod bool) (bool, error) {
	if n.durCol != "" {
		ts := values.Time(times.Value(i))
		if state.prevTime > ts {
			return mod, errors.New(codes.FailedPrecondition, "got an out-of-order timestamp")
		}
		state.prevTime = ts
		mod = true

		if match {
			if !state.durationInState {
				state.durationInState = true
				state.start = ts
				state.duration = 0
			} else {
				state.duration = int64(ts - state.start)
				if n.unit > 0 {
					state.duration = state.duration / n.unit
				}
			}
		} else {
			state.durationInState = false
			state.duration = -1
		}
		mod = true
	}

	if n.countCol != "" {
		if match {
			if !state.countInState {
				state.countInState = true
				state.count = 1
			} else {
				state.count++
			}
		} else {
			state.countInState = false
			state.count = -1
		}
		mod = true
	}
	return mod, nil
}

// Returns a copy of an existing chunk with the new columns appended to it.
// `counts` is released if there isn't a column name for it; ditto for `durations`.
func (n *stateTrackingTransformation) createChunk(chunk table.Chunk, counts, durations *array.IntBuilder) table.Chunk {
	ncols := chunk.NCols()
	newCols := append(make([]flux.ColMeta, 0, ncols+2), chunk.Cols()...)

	vs := make([]array.Array, 0, ncols+2)
	for i := 0; i < ncols; i++ {
		col := chunk.Values(i)
		col.Retain()
		vs = append(vs, col)
	}

	if n.countCol != "" {
		newCols = append(newCols, flux.ColMeta{Label: n.countCol, Type: flux.TInt})
		vs = append(vs, counts.NewArray())
	} else {
		counts.Release()
		counts = nil
	}

	if n.durCol != "" {
		newCols = append(newCols, flux.ColMeta{Label: n.durCol, Type: flux.TInt})
		vs = append(vs, durations.NewArray())
	} else {
		durations.Release()
		durations = nil
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  newCols,
		Values:   vs,
	}
	return table.ChunkFromBuffer(buffer)
}
