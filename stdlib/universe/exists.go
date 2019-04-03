package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

var SpecialFns map[string]values.Function

func init() {
	existsFunc := values.NewFunction(
		"exists",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{"value": semantic.Tvar(1)},
			Required:   semantic.LabelSet{"value"},
			Return:     semantic.Bool,
		}),
		func(args values.Object) (values.Value, error) {
			v, ok := args.Get("value")
			if !ok {
				return nil, errors.New("missing argument value")
			}

			if v.IsNull() {
				return values.NewBool(false), nil
			}
			return values.NewBool(true), nil
		}, false,
	)

	flux.RegisterPackageValue("universe", "exists", existsFunc)

}
