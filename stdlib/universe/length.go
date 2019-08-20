package universe

import (
	"errors"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// MakeLengthFunc create the "length()" function.
//
// Length will return the length of the given arr array.
func MakeLengthFunc() values.Function {
	return values.NewFunction(
		"length",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"arr": semantic.NewArrayPolyType(semantic.Tvar(1)),
			},
			Required: semantic.LabelSet{"arr"},
			Return:   semantic.Int,
		}),
		func(args values.Object) (values.Value, error) {
			v, ok := args.Get("arr")
			if !ok {
				return nil, errors.New("missing argument value")
			}
			l := v.Array().Len()
			return values.NewInt(int64(l)), nil
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("universe", "length", MakeLengthFunc())
}
