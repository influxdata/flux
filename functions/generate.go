package functions

import (
	"fmt"
	"github.com/influxdata/flux/values"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const GenerateKind = "generate"

type GenerateOpSpec struct {
	Start time.Time                    `json:"start"`
	Stop  time.Time                    `json:"stop"`
	Count int64                        `json:"count"`
	Fn    *semantic.FunctionExpression `json:"fn"`
}

var generateSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		"start": semantic.Time,
		"stop":  semantic.Time,
		"count": semantic.Int,
		"fn":    semantic.Function,
	},
	ReturnType: flux.TableObjectType,
}

func init() {
	flux.RegisterFunction(GenerateKind, createGenerateOpSpec, generateSignature)
	flux.RegisterOpSpec(GenerateKind, newGenerateOp)
	plan.RegisterProcedureSpec(GenerateKind, newGenerateProcedure, GenerateKind)
	execute.RegisterSource(GenerateKind, createGenerateSource)
}

func createGenerateOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(GenerateOpSpec)

	if t, err := args.GetRequiredTime("start"); err != nil {
		return nil, err
	} else {
		spec.Start = t.Time(time.Now())
	}

	if t, err := args.GetRequiredTime("stop"); err != nil {
		return nil, err
	} else {
		spec.Stop = t.Time(time.Now())
	}

	if i, err := args.GetRequiredInt("count"); err != nil {
		return nil, err
	} else {
		spec.Count = i
	}

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.Fn = fn
	}

	return spec, nil
}

func newGenerateOp() flux.OperationSpec {
	return new(GenerateOpSpec)
}

func (s *GenerateOpSpec) Kind() flux.OperationKind {
	return GenerateKind
}

type GenerateProcedureSpec struct {
	Start time.Time
	Stop  time.Time
	Count int64
	Param string
	Fn    compiler.Func
}

func newGenerateProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	// TODO: copy over data from the OpSpec to the ProcedureSpec
	spec, ok := qs.(*GenerateOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	fn, param, err := compileFnParam(spec.Fn, semantic.Int, semantic.Int)
	if err != nil {
		return nil, err
	}
	return &GenerateProcedureSpec{
		Count: spec.Count,
		Start: spec.Start,
		Stop:  spec.Stop,
		Param: param,
		Fn:    fn,
	}, nil
}

func (s *GenerateProcedureSpec) Kind() plan.ProcedureKind {
	return GenerateKind
}

func (s *GenerateProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(GenerateProcedureSpec)

	return ns
}

func createGenerateSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*GenerateProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	s := NewGeneratorSource(a.Allocator())
	s.Start = spec.Start
	s.Stop = spec.Stop
	s.Count = spec.Count
	s.Param = spec.Param
	s.Fn = spec.Fn

	return CreateFromSourceIterator(s, dsid, a)
}

type GeneratorSource struct {
	done  bool
	Start time.Time
	Stop  time.Time
	Count int64
	alloc *execute.Allocator
	Fn    compiler.Func
	Param string
}

func NewGeneratorSource(a *execute.Allocator) *GeneratorSource {
	return &GeneratorSource{alloc: a}
}

func (s *GeneratorSource) Connect() error {
	return nil
}

func (s *GeneratorSource) Fetch() (bool, error) {
	return !s.done, nil
}

func (s *GeneratorSource) Decode() (flux.Table, error) {
	defer func() {
		s.done = true
	}()
	ks := []flux.ColMeta{
		flux.ColMeta{
			Label: "_start",
			Type:  flux.TTime,
		},
		flux.ColMeta{
			Label: "_stop",
			Type:  flux.TTime,
		},
	}
	vs := []values.Value{
		values.NewTimeValue(values.ConvertTime(s.Start)),
		values.NewTimeValue(values.ConvertTime(s.Stop)),
	}
	groupKey := execute.NewGroupKey(ks, vs)
	b := execute.NewColListTableBuilder(groupKey, s.alloc)

	cols := []flux.ColMeta{
		flux.ColMeta{
			Label: "_time",
			Type:  flux.TTime,
		},
		flux.ColMeta{
			Label: "_value",
			Type:  flux.TInt,
		},
	}

	for _, col := range cols {
		b.AddCol(col)
	}

	cols = b.Cols()

	colIndex := map[string]int{}
	for i, col := range cols {
		colIndex[col.Label] = i
	}

	deltaT := s.Stop.Sub(s.Start) / time.Duration(s.Count)
	timeIdx := execute.ColIdx("_time", cols)
	valueIdx := execute.ColIdx("_value", cols)
	for i := 0; i < int(s.Count); i++ {
		b.AppendTime(timeIdx, values.ConvertTime(s.Start.Add(time.Duration(i)*deltaT)))
		scope := map[string]values.Value{s.Param: values.NewIntValue(int64(i))}
		v, err := s.Fn.EvalInt(scope)
		if err != nil {
			return nil, err
		}
		b.AppendInt(valueIdx, v)
	}

	return b.Table()
}

