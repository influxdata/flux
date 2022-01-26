package universe

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
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
	t, d := NewLimitTransformation(s, id, a.Allocator())
	return t, d, nil
}

type limitState struct {
	n, offset int
}

type limitTransformation struct {
	n, offset int
}

func NewLimitTransformation(spec *LimitProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset) {
	t := &limitTransformation{
		n:      int(spec.N),
		offset: int(spec.Offset),
	}
	tr, d, _ := execute.NewNarrowStateTransformation(id, t, mem)
	return tr, d
}

func (t *limitTransformation) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	ls := limitState{n: t.n, offset: t.offset}
	if state != nil {
		ls = state.(limitState)
	}

	if ls.n <= 0 {
		return ls, true, nil
	}

	l := chunk.Len()
	if l <= ls.offset {
		// Skip entire batch
		ls.offset -= l
		return ls, true, nil
	}

	start := ls.offset
	stop := l
	count := stop - start
	if count > ls.n {
		count = ls.n
		stop = start + count
	}

	// Reduce the number of rows we will keep from the
	// next buffer and set the offset to zero as it has been
	// entirely consumed.
	ls.n -= count
	ls.offset = 0

	buffer := chunk.Buffer()
	buffer.Values = make([]array.Interface, chunk.NCols())
	for j := range buffer.Values {
		arr := chunk.Values(j)
		if arr.Len() == count {
			arr.Retain()
		} else {
			arr = arrow.Slice(arr, int64(start), int64(stop))
		}
		buffer.Values[j] = arr
	}

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, false, err
	}
	return ls, true, nil
}

func (t *limitTransformation) Close() error {
	return nil
}
