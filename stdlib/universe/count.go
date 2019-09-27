package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

const CountKind = "count"

type CountOpSpec struct {
	execute.AggregateConfig
}

func init() {
	countSignature := execute.AggregateSignature(nil, nil)
	flux.RegisterPackageValue("universe", CountKind, flux.FunctionValue(CountKind, createCountOpSpec, countSignature))
	flux.RegisterOpSpec(CountKind, newCountOp)
	plan.RegisterProcedureSpec(CountKind, newCountProcedure, CountKind)
	execute.RegisterTransformation(CountKind, createCountTransformation)
}

func createCountOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(CountOpSpec)
	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return nil, err
	}
	return s, nil
}

func newCountOp() flux.OperationSpec {
	return new(CountOpSpec)
}

func (s *CountOpSpec) Kind() flux.OperationKind {
	return CountKind
}

type CountProcedureSpec struct {
	execute.AggregateConfig
}

func newCountProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*CountOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &CountProcedureSpec{
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *CountProcedureSpec) Kind() plan.ProcedureKind {
	return CountKind
}

func (s *CountProcedureSpec) Copy() plan.ProcedureSpec {
	return &CountProcedureSpec{
		AggregateConfig: s.AggregateConfig,
	}
}

func (s *CountProcedureSpec) AggregateMethod() string {
	return CountKind
}
func (s *CountProcedureSpec) ReAggregateSpec() plan.ProcedureSpec {
	return new(SumProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *CountProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type CountAgg struct {
	count int64
}

func createCountTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*CountProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	t, d := execute.NewAggregateTransformationAndDataset(id, mode, new(CountAgg), s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *CountAgg) NewBoolAgg() execute.DoBoolAgg {
	return new(CountAgg)
}
func (a *CountAgg) NewIntAgg() execute.DoIntAgg {
	return new(CountAgg)
}
func (a *CountAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(CountAgg)
}
func (a *CountAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(CountAgg)
}
func (a *CountAgg) NewStringAgg() execute.DoStringAgg {
	return new(CountAgg)
}

func (a *CountAgg) DoBool(vs *array.Boolean) {
	a.count += int64(vs.Len())
}
func (a *CountAgg) DoUInt(vs *array.Uint64) {
	a.count += int64(vs.Len())
}
func (a *CountAgg) DoInt(vs *array.Int64) {
	a.count += int64(vs.Len())
}
func (a *CountAgg) DoFloat(vs *array.Float64) {
	a.count += int64(vs.Len())
}
func (a *CountAgg) DoString(vs *array.Binary) {
	a.count += int64(vs.Len())
}

func (a *CountAgg) Type() flux.ColType {
	return flux.TInt
}
func (a *CountAgg) ValueInt() int64 {
	return a.count
}
func (a *CountAgg) IsNull() bool {
	return false
}
