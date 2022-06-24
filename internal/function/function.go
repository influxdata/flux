package function

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux/interpreter"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

type Builder struct {
	PackagePath string
}

func ForPackage(name string) Builder {
	return Builder{
		PackagePath: name,
	}
}

type Definition func(args interpreter.Arguments) (values.Value, error)

func (b Builder) Register(name string, fn Definition) {
	mt := runtime.MustLookupBuiltinType(b.PackagePath, name)
	runtime.RegisterPackageValue(b.PackagePath, name,
		values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCall(fn, args)
		}, false),
	)
}
