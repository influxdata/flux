package system

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var systemTimeFuncName = "time"

func init() {
	flux.RegisterPackageValue("system", systemTimeFuncName, values.NewFunction(
		systemTimeFuncName,
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Return: semantic.Time,
		}),
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			return values.NewTime(values.ConvertTime(time.Now().UTC())), nil
		},
		false,
	))
}
