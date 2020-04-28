package secrets

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const GetKind = "get"

func init() {
	runtime.RegisterPackageValue("influxdata/influxdb/secrets", GetKind, GetFunc)
}

// GetFunc is a function that calls Get.
var GetFunc = makeGetFunc()

func makeGetFunc() values.Function {
	sig := runtime.MustLookupBuiltinType("influxdata/influxdb/secrets", "get")
	return values.NewFunction("get", sig, Get, false)
}

// Get retrieves the secret key identifier for a given secret.
func Get(ctx context.Context, args values.Object) (values.Value, error) {
	fargs := interpreter.NewArguments(args)
	key, err := fargs.GetRequiredString("key")
	if err != nil {
		return nil, err
	}

	ss, err := flux.GetDependencies(ctx).SecretService()
	if err != nil {
		return nil, errors.Wrapf(err, codes.Inherit, "cannot retrieve secret %q", key)
	}

	value, err := ss.LoadSecret(ctx, key)
	if err != nil {
		return nil, err
	}
	return values.NewString(value), nil
}
