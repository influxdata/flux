package universe

import (
	"fmt"

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
			Parameters: map[string]semantic.PolyType{"r": semantic.Tvar(1), "key": semantic.String},
			Required:   semantic.LabelSet{"r", "key"},
			Return:     semantic.Bool,
		}),
		func(args values.Object) (values.Value, error) {
			r, ok := args.Get("r")
			if !ok {
				return nil, errors.New("missing argument r")
			}
			//if r.Type() != *semantic.Object {
			//	return nil, fmt.Errorf("argument r: expected type %V, got type %T", semantic.Object, r.Type())
			//}

			key, ok := args.Get("key")
			if !ok {
				return nil, errors.New("missing argument k")
			}
			if key.Type() != semantic.String {
				return nil, fmt.Errorf("argument key: expected type %T got type %T", semantic.String, key.Type())
			}

			robj := r.Object()
			v, ok := robj.Get(key.Str())

			if !ok || v.IsNull() {
				return values.NewBool(false), nil
			}
			return values.NewBool(true), nil
		}, false,
	)

	flux.RegisterPackageValue("universe", "exists", existsFunc)

}
