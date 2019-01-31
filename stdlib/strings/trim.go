package strings

import (
	"fmt"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func generateMultiArgStringFunction(name string, stringFn func(string, string) string) values.Function {
	return values.NewFunction(
		name,
		semantic.NewFunctionType(semantic.FunctionSignature{
			Parameters: map[string]semantic.Type{
				stringArg: semantic.String,
				cutset:    semantic.String,
			},
			Required: semantic.LabelSet{stringArg, cutset},
			Return:   semantic.String,
		}),
		func(args values.Object) (values.Value, error) {
			var str string
			var cutsetStr string

			v, ok := args.Get(stringArg)
			if !ok {
				return nil, errMissingV
			}

			cutset, ok := args.Get(cutset)
			if !ok {
				return nil, errMissingCutset
			}

			if v.Type().Nature() == semantic.String && cutset.Type().Nature() == semantic.String {
				str = v.Str()
				cutsetStr = cutset.Str()

				str = stringFn(str, cutsetStr)
				return values.NewString(str), nil
			}

			return nil, fmt.Errorf("cannot trim argument of type %v", v.Type().Nature())
		},
		false,
	)
}

const (
	stringArg = "v"
)
const (
	cutset = "cutset"
)

var errMissingV = fmt.Errorf("missing argument %q", stringArg)
var errMissingCutset = fmt.Errorf("missing argument %q", cutset)
