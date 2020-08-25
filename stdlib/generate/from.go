package generate

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FromGeneratorKind = "fromGenerator"

type FromGeneratorOpSpec struct {
	Start flux.Time                    `json:"start"`
	Stop  flux.Time                    `json:"stop"`
	Count int64                        `json:"count"`
	Fn    interpreter.ResolvedFunction `json:"fn"`
}

func init() {
	fromGeneratorSignature := runtime.MustLookupBuiltinType("generate", "from")
	runtime.RegisterPackageValue("generate", "from", flux.MustValue(flux.FunctionValue(FromGeneratorKind, createFromGeneratorOpSpec, fromGeneratorSignature)))
	flux.RegisterOpSpec(FromGeneratorKind, newFromGeneratorOp)
	plan.RegisterProcedureSpec(FromGeneratorKind, newFromGeneratorProcedure, FromGeneratorKind)
	execute.RegisterSource(FromGeneratorKind, createFromGeneratorSource)
}

func createFromGeneratorOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromGeneratorOpSpec)

	if t, err := args.GetRequiredTime("start"); err != nil {
		return nil, err
	} else {
		spec.Start = t
	}
	if t, err := args.GetRequiredTime("stop"); err != nil {
		return nil, err
	} else {
		spec.Stop = t
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

func newFromGeneratorOp() flux.OperationSpec {
	return new(FromGeneratorOpSpec)
}

func (s *FromGeneratorOpSpec) Kind() flux.OperationKind {
	return FromGeneratorKind
}

type FromGeneratorProcedureSpec struct {
	plan.DefaultCost
	Start time.Time
	Stop  time.Time
	Count int64
	Fn    interpreter.ResolvedFunction
}

func newFromGeneratorProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	// TODO: copy over data from the OpSpec to the ProcedureSpec
	spec, ok := qs.(*FromGeneratorOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromGeneratorProcedureSpec{
		Count: spec.Count,
		Start: spec.Start.Time(pa.Now()),
		Stop:  spec.Stop.Time(pa.Now()),
		Fn:    spec.Fn,
	}, nil
}

func (s *FromGeneratorProcedureSpec) Kind() plan.ProcedureKind {
	return FromGeneratorKind
}

func (s *FromGeneratorProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromGeneratorProcedureSpec)

	return ns
}

func createFromGeneratorSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromGeneratorProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	s := NewGeneratorSource(a.Allocator())
	s.Start = spec.Start
	s.Stop = spec.Stop
	s.Count = spec.Count
	fn, err := compiler.Compile(compiler.ToScope(spec.Fn.Scope), spec.Fn.Fn, semantic.NewObjectType(
		[]semantic.PropertyType{
			{Key: []byte("n"), Value: semantic.BasicInt},
		},
	))
	if err != nil {
		return nil, err
	} else if n := fn.Type().Nature(); n != semantic.Int {
		return nil, errors.Newf(codes.Invalid, "function must return type integer, but got %s", n)
	}
	s.Fn = fn

	return execute.CreateSourceFromDecoder(s, dsid, a)
}

type GeneratorSource struct {
	done  bool
	Start time.Time
	Stop  time.Time
	Count int64
	alloc *memory.Allocator
	Fn    compiler.Func
}

func NewGeneratorSource(a *memory.Allocator) *GeneratorSource {
	return &GeneratorSource{alloc: a}
}

func (s *GeneratorSource) Connect(ctx context.Context) error {
	return nil
}

func (s *GeneratorSource) Fetch(ctx context.Context) (bool, error) {
	return !s.done, nil
}

func (s *GeneratorSource) Decode(ctx context.Context) (flux.Table, error) {
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
		values.NewTime(values.ConvertTime(s.Start)),
		values.NewTime(values.ConvertTime(s.Stop)),
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
		_, err := b.AddCol(col)
		if err != nil {
			return nil, err
		}
	}

	cols = b.Cols()

	deltaT := s.Stop.Sub(s.Start) / time.Duration(s.Count)
	timeIdx := execute.ColIdx("_time", cols)
	valueIdx := execute.ColIdx("_value", cols)
	in := values.NewObject(semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("n"), Value: semantic.BasicInt},
	}))
	for i := 0; i < int(s.Count); i++ {
		b.AppendTime(timeIdx, values.ConvertTime(s.Start.Add(time.Duration(i)*deltaT)))
		in.Set("n", values.NewInt(int64(i)))
		v, err := s.Fn.Eval(ctx, in)
		if err != nil {
			return nil, err
		}
		if err := b.AppendValue(valueIdx, v); err != nil {
			return nil, err
		}
	}

	return b.Table()
}

func (s *GeneratorSource) Close() error {
	return nil
}
