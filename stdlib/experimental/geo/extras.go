package geo

import (
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func generateFunc() values.Function {
	return values.NewFunction(
		"containsTag",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"row":    semantic.Tvar(1),
				"tagKey": semantic.String,
				"set":    semantic.NewArrayPolyType(semantic.String),
			},
			Required: semantic.LabelSet{"row", "tagKey", "set"},
			Return:   semantic.Bool,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			row, err := a.GetRequired("row")
			if err != nil {
				return nil, err
			}

			tagKey, err := a.GetRequiredString("tagKey")
			if err != nil {
				return nil, err
			}

			setarg, err := a.GetRequiredArray("set", semantic.String)
			if err != nil {
				return nil, err
			}

			v, vOk := row.Object().Get(tagKey)
			if !vOk || v.IsNull() {
				return values.NewBool(false), nil
			}

			switch v.Type().Nature() {
			case semantic.String:
				found := false
				set := setarg.Array()
				for i := 0; i < set.Len(); i++ {
					if set.Get(i).Equal(v) {
						found = true
						break
					}
				}
				return values.NewBool(found), nil
			default:
				return nil, fmt.Errorf("code %d: cannot use %v as string; %s may reference a field", codes.Invalid, v.Type(), tagKey)
			}
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("experimental/geo", "containsTag", generateFunc())
}
