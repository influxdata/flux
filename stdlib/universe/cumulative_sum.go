package universe

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const CumulativeSumKind = "cumulativeSum"

type CumulativeSumOpSpec struct {
	Columns []string `json:"columns"`
}

func init() {
	cumulativeSumSignature := runtime.MustLookupBuiltinType("universe", "cumulativeSum")

	runtime.RegisterPackageValue("universe", CumulativeSumKind, flux.MustValue(flux.FunctionValue(CumulativeSumKind, createCumulativeSumOpSpec, cumulativeSumSignature)))
	flux.RegisterOpSpec(CumulativeSumKind, newCumulativeSumOp)
	plan.RegisterProcedureSpec(CumulativeSumKind, newCumulativeSumProcedure, CumulativeSumKind)
	execute.RegisterTransformation(CumulativeSumKind, createCumulativeSumTransformation)
}

func createCumulativeSumOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(CumulativeSumOpSpec)
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

func newCumulativeSumOp() flux.OperationSpec {
	return new(CumulativeSumOpSpec)
}

func (s *CumulativeSumOpSpec) Kind() flux.OperationKind {
	return CumulativeSumKind
}

type CumulativeSumProcedureSpec struct {
	plan.DefaultCost
	Columns []string
}

func newCumulativeSumProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*CumulativeSumOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &CumulativeSumProcedureSpec{
		Columns: spec.Columns,
	}, nil
}

func (s *CumulativeSumProcedureSpec) Kind() plan.ProcedureKind {
	return CumulativeSumKind
}
func (s *CumulativeSumProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(CumulativeSumProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *CumulativeSumProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createCumulativeSumTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*CumulativeSumProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewCumulativeSumTransformation(id, s, a.Allocator())
}

type cumulativeSumTransformation struct {
	columns []string
}

type cumulativeSumStateMap map[string]*cumulativeSumState

func NewCumulativeSumTransformation(id execute.DatasetID, spec *CumulativeSumProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &cumulativeSumTransformation{
		columns: spec.Columns,
	}
	return execute.NewNarrowStateTransformation[cumulativeSumStateMap](id, tr, mem)
}

func (c *cumulativeSumTransformation) Process(chunk table.Chunk, state cumulativeSumStateMap, d *execute.TransportDataset, mem memory.Allocator) (cumulativeSumStateMap, bool, error) {
	if state == nil {
		state = make(cumulativeSumStateMap)
	}

	if err := c.processChunk(chunk, state, d, mem); err != nil {
		return nil, false, err
	}
	return state, true, nil
}

func (c *cumulativeSumTransformation) processChunk(chunk table.Chunk, state cumulativeSumStateMap, d *execute.TransportDataset, mem memory.Allocator) error {
	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, chunk.NCols()),
	}

	for j, col := range chunk.Cols() {
		arr := chunk.Values(j)
		if !execute.ContainsStr(c.columns, col.Label) {
			arr.Retain()
			buffer.Values[j] = arr
			continue
		}

		sumer, ok := state[col.Label]
		if !ok {
			sumer = newCumulativeSumState(col.Type)
			if sumer == nil {
				arr.Retain()
				buffer.Values[j] = arr
				continue
			}
			state[col.Label] = sumer
		} else if sumer.inType != col.Type {
			return errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", col.Label, col.Type, sumer.inType)
		}

		buffer.Values[j] = sumer.Sum(arr, mem)
		if execute.ContainsStr(c.columns, col.Label) {
			if _, ok := state[col.Label]; !ok {
				state[col.Label] = &cumulativeSumState{}
			}
		}
	}

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (c *cumulativeSumTransformation) Close() error {
	return nil
}

type cumulativeSum interface {
	Sum(arr array.Array, mem memory.Allocator) array.Array
}

type cumulativeSumState struct {
	inType flux.ColType
	cumulativeSum
}

func newCumulativeSumState(inType flux.ColType) *cumulativeSumState {
	state := &cumulativeSumState{inType: inType}
	switch inType {
	case flux.TFloat:
		state.cumulativeSum = &cumulativeSumFloat{}
	case flux.TInt:
		state.cumulativeSum = &cumulativeSumInt{}
	case flux.TUInt:
		state.cumulativeSum = &cumulativeSumUint{}
	default:
		return nil
	}
	return state
}

type cumulativeSumFloat struct {
	sum float64
}

func (c *cumulativeSumFloat) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewFloatBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Float)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}

type cumulativeSumInt struct {
	sum int64
}

func (c *cumulativeSumInt) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewIntBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Int)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}

type cumulativeSumUint struct {
	sum uint64
}

func (c *cumulativeSumUint) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewUintBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Uint)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}
