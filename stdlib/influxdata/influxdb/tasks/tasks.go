package tasks

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const LastSuccessKind = "lastSuccess"

func init() {
	runtime.RegisterPackageValue("influxdata/influxdb/tasks", LastSuccessKind, LastSuccessFunction)
}

// LastSuccessFunction is a function that calls LastSuccess.
var LastSuccessFunction = makeLastSuccessFunc()

func makeLastSuccessFunc() values.Function {
	sig := runtime.MustLookupBuiltinType("influxdata/influxdb/tasks", "lastSuccess")
	return values.NewFunction("lastSuccess", sig, LastSuccess, false)
}

// LastSuccess retrieves the last successful run of the task, or returns the value of the
// orTime parameter if the task has never successfully run.
func LastSuccess(ctx context.Context, args values.Object) (values.Value, error) {
	fargs := interpreter.NewArguments(args)
	_, err := fargs.GetRequired("orTime")
	if err != nil {
		return nil, err
	}

	return nil, errors.Newf(codes.Unimplemented, "This function is not yet implemented.")
}
