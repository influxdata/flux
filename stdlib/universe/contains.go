package universe

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

// MakeContainsFunc will construct the "contains()" function.
//
// Contains will test whether a given value is a member of the given set array.
func MakeContainsFunc() values.Function {
	return values.NewFunction(
		"contains",
		runtime.MustLookupBuiltinType("universe", "contains"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			v, err := a.GetRequired("value")
			if err != nil {
				return nil, err
			}

			setarg, err := a.GetRequiredArrayAllowEmpty("set", v.Type().Nature())
			if err != nil {
				return nil, err
			}

			set := setarg.Array()
			found := false
			if set.Len() > 0 {
				if set.Get(0).Type() != v.Type() {
					err = errors.Newf(codes.Invalid, "value type %T does not match set type %T", v.Type(), set.Get(0).Type())
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
	runtime.RegisterPackageValue("universe", "contains", MakeContainsFunc())
}
