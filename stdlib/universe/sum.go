package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/math"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const SumKind = "sum"

type SumOpSpec struct {
	execute.AggregateConfig
}

func init() {
	sumSignature := semantic.MustLookupBuiltinType("universe", "sum")

	runtime.RegisterPackageValue("universe", SumKind, flux.MustValue(flux.FunctionValue(SumKind, createSumOpSpec, sumSignature)))
	flux.RegisterOpSpec(SumKind, newSumOp)
	plan.RegisterProcedureSpec(SumKind, newSumProcedure, SumKind)
	execute.RegisterTransformation(SumKind, createSumTransformation)
}

func createSumOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(SumOpSpec)
	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return s, err
	}
	return s, nil
}

func newSumOp() flux.OperationSpec {
	return new(SumOpSpec)
}

func (s *SumOpSpec) Kind() flux.OperationKind {
	return SumKind
}

type SumProcedureSpec struct {
	execute.AggregateConfig
}

func newSumProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SumOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &SumProcedureSpec{
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *SumProcedureSpec) Kind() plan.ProcedureKind {
	return SumKind
}

func (s *SumProcedureSpec) Copy() plan.ProcedureSpec {
	return &SumProcedureSpec{
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SumProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func (s *SumProcedureSpec) AggregateMethod() string {
	return SumKind
}
func (s *SumProcedureSpec) ReAggregateSpec() plan.ProcedureSpec {
	return new(SumProcedureSpec)
}

type SumAgg struct{}

func createSumTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SumProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	t, d := execute.NewAggregateTransformationAndDataset(id, mode, new(SumAgg), s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *SumAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}
func (a *SumAgg) NewIntAgg() execute.DoIntAgg {
	return new(SumIntAgg)
}
func (a *SumAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(SumUIntAgg)
}
func (a *SumAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(SumFloatAgg)
}
func (a *SumAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

type SumIntAgg struct {
	sum int64
	ok  bool
}

func (a *SumIntAgg) DoInt(vs *array.Int64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.sum += math.Int64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *SumIntAgg) Type() flux.ColType {
	return flux.TInt
}
func (a *SumIntAgg) ValueInt() int64 {
	return a.sum
}
func (a *SumIntAgg) IsNull() bool {
	return !a.ok
}

type SumUIntAgg struct {
	sum uint64
	ok  bool
}

func (a *SumUIntAgg) DoUInt(vs *array.Uint64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.sum += math.Uint64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *SumUIntAgg) Type() flux.ColType {
	return flux.TUInt
}
func (a *SumUIntAgg) ValueUInt() uint64 {
	return a.sum
}
func (a *SumUIntAgg) IsNull() bool {
	return !a.ok
}

type SumFloatAgg struct {
	sum float64
	ok  bool
}

func (a *SumFloatAgg) DoFloat(vs *array.Float64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.sum += math.Float64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *SumFloatAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *SumFloatAgg) ValueFloat() float64 {
	return a.sum
}
func (a *SumFloatAgg) IsNull() bool {
	return !a.ok
}
