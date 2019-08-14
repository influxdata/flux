package experimental

import (
	"context"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("experimental", "objectKeys", values.NewFunction(
		"objectKeys",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"o": semantic.Tvar(1),
			},
			Required: []string{"o"},
			Return:   semantic.NewArrayPolyType(semantic.String),
		}),
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			o, ok := args.Get("o")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing parameter \"o\"")
			}
			if o.Type().Nature() != semantic.Object {
				return nil, errors.New(codes.Invalid, "parameter \"o\" is not an object")
			}
			obj := o.Object()
			keys := make([]values.Value, 0, obj.Len())
			obj.Range(func(name string, _ values.Value) {
				keys = append(keys, values.NewString(name))
			})
			return values.NewArrayWithBacking(semantic.String, keys), nil
		},
		false,
	))
}
