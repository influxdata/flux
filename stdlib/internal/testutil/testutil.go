package testutil

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("internal/testutil", "fail", values.NewFunction(
		"fail",
		semantic.MustLookupBuiltinType("internal/testutil", "fail"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return nil, errors.New(codes.Aborted, "fail")
		},
		false,
	))
	runtime.RegisterPackageValue("internal/testutil", "yield", values.NewFunction(
		"yield",
		semantic.MustLookupBuiltinType("internal/testutil", "yield"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			in, ok := args.Get("v")
			if !ok {
				return nil, errors.Newf(codes.Invalid, "missing required keyword argument %q", "v")
			}
			return in, nil
		}, true))
	runtime.RegisterPackageValue("internal/testutil", "makeRecord", values.NewFunction(
		"makeRecord",
		semantic.MustLookupBuiltinType("internal/testutil", "makeRecord"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			in, ok := args.Get("o")
			if !ok {
				return nil, errors.Newf(codes.Invalid, "missing required keyword argument %q", "o")
			}
			return in, nil
		}, false))
}
