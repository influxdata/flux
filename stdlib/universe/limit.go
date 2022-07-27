package universe

import (
	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const LimitKind = "limit"

// LimitOpSpec limits the number of rows returned per table.
type LimitOpSpec struct {
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func init() {
	limitSignature := runtime.MustLookupBuiltinType("universe", "limit")

	runtime.RegisterPackageValue("universe", LimitKind, flux.MustValue(flux.FunctionValue(LimitKind, createLimitOpSpec, limitSignature)))
	flux.RegisterOpSpec(LimitKind, newLimitOp)
	plan.RegisterProcedureSpec(LimitKind, newLimitProcedure, LimitKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	execute.RegisterTransformation(LimitKind, createLimitTransformation)
}

func createLimitOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LimitOpSpec)

	n, err := args.GetRequiredInt("n")
	if err != nil {
		return nil, err
	}
	spec.N = n

	if offset, ok, err := args.GetInt("offset"); err != nil {
		return nil, err
	} else if ok {
		spec.Offset = offset
	}

	return spec, nil
}

func newLimitOp() flux.OperationSpec {
	return new(LimitOpSpec)
}

func (s *LimitOpSpec) Kind() flux.OperationKind {
	return LimitKind
}

type LimitProcedureSpec struct {
	plan.DefaultCost
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func newLimitProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LimitOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &LimitProcedureSpec{
		N:      spec.N,
		Offset: spec.Offset,
	}, nil
}

func (s *LimitProcedureSpec) Kind() plan.ProcedureKind {
	return LimitKind
}
func (s *LimitProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LimitProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LimitProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createLimitTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LimitProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewLimitTransformation(s, id, a.Allocator())
}

type limitState struct {
	n      int
	offset int
}

func NewLimitTransformation(
	spec *LimitProcedureSpec,
	id execute.DatasetID,
	mem memory.Allocator,
) (execute.Transformation, execute.Dataset, error) {
	t := &limitTransformation{
		n:      int(spec.N),
		offset: int(spec.Offset),
	}
	return execute.NewNarrowStateTransformation(id, t, mem)
}

type limitTransformation struct {
	n, offset int
}

func (t *limitTransformation) Process(
	chunk table.Chunk,
	state interface{},
	dataset *execute.TransportDataset,
	_ arrowmem.Allocator,
) (interface{}, bool, error) {

	var state_ *limitState
	// `.Process` is reentrant, so to speak. The first invocation will not
	// include a value for `state`. Initialization happens here then is passed
	// in/out for the subsequent calls.
	if state == nil {
		state_ = &limitState{n: t.n, offset: t.offset}
	} else {
		state_ = state.(*limitState)
	}
	return t.processChunk(chunk, state_, dataset)
}

func (t *limitTransformation) processChunk(
	chunk table.Chunk,
	state *limitState,
	dataset *execute.TransportDataset,
) (*limitState, bool, error) {

	chunkLen := chunk.Len()

	// Pass empty chunks along to downstream transformations for these cases.
	if state.n <= 0 || chunkLen == 0 {
		// TODO(onelson): seems like there should be a more simple way to produce an empty chunk
		buf := chunk.Buffer()
		buf.Values = make([]array.Array, chunk.NCols())
		for idx := range buf.Values {
			values := chunk.Values(idx)
			if values.Len() == 0 {
				values.Retain()
			} else {
				values = arrow.Slice(values, int64(0), int64(0))
			}
			buf.Values[idx] = values
		}
		out := table.ChunkFromBuffer(buf)
		if err := dataset.Process(out); err != nil {
			return nil, false, err
		}
		return state, true, nil
	}

	if chunkLen <= state.offset {
		state.offset -= chunkLen
		return state, true, nil
	}

	start := state.offset
	stop := chunkLen
	count := stop - start
	if count > state.n {
		count = state.n
		stop = start + count
	}

	// Update state for the next iteration
	state.n -= count
	state.offset = 0

	buf := chunk.Buffer()
	// XXX(onelson): seems like we're building a 2D array where the outer is by
	// column, and the inners are the column values per row?
	buf.Values = make([]array.Array, chunk.NCols())
	for idx := range buf.Values {
		values := chunk.Values(idx)
		// If there's no cruft at the end, just keep the original array,
		// otherwise we need to truncate it to ensure all inners have the
		// expected size.
		// XXX(onelson): Could there be a 3rd case where we have less than the count?
		if values.Len() == count {
			values.Retain()
		} else {
			values = arrow.Slice(values, int64(start), int64(stop))
		}
		buf.Values[idx] = values
	}
	out := table.ChunkFromBuffer(buf)
	if err := dataset.Process(out); err != nil {
		return nil, false, err
	}
	return state, true, nil
}

func (*limitTransformation) Close() error {
	return nil
}
