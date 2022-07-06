package experimental

import (
	"context"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const CatchKind = "catch"
const pkg = "experimental"

func init() {
	runtime.RegisterPackageValue(pkg, CatchKind, Catch())
}

func Catch() values.Function {
	return values.NewFunction(
		CatchKind,
		runtime.MustLookupBuiltinType(pkg, CatchKind),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
				fn, err := args.GetRequiredFunction("fn")
				if err != nil {
					return nil, err
				}

				value, err := fn.Call(ctx, values.NewObject(semantic.NewObjectType(nil)))

				if err == nil {
					return values.Stringify(value)
				}

				return values.NewString(err.Error()), nil
			}, ctx, args)
		}, false,
	)
}
