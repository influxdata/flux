package function

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

type Builder struct {
	PackagePath string
}

func ForPackage(name string) Builder {
	return Builder{
		PackagePath: name,
	}
}

type (
	Definition        func(args *Arguments) (values.Value, error)
	DefinitionContext func(ctx context.Context, args *Arguments) (values.Value, error)
)

func (b Builder) Register(name string, fn Definition) {
	b.RegisterContext(name, func(ctx context.Context, args *Arguments) (values.Value, error) {
		return fn(args)
	})
}

func (b Builder) RegisterContext(name string, fn DefinitionContext) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
			return InvokeContext(fn, ctx, args)
		}, false),
	)
}

type (
	Source interface {
		plan.ProcedureSpec
		CreateSource(id execute.DatasetID, a execute.Administration) (execute.Source, error)
	}
	SourceDefinition func(args *Arguments) (Source, error)
	Transformation   interface {
		plan.ProcedureSpec
		CreateTransformation(id execute.DatasetID, a execute.Administration) (execute.Transformation, execute.Dataset, error)
	}
	TransformationDefinition func(args *Arguments) (Transformation, error)
)

func (b Builder) RegisterSource(name string, kind plan.ProcedureKind, fn SourceDefinition) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		flux.MustValue(
			flux.FunctionValue(name, func(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
				return InvokeContext(func(ctx context.Context, args *Arguments) (flux.OperationSpec, error) {
					spec, err := fn(args)
					if err != nil {
						return nil, err
					}
					return &operationSpec{spec: spec}, nil
				}, context.Background(), args.RawObject())
			}, mt),
		),
	)
	plan.RegisterProcedureSpec(kind, newProcedure, flux.OperationKind(kind))
	execute.RegisterSource(kind, func(spec plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
		s, ok := spec.(Source)
		if !ok {
			return nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
		}
		return s.CreateSource(id, a)
	})
}

func (b Builder) RegisterTransformation(name string, kind plan.ProcedureKind, fn TransformationDefinition) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		flux.MustValue(
			flux.FunctionValue(name, func(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
				return InvokeContext(func(ctx context.Context, args *Arguments) (flux.OperationSpec, error) {
					spec, err := fn(args)
					if err != nil {
						return nil, err
					}
					return &operationSpec{spec: spec}, nil
				}, context.Background(), args.RawObject())
			}, mt),
		),
	)
	plan.RegisterProcedureSpec(kind, newProcedure, flux.OperationKind(kind))
	execute.RegisterTransformation(kind, func(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
		s, ok := spec.(Transformation)
		if !ok {
			return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
		}
		return s.CreateTransformation(id, a)
	})
}

// InvokeContext calls a function and returns the result.
//
// It passes the object as the arguments to the function and returns an error
// if any supplied arguments are not used.
func InvokeContext[T any](f func(ctx context.Context, args *Arguments) (T, error), ctx context.Context, argsObj values.Object) (T, error) {
	args := newArguments(argsObj)
	v, err := f(ctx, args)
	if err == nil {
		if unused := args.listUnused(); len(unused) > 0 {
			err = errors.Newf(codes.Invalid, "unused arguments %v", unused)
		}
	}
	return v, err
}

// Invoke calls a function and returns the result.
//
// This is the same as InvokeContext but with a background context.
func Invoke[T any](f func(args *Arguments) (T, error), argsObj values.Object) (T, error) {
	args := newArguments(argsObj)
	v, err := f(args)
	if err == nil {
		if unused := args.listUnused(); len(unused) > 0 {
			err = errors.Newf(codes.Invalid, "unused arguments %v", unused)
		}
	}
	return v, err
}

type operationSpec struct {
	spec plan.ProcedureSpec
}

func (s *operationSpec) Kind() flux.OperationKind {
	return flux.OperationKind(s.spec.Kind())
}

func newProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*operationSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return spec.spec, nil
}
