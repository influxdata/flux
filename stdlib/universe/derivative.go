package universe

import (
	"context"
	"time"

	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@../../internal/types.tmpldata -o derivative.gen.go derivative.gen.go.tmpl

const DerivativeKind = "derivative"

type DerivativeOpSpec struct {
	Unit        flux.Duration `json:"unit"`
	NonNegative bool          `json:"nonNegative"`
	Columns     []string      `json:"columns"`
	TimeColumn  string        `json:"timeColumn"`
	InitialZero bool          `json:"initialZero"`
}

func init() {
	derivativeSignature := runtime.MustLookupBuiltinType("universe", "derivative")

	runtime.RegisterPackageValue("universe", DerivativeKind, flux.MustValue(flux.FunctionValue(DerivativeKind, createDerivativeOpSpec, derivativeSignature)))
	flux.RegisterOpSpec(DerivativeKind, newDerivativeOp)
	plan.RegisterProcedureSpec(DerivativeKind, newDerivativeProcedure, DerivativeKind)
	execute.RegisterTransformation(DerivativeKind, createDerivativeTransformation)
}

func createDerivativeOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DerivativeOpSpec)

	if unit, ok, err := args.GetDuration("unit"); err != nil {
		return nil, err
	} else if ok {
		spec.Unit = unit
	} else {
		// Default is 1s
		spec.Unit = flux.ConvertDuration(time.Second)
	}

	if nn, ok, err := args.GetBool("nonNegative"); err != nil {
		return nil, err
	} else if ok {
		spec.NonNegative = nn
	}

	if timeCol, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = timeCol
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if iz, ok, err := args.GetBool("initialZero"); err != nil {
		return nil, err
	} else if ok {
		spec.InitialZero = iz
	}

	if cols, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	} else {
		spec.Columns = []string{execute.DefaultValueColLabel}
	}
	return spec, nil
}

func newDerivativeOp() flux.OperationSpec {
	return new(DerivativeOpSpec)
}

func (s *DerivativeOpSpec) Kind() flux.OperationKind {
	return DerivativeKind
}

type DerivativeProcedureSpec struct {
	plan.DefaultCost
	Unit        flux.Duration `json:"unit"`
	NonNegative bool          `json:"non_negative"`
	Columns     []string      `json:"columns"`
	TimeColumn  string        `json:"timeColumn"`
	InitialZero bool          `json:"initialZero"`
}

func newDerivativeProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DerivativeOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &DerivativeProcedureSpec{
		Unit:        spec.Unit,
		NonNegative: spec.NonNegative,
		Columns:     spec.Columns,
		TimeColumn:  spec.TimeColumn,
		InitialZero: spec.InitialZero,
	}, nil
}

func (s *DerivativeProcedureSpec) Kind() plan.ProcedureKind {
	return DerivativeKind
}
func (s *DerivativeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DerivativeProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *DerivativeProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createDerivativeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DerivativeProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewDerivativeTransformation(a.Context(), id, s, a.Allocator())
}

func NewDerivativeTransformation(ctx context.Context, id execute.DatasetID, spec *DerivativeProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &derivativeTransformation{
		unit:        float64(spec.Unit.Duration()),
		nonNegative: spec.NonNegative,
		columns:     spec.Columns,
		timeCol:     spec.TimeColumn,
		initialZero: spec.InitialZero,
	}
	return execute.NewNarrowStateTransformation[*derivativeState](id, tr, mem)
}

type derivativeTransformation struct {
	unit        float64
	nonNegative bool
	columns     []string
	timeCol     string
	initialZero bool
}

func (t *derivativeTransformation) Process(chunk table.Chunk, state *derivativeState, d *execute.TransportDataset, mem memory.Allocator) (*derivativeState, bool, error) {
	ns, err := t.processChunk(chunk, state, d, mem)
	if err != nil {
		return nil, false, err
	}
	return ns, true, nil
}

func (t *derivativeTransformation) processChunk(chunk table.Chunk, state *derivativeState, d *execute.TransportDataset, mem memory.Allocator) (*derivativeState, error) {
	timeIdx := chunk.Index(t.timeCol)
	if timeIdx < 0 {
		return nil, errors.Newf(codes.FailedPrecondition, "no column %q exists", t.timeCol)
	} else if want, got := flux.TTime, chunk.Col(timeIdx).Type; want != got {
		return nil, errors.Newf(codes.FailedPrecondition, "time column %q is type %s and not %s", t.timeCol, got, want)
	}

	// Initialize or reconcile the state depending on if we have existing state
	// for this group key.
	if state == nil {
		ns, err := t.initializeState(chunk)
		if err != nil {
			return nil, err
		}
		state = ns
	} else {
		if err := t.reconcileState(chunk, state); err != nil {
			return nil, err
		}
	}

	// Count the number of time values that will be considered by
	// the derivative. This is done once here so we can pre-allocate
	// arrays before we start to compute derivatives.
	ts := chunk.Ints(timeIdx)

	// Due to duplicate time values, we need to pre-process the times to determine
	// if they are in a valid order and filter out any duplicates.
	// If they are out of order, this will return an error. If there are duplicates,
	// a mask will be returned that tells us how to filter each column.
	var mask []byte
	bitset, err := t.timeMask(ts, state, mem)
	if err != nil {
		return nil, err
	} else if bitset != nil {
		mask = bitset.Bytes()
		defer bitset.Release()

		// If a mask was returned, use it to filter the time values.
		ts = arrowutil.FilterInts(ts, mask, mem)
		defer ts.Release()
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  state.cols,
		Values:   make([]array.Array, len(state.cols)),
	}
	for i, col := range buffer.Columns {
		// Retrieve the input column for this column.
		var vs array.Array
		if idx := chunk.Index(col.Label); idx >= 0 {
			// Retrieve the input column and apply a mask if required.
			vs = chunk.Values(i)
			if len(mask) > 0 {
				vs = arrowutil.Filter(vs, mask, mem)
			} else {
				vs.Retain()
			}
		} else {
			// If the input column does not exist, produce
			// an array of null values for the given input
			// type as determined by the state.
			//
			// This allows us to continue using the same code
			// below and to process the derivative with an
			// array of null values.
			vs = arrow.Nulls(state.data[i].inputType, ts.Len(), mem)
		}

		// Process the input array with the derivative state.
		colState := state.data[i]
		buffer.Values[i] = colState.state.Do(ts, vs, mem)

		// Release the array. We either retained a copy earlier
		// or used a version that was created by us so we now
		// need to release it.
		vs.Release()
	}

	// Record if at least one row was processed.
	// This is used to ensure that reconciled columns are
	// put into the same initialization state as already existing
	// columns if they are discovered by future table chunks.
	//
	// This condition can end up being false when something like
	// filter returns an empty table chunk because the rows were all
	// filtered and then a future chunk returns at least one value.
	// We want the chunk with at least one value to signify that the
	// derivative was initialized.
	if chunk.Len() > 0 {
		state.initialized = true
	}

	// Validate the buffer was constructed correctly.
	if err := buffer.Validate(); err != nil {
		return nil, err
	}

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, err
	}
	return state, nil
}

// timeMask will produce a mask to exclude duplicate time columns and it will validate
// that the times are strictly ascending.
//
// If the time column is strictly ascending and there are no duplicates, this will
// return nil for the mask which implies that a mask should not be applied.
func (t *derivativeTransformation) timeMask(ts *array.Int, d *derivativeState, mem memory.Allocator) (*memory.Buffer, error) {
	if ts.NullN() > 0 {
		return nil, errors.New(codes.FailedPrecondition, "derivative found null time in time column")
	} else if ts.Len() == 0 {
		return nil, nil
	}

	bitset := memory.NewResizableBuffer(mem)
	bitset.Resize(ts.Len())

	i := 0
	if !d.initialized {
		d.t = ts.Value(0)
		bitutil.SetBit(bitset.Buf(), 0)
		i++
	}

	for ; i < ts.Len(); i++ {
		t := ts.Value(i)
		if t < d.t {
			return nil, errors.New(codes.FailedPrecondition, derivativeUnsortedTimeErr)
		} else if t == d.t {
			// If time did not increase with this row, ignore it.
			bitutil.ClearBit(bitset.Buf(), i)
			continue
		}
		d.t = t
		bitutil.SetBit(bitset.Buf(), i)
	}

	// If the bitset indicates that all rows were selected,
	// do not return a mask.
	n := bitutil.CountSetBits(bitset.Bytes(), 0, bitset.Len())
	if n == ts.Len() {
		bitset.Release()
		return nil, nil
	}
	return bitset, nil
}

// initializeState will initialize the derivativeState using the first table.Chunk for
// the given group key.
func (t *derivativeTransformation) initializeState(chunk table.Chunk) (*derivativeState, error) {
	state := &derivativeState{
		cols: make([]flux.ColMeta, 0, chunk.NCols()),
		data: make([]*derivativeColumn, 0, chunk.NCols()),
	}

	for _, col := range chunk.Cols() {
		if err := t.initializeColumnState(col, state); err != nil {
			return nil, err
		}
	}
	return state, nil
}

// reconcileState will take the existing state and a table.Chunk and it will ensure
// that the types still match and add new columns to the derivativeState if they weren't
// present originally.
func (t *derivativeTransformation) reconcileState(chunk table.Chunk, state *derivativeState) error {
	for _, col := range chunk.Cols() {
		idx := execute.ColIdx(col.Label, state.cols)
		if idx >= 0 {
			// The column previously existed so it needs to have the same
			// input type otherwise it is not valid.
			if want, got := state.data[idx].inputType, col.Type; want != got {
				return errors.Newf(codes.FailedPrecondition, "schema collision detected: column %q is both of type %s and %s", col.Label, want, got)
			}
			continue
		}

		// The column has not previously been seen.
		// Add it and pre-initialize it if the previous columns
		// were already initialized. The pre-initialization is done
		// within the method call.
		if err := t.initializeColumnState(col, state); err != nil {
			return err
		}
	}
	return nil
}

// initializeColumnState will initialize a derivative for the given column and add
// it to the derivativeState.
func (t *derivativeTransformation) initializeColumnState(col flux.ColMeta, state *derivativeState) error {
	data, err := t.derivativeStateFor(col, state)
	if err != nil {
		return err
	}
	state.cols = append(state.cols, flux.ColMeta{
		Label: col.Label,
		Type:  data.Type(),
	})
	state.data = append(state.data, &derivativeColumn{
		inputType: col.Type,
		state:     data,
	})
	return nil
}

// derivativeStateFor will create the derivativeColumnState for the given column.
func (t *derivativeTransformation) derivativeStateFor(col flux.ColMeta, state *derivativeState) (derivativeColumnState, error) {
	if execute.ContainsStr(t.columns, col.Label) {
		switch col.Type {
		case flux.TInt:
			return &derivativeInt{
				unit:        t.unit,
				nonNegative: t.nonNegative,
				initialized: state.initialized,
				initialZero: t.initialZero,
			}, nil
		case flux.TUInt:
			return &derivativeUint{
				unit:        t.unit,
				nonNegative: t.nonNegative,
				initialized: state.initialized,
				initialZero: t.initialZero,
			}, nil
		case flux.TFloat:
			return &derivativeFloat{
				unit:        t.unit,
				nonNegative: t.nonNegative,
				initialized: state.initialized,
				initialZero: t.initialZero,
			}, nil
		default:
			return nil, errors.Newf(codes.FailedPrecondition, "unsupported derivative column type %s:%s", col.Label, col.Type)
		}
	}

	return &derivativePassthrough{
		typ:         col.Type,
		initialized: state.initialized,
	}, nil
}

func (t *derivativeTransformation) Close() error { return nil }

const derivativeUnsortedTimeErr = "derivative found out-of-order times in time column"

type derivativeState struct {
	cols        []flux.ColMeta
	data        []*derivativeColumn
	t           int64
	initialized bool
}

type derivativeColumn struct {
	inputType flux.ColType
	state     derivativeColumnState
}

type derivativeColumnState interface {
	Type() flux.ColType
	Do(ts *array.Int, vs array.Array, mem memory.Allocator) array.Array
}

type derivativePassthrough struct {
	typ         flux.ColType
	initialized bool
}

func (d *derivativePassthrough) Type() flux.ColType {
	return d.typ
}

func (d *derivativePassthrough) Do(ts *array.Int, vs array.Array, mem memory.Allocator) array.Array {
	// Empty column chunk returns an empty array
	// and does not initialize the derivative.
	if vs.Len() == 0 {
		vs.Retain()
		return vs
	}

	// If the derivative has not been initialized, we are going
	// to slice off the first element.
	if !d.initialized {
		d.initialized = true
		return array.Slice(vs, 1, vs.Len())
	} else {
		vs.Retain()
		return vs
	}
}
