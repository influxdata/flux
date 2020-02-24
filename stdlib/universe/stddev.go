package universe

import (
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const (
	StddevKind = "stddev"

	modePopulation = "population"
	modeSample     = "sample"
)

type StddevOpSpec struct {
	Mode string `json:"mode"`
	execute.AggregateConfig
}

func init() {
	stddevSignature := semantic.MustLookupBuiltinType("universe", "stddev")

	runtime.RegisterPackageValue("universe", StddevKind, flux.MustValue(flux.FunctionValue(StddevKind, createStddevOpSpec, stddevSignature)))
	flux.RegisterOpSpec(StddevKind, newStddevOp)
	plan.RegisterProcedureSpec(StddevKind, newStddevProcedure, StddevKind)
	execute.RegisterTransformation(StddevKind, createStddevTransformation)
}
func createStddevOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	s := new(StddevOpSpec)

	if mode, ok, err := args.GetString("mode"); err != nil {
		return nil, err
	} else if ok {
		if mode != modePopulation && mode != modeSample {
			return nil, errors.Newf(codes.Invalid, "%q is not a valid standard deviation mode", mode)
		}
		s.Mode = mode
	} else {
		s.Mode = modeSample
	}

	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return s, err
	}
	return s, nil
}

func newStddevOp() flux.OperationSpec {
	return new(StddevOpSpec)
}

func (s *StddevOpSpec) Kind() flux.OperationKind {
	return StddevKind
}

type StddevProcedureSpec struct {
	Mode string `json:"mode"`
	execute.AggregateConfig
}

func newStddevProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*StddevOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &StddevProcedureSpec{
		Mode:            spec.Mode,
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *StddevProcedureSpec) Kind() plan.ProcedureKind {
	return StddevKind
}
func (s *StddevProcedureSpec) Copy() plan.ProcedureSpec {
	return &StddevProcedureSpec{
		Mode:            s.Mode,
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *StddevProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type StddevAgg struct {
	Mode        string
	n, m2, mean float64
}

func createStddevTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*StddevProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, &StddevAgg{Mode: s.Mode}, s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *StddevAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *StddevAgg) NewIntAgg() execute.DoIntAgg {
	return &StddevAgg{Mode: a.Mode}
}

func (a *StddevAgg) NewUIntAgg() execute.DoUIntAgg {
	return &StddevAgg{Mode: a.Mode}
}

func (a *StddevAgg) NewFloatAgg() execute.DoFloatAgg {
	return &StddevAgg{Mode: a.Mode}
}

func (a *StddevAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}
func (a *StddevAgg) DoInt(vs *array.Int64) {
	var delta, delta2 float64
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)
		a.n++
		// TODO handle overflow
		delta = float64(v) - a.mean
		a.mean += delta / a.n
		delta2 = float64(v) - a.mean
		a.m2 += delta * delta2
	}
}
func (a *StddevAgg) DoUInt(vs *array.Uint64) {
	var delta, delta2 float64
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)
		a.n++
		// TODO handle overflow
		delta = float64(v) - a.mean
		a.mean += delta / a.n
		delta2 = float64(v) - a.mean
		a.m2 += delta * delta2
	}
}
func (a *StddevAgg) DoFloat(vs *array.Float64) {
	var delta, delta2 float64
	for i := 0; i < vs.Len(); i++ {
		if vs.IsNull(i) {
			continue
		}
		v := vs.Value(i)
		a.n++
		delta = v - a.mean
		a.mean += delta / a.n
		delta2 = v - a.mean
		a.m2 += delta * delta2
	}
}
func (a *StddevAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *StddevAgg) ValueFloat() float64 {
	var n = a.n
	if a.Mode == modeSample {
		n--
	}
	if n < 1 {
		return math.NaN()
	}
	return math.Sqrt(a.m2 / float64(n))
}
func (a *StddevAgg) IsNull() bool {
	return a.n == 0
}
