package universe

import (
	"fmt"

	"sort"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/math"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const MadKind = "mad"

type MadOpSpec struct {
	execute.AggregateConfig
}

func init() {
	madSignature := execute.AggregateSignature(nil, nil)

	flux.RegisterPackageValue("universe", MadKind, flux.FunctionValue(MadKind, createMadOpSpec, madSignature))
	flux.RegisterOpSpec(MadKind, newMadOp)
	plan.RegisterProcedureSpec(MadKind, newMadProcedure, MadKind)
	execute.RegisterTransformation(MadKind, createMadTransformation)
}

func createMadOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(MadOpSpec)
	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return s, err
	}
	return s, nil
}

func newMadOp() flux.OperationSpec {
	return new(MadOpSpec)
}

func (s *MadOpSpec) Kind() flux.OperationKind {
	return MadKind
}

type MadProcedureSpec struct {
	execute.AggregateConfig
}

func newMadProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MadOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &MadProcedureSpec{
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *MadProcedureSpec) Kind() plan.ProcedureKind {
	return MadKind
}

func (s *MadProcedureSpec) Copy() plan.ProcedureSpec {
	return &MadProcedureSpec{
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *MadProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func (s *MadProcedureSpec) AggregateMethod() string {
	return MadKind
}
func (s *MadProcedureSpec) ReAggregateSpec() plan.ProcedureSpec {
	return new(MadProcedureSpec)
}

type MadAgg struct{}

func createMadTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MadProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	t, d := execute.NewAggregateTransformationAndDataset(id, mode, new(MadAgg), s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *MadAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}
func (a *MadAgg) NewIntAgg() execute.DoIntAgg {
	return new(MadIntAgg)
}
func (a *MadAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(MadUIntAgg)
}
func (a *MadAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(MadFloatAgg)
}
func (a *MadAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

type MadIntAgg struct {
	mad int64
	ok  bool
}

func (a *MadIntAgg) DoInt(vs *array.Int64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.mad += math.Int64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.mad += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *MadIntAgg) Type() flux.ColType {
	return flux.TInt
}
func (a *MadIntAgg) ValueInt() int64 {
	return a.mad
}
func (a *MadIntAgg) IsNull() bool {
	return !a.ok
}

type MadUIntAgg struct {
	mad uint64
	ok  bool
}

func (a *MadUIntAgg) DoUInt(vs *array.Uint64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.mad += math.Uint64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.mad += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *MadUIntAgg) Type() flux.ColType {
	return flux.TUInt
}
func (a *MadUIntAgg) ValueUInt() uint64 {
	return a.mad
}
func (a *MadUIntAgg) IsNull() bool {
	return !a.ok
}

type MadFloatAgg struct {
	mad float64
	ok  bool
}

func (a *MadFloatAgg) DoFloat(vs *array.Float64) {
	// test if we can order the dataset CHECK
	// test if we can find the median CHECK
	// test if we can find the differences between median and each value
	// test if we can find the sort that difference
	// test if we can find the median value of that
	sort.Slice(vs.Float64Values(), func(i, j int) bool {
		return vs.Float64Values()[i] < vs.Float64Values()[j]
	})
	fmt.Printf("sorted %v", vs.Float64Values())
	lDataset := len(vs.Float64Values()) // find length of Dataset
	mNumber := lDataset / 2             // find the median
	fmt.Printf("length %v\nmiddle %v\n", lDataset, mNumber)
	var median float64

	if lDataset%2 == 0 {
		median = (vs.Float64Values()[mNumber-1] + vs.Float64Values()[mNumber]) / 2
	} else {
		median = vs.Float64Values()[mNumber]
	}
	fmt.Printf("median %v\n\n", median)

	var diff []float64
	for _, j := range vs.Float64Values() {
		diff = append(diff, j-median)
	}
	sort.Slice(diff, func(i, j int) bool {
		return diff[i] < diff[j]
	})
	lDataset = len(diff) // find length of Dataset
	mNumber = lDataset / 2
	if lDataset%2 == 0 {
		a.mad = (diff[mNumber-1] + diff[mNumber]) / 2
	} else {
		a.mad = diff[mNumber]
	}
	a.ok = true
}
func (a *MadFloatAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *MadFloatAgg) ValueFloat() float64 {
	return a.mad
}
func (a *MadFloatAgg) IsNull() bool {
	return !a.ok
}
