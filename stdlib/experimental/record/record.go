package record

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const packagePath = "experimental/record"

func init() {
	runtime.RegisterPackageValue(packagePath, "any", values.NewObjectWithValues(nil))
	runtime.RegisterPackageValue(packagePath, "get", values.NewFunction(
		"get",
		runtime.MustLookupBuiltinType(packagePath, "get"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
				r, err := args.GetRequiredObject("r")
				if err != nil {
					return nil, err
				}

				key, err := args.GetRequiredString("key")
				if err != nil {
					return nil, err
				}

				def, err := args.GetRequired("default")
				if err != nil {
					return nil, err
				}

				if v, ok := r.Get(key); ok {
					return v, nil
				}

				return def, nil
			}, ctx, args)
		}, false,
	))
	runtime.RegisterPackageValue(packagePath, "at", values.NewFunction(
		"at",
		runtime.MustLookupBuiltinType(packagePath, "at"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
				r, err := args.GetRequiredObject("r")
				if err != nil {
					return nil, err
				}

				key, err := args.GetRequiredString("key")
				if err != nil {
					return nil, err
				}

				if v, ok := r.Get(key); ok {
					return v, nil
				}

				return nil, errors.Newf(codes.Internal, "record.at property %q doesn't exist, type checking failed to ensure it did", key)
			}, ctx, args)
		}, false,
	))
}
