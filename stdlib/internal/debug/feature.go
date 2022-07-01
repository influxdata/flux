package debug

import (
	"context"

	"github.com/influxdata/flux/internal/feature"
	featurepkg "github.com/influxdata/flux/internal/pkg/feature"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	mt := runtime.MustLookupBuiltinType("internal/debug", "feature")
	runtime.RegisterPackageValue("internal/debug", "feature",
		values.NewFunction("feature", mt, func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(Feature, ctx, args)
		}, false),
	)
}

func Feature(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
	key, err := args.GetRequiredString("key")
	if err != nil {
		return nil, err
	}

	flag, ok := feature.ByKey(key)
	if !ok {
		return values.Null, nil
	}

	flagger := featurepkg.GetFlagger(ctx)
	v := flagger.FlagValue(ctx, flag)
	if iv, ok := v.(int); ok {
		v = int64(iv)
	}
	return values.New(v), nil
}
