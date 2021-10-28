package runtime

import (
	"context"

	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var nowFuncName = "now"

func init() {
	runtime.RegisterPackageValue("runtime", nowFuncName, values.NewFunction(
		nowFuncName,
		semantic.NewFunctionType(semantic.BasicTime, nil),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewTime(values.ConvertTime(lang.GetRuntimeNow(ctx))), nil
		},
		false,
	))
}
