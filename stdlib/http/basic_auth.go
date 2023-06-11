package http

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/InfluxCommunity/flux/interpreter"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/values"
)

func init() {
	runtime.RegisterPackageValue("http", "basicAuth", basicAuthFunc)
}

const (
	basicAuthUsernameArg = "u"
	basicAuthPasswordArg = "p"
)

var basicAuthFunc = values.NewFunction(
	"basicAuth",
	runtime.MustLookupBuiltinType("http", "basicAuth"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(BasicAuth, args)
	},
	false,
)

func BasicAuth(args interpreter.Arguments) (values.Value, error) {
	u, err := args.GetRequiredString(basicAuthUsernameArg)
	if err != nil {
		return nil, err
	}

	p, err := args.GetRequiredString(basicAuthPasswordArg)
	if err != nil {
		return nil, err
	}

	combined := fmt.Sprintf("%s:%s", u, p)
	v := base64.StdEncoding.EncodeToString([]byte(combined))
	return values.NewString("Basic " + v), nil
}
