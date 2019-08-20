package universe

import (
	"errors"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// MakeContainsFunc will construct the "contains()" function.
//
// Contains will test whether a given value is a member of the given set array.
func MakeContainsFunc() values.Function {
	return values.NewFunction(
		"contains",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"value": semantic.Tvar(1),
				"set":   semantic.NewArrayPolyType(semantic.Tvar(1)),
			},
			Required: semantic.LabelSet{"value", "set"},
			Return:   semantic.Bool,
		}),
		func(args values.Object) (values.Value, error) {
			v, ok := args.Get("value")
			if !ok {
				return nil, errors.New("missing argument value")
			}

			setarg, ok := args.Get("set")
			if !ok {
				return nil, errors.New("missing argument set")
			}

			set := setarg.Array()
			found := false
			var err error
			if set.Len() > 0 {
				if set.Get(0).Type() != v.Type() {
					err = fmt.Errorf("value type %T does not match set type %T", v.Type(), set.Get(0).Type())
				} else {
					for i := 0; i < set.Len(); i++ {
						if set.Get(i).Equal(v) {
							found = true
							break
						}
					}
				}
			}

			return values.NewBool(found), err
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("universe", "contains", MakeContainsFunc())
}
