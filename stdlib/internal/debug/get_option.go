package debug

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func getStringArgument(args values.Object, name string) (string, error) {
	v, ok := args.Get(name)
	if !ok {
		return "", errors.Newf(codes.Invalid, "missing argument %s", name)
	}
	if v.Type().Nature() != semantic.String {
		return "", errors.Newf(codes.Invalid, "cannot convert argument of type %v to string", v.Type().Nature())
	}
	return v.Str(), nil
}

func init() {
	name := "getOption"
	runtime.RegisterPackageValue("internal/debug", name, values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("internal/debug", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			pkg, err := getStringArgument(args, "pkg")
			if err != nil {
				return nil, err
			}

			name, err := getStringArgument(args, "name")
			if err != nil {
				return nil, err
			}

			v, ok := interpreter.GetOption(ctx, pkg, name)
			if !ok {
				return nil, errors.Newf(codes.Invalid, "option does not exist")
			}

			return v, nil
		}, false,
	))
}
