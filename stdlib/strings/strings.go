package strings

import (
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func generateStringFunction(name string, stringFn func(string) string) values.Function {
	return values.NewFunction(
		name,
		semantic.NewFunctionType(semantic.FunctionSignature{
			Parameters: map[string]semantic.Type{stringArg: semantic.String},
			Required:   semantic.LabelSet{stringArg},
			Return:     semantic.String,
		}),
		func(args values.Object) (values.Value, error) {
			var str string

			v, ok := args.Get(stringArg)
			if !ok {
				return nil, errMissingV
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

func init() {
	flux.RegisterPackageValue("strings", "trim", generateMultiArgStringFunction("trim", strings.Trim))
	flux.RegisterPackageValue("strings", "trimSpace", generateStringFunction("trimSpace", strings.TrimSpace))
	flux.RegisterPackageValue("strings", "title", generateStringFunction("title", strings.Title))
	flux.RegisterPackageValue("strings", "toUpper", generateStringFunction("toUpper", strings.ToUpper))
	flux.RegisterPackageValue("strings", "toLower", generateStringFunction("toLower", strings.ToLower))
}
