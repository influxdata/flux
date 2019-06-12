package regexp

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {

	SpecialFns = map[string]values.Function{
		"compile": values.NewFunction(
			"compile",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"v": semantic.String},
				Required:   semantic.LabelSet{"v"},
				Return:     semantic.Regexp,
			}),
			func(args values.Object) (values.Value, error) {
				v, ok := args.Get("v")
				if !ok {
					return nil, errors.New("missing argument v")
				}

				if v.Type().Nature() == semantic.String {
					re, err := regexp.Compile(v.Str())
					if err != nil {
						return nil, err
					}
					return values.NewRegexp(re), err
				}
				return nil, fmt.Errorf("cannot convert argument v of type %v to Regex", v.Type().Nature())
			},
			false,
		),
	}

	flux.RegisterPackageValue("regexp", "compile", SpecialFns["compile"])
}
