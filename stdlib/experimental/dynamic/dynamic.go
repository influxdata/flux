package dynamic

import (
	"context"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/dynamic", "dynamic", dynamicConv)
}

var dynamicConv = values.NewFunction(
	"dynamic",
	runtime.MustLookupBuiltinType("experimental/dynamic", "dynamic"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		v, err := arguments.GetRequired("v")
		if err != nil {
			return nil, err
		}

		// FIXME(onelson): do we need to special case nulls here?
		//   Ex: if the input is null, should the output be null or a
		//   Dynamic with a null in it? My sense is the latter.
		return values.NewDynamic(v), nil
	},
	false,
)
