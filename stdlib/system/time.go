package system

import (
	"context"
	"time"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var systemTimeFuncName = "time"

func init() {
	runtime.RegisterPackageValue("system", systemTimeFuncName, values.NewFunction(
		systemTimeFuncName,
		semantic.NewFunctionType(semantic.BasicTime, nil),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewTime(values.ConvertTime(time.Now().UTC())), nil
		},
		false,
	))
}
