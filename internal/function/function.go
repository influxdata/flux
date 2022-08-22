package function

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
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
