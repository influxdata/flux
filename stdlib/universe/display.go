package universe

import (
	"context"

	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/values"
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
