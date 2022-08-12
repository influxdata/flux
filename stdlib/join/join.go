package join

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const Join2Kind = "join.tables"

func init() {
	signature := runtime.MustLookupBuiltinType("join", "tables")
	runtime.RegisterPackageValue(
		"join", "tables", flux.MustValue(flux.FunctionValue("tables", createJoinOpSpec, signature)),
	)
	plan.RegisterProcedureSpec(Join2Kind, newJoinProcedure, Join2Kind)
	execute.RegisterTransformation(Join2Kind, createJoinTransformation)
}

type JoinOpSpec struct {
	on     interpreter.ResolvedFunction
	as     interpreter.ResolvedFunction
	left   *flux.TableObject
	right  *flux.TableObject
	method string
}

func (o *JoinOpSpec) Kind() flux.OperationKind {
	return flux.OperationKind(Join2Kind)
}

func createJoinOpSpec(args flux.Arguments, p *flux.Administration) (flux.OperationSpec, error) {
	l, ok := args.Get("left")
	if !ok {
		return nil, errors.New(codes.Invalid, "missing required argument 'left'")
	}
	left, ok := l.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'left' must be a table stream")
	}
	p.AddParent(left)

	r, ok := args.Get("right")
	if !ok {
		return nil, errors.New(codes.Invalid, "missing required argument 'right'")
	}
	right, ok := r.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'right' must be a table stream")
	}
	p.AddParent(right)

	o, err := args.GetRequiredFunction("on")
	if err != nil {
		return nil, err
	}
	on, err := interpreter.ResolveFunction(o)
	if err != nil {
		return nil, err
	}

	a, err := args.GetRequiredFunction("as")
	if err != nil {
		return nil, err
	}
	as, err := interpreter.ResolveFunction(a)
	if err != nil {
		return nil, err
	}

	method, err := args.GetRequiredString("method")
	if err != nil {
		return nil, err
	}

	if method != "inner" && method != "left" && method != "right" && method != "full" {
		return nil, errors.New(
			codes.Invalid,
			"invalid argument for 'method' - must be \"inner\", \"left\", \"right\", or \"full\"",
		)
	}

	op := JoinOpSpec{
		left:   left,
		right:  right,
		on:     on,
		as:     as,
		method: method,
	}
	return &op, nil
}

type JoinProcedureSpec struct {
	On     interpreter.ResolvedFunction
	As     interpreter.ResolvedFunction
	Left   *flux.TableObject
	Right  *flux.TableObject
	Method string
}

func (p *JoinProcedureSpec) Kind() plan.ProcedureKind {
	return plan.ProcedureKind(Join2Kind)
}

func (p *JoinProcedureSpec) Copy() plan.ProcedureSpec {
	return &JoinProcedureSpec{
		On:     p.On,
		As:     p.As,
		Left:   p.Left,
		Right:  p.Right,
		Method: p.Method,
	}
}

func newJoinProcedure(spec flux.OperationSpec, p plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := spec.(*JoinOpSpec)
	if !ok {
		return nil, errors.New(codes.Internal, "invalid op spec for join procedure")
	}
	proc := JoinProcedureSpec{
		On:     s.on,
		As:     s.as,
		Left:   s.left,
		Right:  s.right,
		Method: s.method,
	}
	return &proc, nil
}

func createJoinTransformation(
	id execute.DatasetID,
	mode execute.AccumulationMode,
	spec plan.ProcedureSpec,
	a execute.Administration,
) (execute.Transformation, execute.Dataset, error) {
	t, err := NewMergeJoinTransformation(
		a.Context(),
		id,
		spec,
		a.Parents()[0],
		a.Parents()[1],
		a.Allocator(),
	)
	if err != nil {
		return nil, nil, err
	}
	tr := execute.NewTransformationFromTransport(t)
	return tr, t.d, nil
}
