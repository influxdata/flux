package secrets

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const GetKind = "get"

func init() {
	flux.RegisterPackageValue("influxdata/influxdb/secrets", GetKind, GetFunc)
}

// GetFunc is a function that calls Get.
var GetFunc = makeGetFunc()

func makeGetFunc() values.Function {
	sig := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"key": semantic.String,
		},
		Required: semantic.LabelSet{"key"},
		Return: semantic.NewObjectPolyType(
			map[string]semantic.PolyType{
				"secretKey": semantic.String,
			},
			semantic.LabelSet{"secretKey"},
			semantic.LabelSet{"secretKey"},
		),
	})
	return values.NewFunction("get", sig, Get, false)
}

// Get retrieves the secret key identifier for a given secret.
func Get(args values.Object) (values.Value, error) {
	fargs := interpreter.NewArguments(args)
	key, err := fargs.GetRequiredString("key")
	if err != nil {
		return nil, err
	}
	return New(key), nil
}
