package universe

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/math"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const LinregKind = "linreg"

type LinregOpSpec struct {
	execute.AggregateConfig
}

func init() {
	linregSignature := execute.AggregateSignature(nil, nil)

	flux.RegisterPackageValue("universe", LinregKind, flux.FunctionValue(LinregKind, createLinregOpSpec, linregSignature))
	flux.RegisterOpSpec(LinregKind, newLinregOp)
	plan.RegisterProcedureSpec(LinregKind, newLinregProcedure, LinregKind)
	execute.RegisterTransformation(LinregKind, createLinregTransformation)
}

func createLinregOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(LinregOpSpec)
	if err := s.AggregateConfig.ReadArgs(args); err != nil {
		return s, err
	}
	return s, nil
}

func newLinregOp() flux.OperationSpec {
	return new(LinregOpSpec)
}

func (s *LinregOpSpec) Kind() flux.OperationKind {
	return LinregKind
}

type LinregProcedureSpec struct {
	execute.AggregateConfig
}

func newLinregProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LinregOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &LinregProcedureSpec{
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *LinregProcedureSpec) Kind() plan.ProcedureKind {
	return LinregKind
}

func (s *LinregProcedureSpec) Copy() plan.ProcedureSpec {
	return &LinregProcedureSpec{
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LinregProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func (s *LinregProcedureSpec) AggregateMethod() string {
	return LinregKind
}
func (s *LinregProcedureSpec) ReAggregateSpec() plan.ProcedureSpec {
	return new(LinregProcedureSpec)
}

type LinregAgg struct{}

func createLinregTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LinregProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	t, d := execute.NewAggregateTransformationAndDataset(id, mode, new(LinregAgg), s.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *LinregAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}
func (a *LinregAgg) NewIntAgg() execute.DoIntAgg {
	return new(LinregIntAgg)
}
func (a *LinregAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(LinregUIntAgg)
}
func (a *LinregAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(LinregFloatAgg)
}
func (a *LinregAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

type LinregIntAgg struct {
	linreg int64
	ok  bool
}

func (a *LinregIntAgg) DoInt(vs *array.Int64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.linreg += math.Int64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.linreg += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *LinregIntAgg) Type() flux.ColType {
	return flux.TInt
}
func (a *LinregIntAgg) ValueInt() int64 {
	return a.linreg
}
func (a *LinregIntAgg) IsNull() bool {
	return !a.ok
}

type LinregUIntAgg struct {
	linreg uint64
	ok  bool
}

func (a *LinregUIntAgg) DoUInt(vs *array.Uint64) {
	if l := vs.Len() - vs.NullN(); l > 0 {
		if vs.NullN() == 0 {
			a.linreg += math.Uint64.Sum(vs)
			a.ok = true
		} else {
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					a.linreg += vs.Value(i)
					a.ok = true
				}
			}
		}
	}
}
func (a *LinregUIntAgg) Type() flux.ColType {
	return flux.TUInt
}
func (a *LinregUIntAgg) ValueUInt() uint64 {
	return a.linreg
}
func (a *LinregUIntAgg) IsNull() bool {
	return !a.ok
}

type LinregFloatAgg struct {
	linreg []*array.Float64
	ok  bool
	x, y                     []float64
	sx, sy, sxx, sxy, syy, n float64
}



func (a *LinregFloatAgg) DoFloat(vs *array.Float64) {
	vs.Retain()
	a.linreg = append(a.linreg, vs)
}

// test if we can order the dataset CHECK
// test if we can find the median CHECK
// test if we can find the differences between median and each value
// test if we can find the sort that difference
// test if we can find the median value of that

//vs.Retain tells arrow. says I still need this memory.
//DoFloat gets called values get passed in, calling retain says don't release and keep a reference to it. It doesn't reuse it.
//Implementation detail: Arrow can allocate memory and free memory (in the future sync.pool to not have to call make everytime because make is expensive)

func (a *LinregFloatAgg) Type() flux.ColType {
	return flux.TFloat
}
func (a *LinregFloatAgg) ValueFloat() float64 {
	//Iterate over each of the columns
	for _, vs := range a.linreg {
		b := arrow.NewFloatBuilder(nil)
		b.Resize(vs.Len())
		a.n = float64(vs.Len())
		for i, l := 0, vs.Len(); i < l; i++ {
			if vs.IsNull(i) {
				b.AppendNull()
				continue
			}
			v := vs.Value(i)
			diff := v - median
			b.Append(diff)
		}
		vs.Release()
		//Construct the Array
		diffs := b.NewFloat64Array()
		q.DoFloat(diffs)
		diffs.Release()
	}
	a.linreg = nil
	return q.ValueFloat()
}


func (r *LinReg) raw() []float64 {
	return r.y
}

func (r *LinReg) xSum() float64 {
	total := 0
	for i := 0; i < len(r.y); i++ {
		total += i
	}
	r.sx = float64(total)
	return r.sx
}

func (r *LinReg) ySum() float64 {
	total := 0.0
	for _, value := range r.y {
		total += value
	}
	r.sy = total
	return r.sy
}

func (r *LinReg) xSquared() float64 {
	total := 0.0
	for i := 0; i < len(r.y); i++ {
		total += float64(i) * float64(i)
	}
	r.sxx = total
	// fmt.Println("sxx=", r.sxx)
	return r.sxx
}

func (r *LinReg) ySquared() float64 {
	tot	l := 0.0
	for _, value := range r.y {
		total += value * value
	}
	r.syy = total
	// fmt.Println("syy=", r.syy)
	return r.syy
}

func (r *LinReg) xySum() float64 {
	total := 0.0
	for _, value := range r.y {
		total += value * float64(i)
	}
	r.sxy = total
	// fmt.Println("sxy=", r.sxy)
	return r.sxy
}

func (r *LinReg) slope() float64 {
	r.length()
	ss_xy := r.n*r.sxy - r.sx*r.sy
	ss_xx := r.n*r.sxx - r.sx*r.sx
	return ss_xy / ss_xx
}

func (r *LinReg) intercept() float64 {
	r.length()
	return (r.sy - r.slope()*r.sx) / r.n
}


//For each col, compute diff, send them to quantile agg
//Return quanitile agg

//do calculations here. This is for retaining the data
//use dofloat (interface that gets called) to store and keep the arrays using retain on vs
//value float you actually perform the calculation

func (a *LinregFloatAgg) IsNull() bool {
	return !a.ok
}
