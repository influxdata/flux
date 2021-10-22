package mqtt

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/mqtt", "publish", values.NewFunction(
		"publish",
		runtime.MustLookupBuiltinType("experimental/mqtt", "publish"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
				spec := &CommonMQTTOpSpec{}
				if err := spec.ReadArgs(flux.Arguments{Arguments: args}); err != nil {
					return nil, err
				}

				topic, err := args.GetRequiredString("topic")
				if err != nil {
					return nil, err
				}
				if topic == "" {
					return nil, errors.New(codes.Invalid, "empty topic")
				}

				message, err := args.GetRequiredString("message")
				if err != nil {
					return nil, err
				}
				if message == "" {
					return nil, errors.New(codes.Invalid, "empty message")
				}

				published, err := publish(ctx, topic, message, spec)
				if err != nil {
					return nil, err
				}

				return values.NewBool(published), nil
			}, ctx, args)
		}, false,
	))
}
