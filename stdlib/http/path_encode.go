package http

import (
	"context"
	"net/url"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/interpreter"
)

func init() {
	runtime.RegisterPackageValue("http", "pathEscape", pathEscapeFunc)
}

const inputStringArg = "x"

var pathEscapeFunc = values.NewFunction(
	"pathEscape",
	runtime.MustLookupBuiltinType("http", "pathEscape"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(PathEncode, args)
	},
	false,
)


func PathEncode(args interpreter.Arguments) (values.Value, error) {
	x, err := args.GetRequiredString(inputStringArg)
	if err != nil {
		return nil, err
	}
	return values.NewString(url.PathEscape(x)), nil
}