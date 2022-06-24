package universe

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

func init() {
	runtime.RegisterPackageValue("universe", "display", values.NewFunction(
		"display",
		runtime.MustLookupBuiltinType("universe", "display"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			v, ok := args.Get("v")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing argument v")
			}
			return values.NewString(values.DisplayString(v)), nil
		},
		false,
	))
}
