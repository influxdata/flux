package system

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/lang/execdeps"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var systemTimeFunc = values.NewFunction(
	"time",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Return: semantic.Time,
	}),
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
	flux.RegisterPackageValue("system", "time", systemTimeFunc)
}
