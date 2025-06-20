package universe

import (
	"math"

	arrowmath "github.com/apache/arrow-go/v18/arrow/math"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const MeanKind = "mean"

type MeanOpSpec struct {
	execute.SimpleAggregateConfig
}

func init() {
	meanSignature := runtime.MustLookupBuiltinType("universe", "mean")

	runtime.RegisterPackageValue("universe", MeanKind, flux.MustValue(flux.FunctionValue(MeanKind, CreateMeanOpSpec, meanSignature)))
	plan.RegisterProcedureSpec(MeanKind, newMeanProcedure, MeanKind)
	execute.RegisterTransformation(MeanKind, createMeanTransformation)
}
func CreateMeanOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := &MeanOpSpec{}
	if err := spec.SimpleAggregateConfig.ReadArgs(args); err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *MeanOpSpec) Kind() flux.OperationKind {
	return MeanKind
}

type MeanProcedureSpec struct {
	execute.SimpleAggregateConfig
}

func newMeanProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MeanOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &MeanProcedureSpec{
		SimpleAggregateConfig: spec.SimpleAggregateConfig,
	}, nil
}

func (s *MeanProcedureSpec) Kind() plan.ProcedureKind {
	return MeanKind
}
func (s *MeanProcedureSpec) Copy() plan.ProcedureSpec {
	return &MeanProcedureSpec{
		SimpleAggregateConfig: s.SimpleAggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *MeanProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type MeanAgg struct {
	count int64
	sum   float64
}

func createMeanTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MeanProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return execute.NewSimpleAggregateTransformation(a.Context(), id, new(MeanAgg), s.SimpleAggregateConfig, a.Allocator())
}

func (a *MeanAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *MeanAgg) NewIntAgg() execute.DoIntAgg {
	return new(MeanAgg)
}

func (a *MeanAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(MeanAgg)
}

func (a *MeanAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(MeanAgg)
}

func (a *MeanAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

func (a *MeanAgg) DoInt(vs *array.Int) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		a.count += int64(l)
		if vs.NullN() == 0 {
			a.sum += float64(arrowmath.Int64.Sum(vs))
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += float64(vs.Value(i))
				}
			}
		}
	}
}
func (a *MeanAgg) DoUInt(vs *array.Uint) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		a.count += int64(l)
		if vs.NullN() == 0 {
			a.sum += float64(arrowmath.Uint64.Sum(vs))
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += float64(vs.Value(i))
				}
			}
		}
	}
}
func (a *MeanAgg) DoFloat(vs *array.Float) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		a.count += int64(l)
		if vs.NullN() == 0 {
			a.sum += arrowmath.Float64.Sum(vs)
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.sum += float64(vs.Value(i))
				}
			}
		}
	}
}
func (a *MeanAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *MeanAgg) ValueFloat() float64 {
	if a.count < 1 {
		return math.NaN()
	}
	return a.sum / float64(a.count)
}
func (a *MeanAgg) IsNull() bool {
	return a.count == 0
}
