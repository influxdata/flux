package universe

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
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

			// In situations where value `v` is "invalid" (like when it's a field that is missing from an incoming
			// unknown record type, we give a surprising and confusing error message indicating
			// `keyword argument "set" should be of an array of type invalid`.
			// This error is normally produced by the `GetRequiredArrayAllowEmpty` call immediately below.
			// To avoid the confusion and handle this case more intuitively, short-circuit with a "false" outcome.
			if v.Type().Nature() == semantic.Invalid {
				return values.NewBool(false), nil
			}
			setarg, err := a.GetRequiredArrayAllowEmpty("set", v.Type().Nature())
			if err != nil {
				return nil, err
			}

			set := setarg.Array()
			found := false

			if set.Len() > 0 {
				for i := 0; i < set.Len(); i++ {
					// Skip any members of the `set` array that are invalid.
					if set.Get(i).Type().Nature() == semantic.Invalid {
						continue
					} else if set.Get(i).Type() != v.Type() {
						err = errors.Newf(codes.Invalid, "value type %T does not match set type %T", v.Type(), set.Get(0).Type())
						break
					}
					if set.Get(i).Equal(v) {
						found = true
						break
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
