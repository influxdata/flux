package universe

import (
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

const SkewKind = "skew"

type SkewOpSpec struct {
	execute.AggregateConfig
}

func init() {
	skewSignature := execute.AggregateSignature(nil, nil)

	flux.RegisterPackageValue("universe", SkewKind, flux.FunctionValue(SkewKind, createSkewOpSpec, skewSignature))
	flux.RegisterOpSpec(SkewKind, newSkewOp)
	plan.RegisterProcedureSpec(SkewKind, newSkewProcedure, SkewKind)
	execute.RegisterTransformation(SkewKind, createSkewTransformation)
}
func createSkewOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	s := new(SkewOpSpec)
	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return nil, err
	}

	return s, nil
}

func newSkewOp() flux.OperationSpec {
	return new(SkewOpSpec)
}

func (s *SkewOpSpec) Kind() flux.OperationKind {
	return SkewKind
}

type SkewProcedureSpec struct {
	execute.AggregateConfig
}

func newSkewProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SkewOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &SkewProcedureSpec{
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *SkewProcedureSpec) Kind() plan.ProcedureKind {
	return SkewKind
}
func (s *SkewProcedureSpec) Copy() plan.ProcedureSpec {
	return &SkewProcedureSpec{
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SkewProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type SkewAgg struct {
	n, m1, m2, m3 float64
}

func createSkewTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SkewProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, new(SkewAgg), s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *SkewAgg) reset() {
	a.n = 0
	a.m1 = 0
	a.m2 = 0
	a.m3 = 0
}
func (a *SkewAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *SkewAgg) NewIntAgg() execute.DoIntAgg {
	a.reset()
	return a
}

func (a *SkewAgg) NewUIntAgg() execute.DoUIntAgg {
	a.reset()
	return a
}

func (a *SkewAgg) NewFloatAgg() execute.DoFloatAgg {
	a.reset()
	return a
}

func (a *SkewAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

func (a *SkewAgg) DoInt(vs *array.Int64) {
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)

		n0 := a.n
		a.n++
		// TODO handle overflow
		delta := float64(v) - a.m1
		deltaN := delta / a.n
		t := delta * deltaN * n0
		a.m3 += t*deltaN*(a.n-2) - 3*deltaN*a.m2
		a.m2 += t
		a.m1 += deltaN
	}
}
func (a *SkewAgg) DoUInt(vs *array.Uint64) {
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)

		n0 := a.n
		a.n++
		// TODO handle overflow
		delta := float64(v) - a.m1
		deltaN := delta / a.n
		t := delta * deltaN * n0
		a.m3 += t*deltaN*(a.n-2) - 3*deltaN*a.m2
		a.m2 += t
		a.m1 += deltaN
	}
}
func (a *SkewAgg) DoFloat(vs *array.Float64) {
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)

		n0 := a.n
		a.n++
		delta := v - a.m1
		deltaN := delta / a.n
		t := delta * deltaN * n0
		a.m3 += t*deltaN*(a.n-2) - 3*deltaN*a.m2
		a.m2 += t
		a.m1 += deltaN
	}
}
func (a *SkewAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *SkewAgg) ValueFloat() float64 {
	if a.n < 2 {
		return math.NaN()
	}
	return math.Sqrt(a.n) * a.m3 / math.Pow(a.m2, 1.5)
}
func (a *SkewAgg) IsNull() bool {
	return a.n == 0
}
