package universe

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/interpreter"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

// MakeLengthFunc create the "length()" function.
//
// Length will return the length of the given arr array.
func MakeLengthFunc() values.Function {
	return values.NewFunction(
		"length",
		runtime.MustLookupBuiltinType("universe", "length"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			v, err := a.GetRequired("arr")
			if err != nil {
				return nil, err
			} else if got := v.Type().Nature(); got != semantic.Array {
				return nil, errors.Newf(codes.Invalid, "arr must be an array, got %s", got)
			}

			arr := v.Array()
			l := arr.Len()
			return values.NewInt(int64(l)), nil
		}, false,
	)
}

func init() {
	runtime.RegisterPackageValue("universe", "length", MakeLengthFunc())
}
