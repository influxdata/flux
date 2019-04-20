package strings

import (
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	stringArg = "v"
	cutset    = "cutset"
	prefix    = "prefix"
	suffix    = "suffix"
)

func generateSingleArgStringFunction(name string, stringFn func(string) string) values.Function {
	return values.NewFunction(
		name,
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{stringArg: semantic.String},
			Required:   semantic.LabelSet{stringArg},
			Return:     semantic.String,
		}),
		func(args values.Object) (values.Value, error) {
			var str string

			v, ok := args.Get(stringArg)
			if !ok {
				return nil, fmt.Errorf("missing argument %q", stringArg)
			}

			if v.Type().Nature() == semantic.String {
				str = v.Str()

				str = stringFn(str)
				return values.NewString(str), nil
			}

			return nil, fmt.Errorf("cannot convert argument of type %v to upper case", v.Type().Nature())
		}, false,
	)
}

func generateDualArgStringFunction(name string, argNames []string, stringFn func(string, string) string) values.Function {
	if len(argNames) != 2 {
		panic("unexpected number of argument names")
	}

	return values.NewFunction(
		name,
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				argNames[0]: semantic.String,
				argNames[1]: semantic.String,
			},
			Required: semantic.LabelSet{argNames[0], argNames[1]},
			Return:   semantic.String,
		}),
		func(args values.Object) (values.Value, error) {
			var argVals = make([]values.Value, 2)

			for i, name := range argNames {
				val, ok := args.Get(name)
				if !ok {
					return nil, fmt.Errorf("missing argument %q", name)
				}

				if val.Type().Nature() != semantic.String {
					return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", name, semantic.String, val.Type().Nature())
				}

				argVals[i] = val
			}

			return values.NewString(stringFn(argVals[0].Str(), argVals[1].Str())), nil
		},
		false,
	)
}

func init() {
	flux.RegisterPackageValue("strings", "trim", generateDualArgStringFunction("trim", []string{stringArg, cutset}, strings.Trim))
	flux.RegisterPackageValue("strings", "trimSpace", generateSingleArgStringFunction("trimSpace", strings.TrimSpace))
	flux.RegisterPackageValue("strings", "trimPrefix", generateDualArgStringFunction("trimSuffix", []string{stringArg, prefix}, strings.TrimPrefix))
	flux.RegisterPackageValue("strings", "trimSuffix", generateDualArgStringFunction("trimSuffix", []string{stringArg, suffix}, strings.TrimSuffix))
	flux.RegisterPackageValue("strings", "title", generateSingleArgStringFunction("title", strings.Title))
	flux.RegisterPackageValue("strings", "toUpper", generateSingleArgStringFunction("toUpper", strings.ToUpper))
	flux.RegisterPackageValue("strings", "toLower", generateSingleArgStringFunction("toLower", strings.ToLower))
}
