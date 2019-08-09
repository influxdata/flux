package env

import (
	"errors"
	"os"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var getStringName = "getString"

func init() {
	flux.RegisterPackageValue("env", getStringName, getString())
}

// getString returns a function value that when called will get a string from the OS environment
func getString() values.Function {
	ftype := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"name":    semantic.String,
			"default": semantic.String,
		},
		Required: semantic.LabelSet{"name"},
		Return:   semantic.String,
	})
	call := func(args values.Object) (values.Value, error) {
		name, ok := args.Get("name")
		if !ok {
			return nil, errors.New("missing argument name")
		}
		if name.Type().Nature() != semantic.String {
			return nil, errors.New("name argument is not of type String")
		}
		return values.NewString(os.Getenv(name.Str())), nil
	}
	sideEffect := true
	return values.NewFunction(getStringName, ftype, call, sideEffect)
}
