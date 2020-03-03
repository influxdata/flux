package system

import (
	"context"
	"time"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

var systemTimeFuncName = "time"

func init() {
	runtime.RegisterPackageValue("system", systemTimeFuncName, values.NewFunction(
		systemTimeFuncName,
		runtime.MustLookupBuiltinType("system", systemTimeFuncName),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewTime(values.ConvertTime(time.Now().UTC())), nil
		},
		false,
	))
}
