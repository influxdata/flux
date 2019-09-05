package http

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("http", "basicAuth", basicAuthFunc)
}

const (
	basicAuthUsernameArg = "u"
	basicAuthPasswordArg = "p"
)

var basicAuthFunc = values.NewFunction(
	"basicAuth",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			basicAuthUsernameArg: semantic.String,
			basicAuthPasswordArg: semantic.String,
		},
		Required: semantic.LabelSet{basicAuthUsernameArg, basicAuthPasswordArg},
		Return:   semantic.String,
	}),
	func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
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
