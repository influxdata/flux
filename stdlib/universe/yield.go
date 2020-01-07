package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

const YieldKind = "yield"

type YieldOpSpec struct {
	Name string `json:"name"`
}

func init() {
	yieldSignature := semantic.LookupBuiltInType("universe", "yield")

	flux.RegisterPackageValue("universe", YieldKind, flux.MustValue(flux.FunctionValueWithSideEffect(YieldKind, createYieldOpSpec, yieldSignature)))
	flux.RegisterOpSpec(YieldKind, newYieldOp)
	plan.RegisterProcedureSpecWithSideEffect(YieldKind, newYieldProcedure, YieldKind)
}

func createYieldOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(YieldOpSpec)

	name, ok, err := args.GetString("name")
	if err != nil {
		return nil, err
	} else if ok {
		spec.Name = name
	} else {
		spec.Name = plan.DefaultYieldName
	}

	return spec, nil
}

func newYieldOp() flux.OperationSpec {
	return new(YieldOpSpec)
}

func (s *YieldOpSpec) Kind() flux.OperationKind {
	return YieldKind
}

type YieldProcedureSpec struct {
	plan.DefaultCost
	Name string `json:"name"`
}

func newYieldProcedure(qs flux.OperationSpec, _ plan.Administration) (plan.ProcedureSpec, error) {
	if spec, ok := qs.(*YieldOpSpec); ok {
		return &YieldProcedureSpec{Name: spec.Name}, nil
	}

	return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
}

func (s *YieldProcedureSpec) Kind() plan.ProcedureKind {
	return YieldKind
}

func (s *YieldProcedureSpec) Copy() plan.ProcedureSpec {
	return &YieldProcedureSpec{Name: s.Name}
}

func (s *YieldProcedureSpec) YieldName() string {
	return s.Name
}
