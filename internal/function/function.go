package function

import (
	"context"

	"github.com/influxdata/flux/interpreter"
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
	Definition        func(args interpreter.Arguments) (values.Value, error)
	DefinitionContext func(ctx context.Context, args interpreter.Arguments) (values.Value, error)
)

func (b Builder) Register(name string, fn Definition) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCall(fn, args)
		}, false),
	)
}

func (b Builder) RegisterContext(name string, fn DefinitionContext) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(fn, ctx, args)
		}, false),
	)
}
