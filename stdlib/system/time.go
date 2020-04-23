package system

import (
	"context"
	"time"

	"github.com/influxdata/flux/lang/execdeps"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

var systemTimeFunc = values.NewFunction(
	"time",
	runtime.MustLookupBuiltinType("system", "time"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		if execdeps.HaveExecutionDependencies(ctx) {
			if dep := execdeps.GetExecutionDependencies(ctx); dep.Now != nil {
				return values.NewTime(values.ConvertTime(*dep.Now)), nil
			}
		}
		return values.NewTime(values.ConvertTime(time.Now().UTC())), nil
	},
	false,
)

func init() {
	runtime.RegisterPackageValue("system", "time", systemTimeFunc)
}
