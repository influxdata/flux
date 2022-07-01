package universe

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
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
