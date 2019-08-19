package slack

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/flux"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var defaultColors = map[string]struct{}{
	"good":    struct{}{},
	"warning": struct{}{},
	"danger":  struct{}{},
}

var errColorParse = errors.New("could not parse color string")

func validateColorString(color string) (string, error) {
	if _, ok := defaultColors[color]; ok {
		return color, nil
	}

	if strings.HasPrefix(color, "#") {
		hex, err := strconv.ParseInt(color[1:], 16, 64)
		if err != nil {
			return "", err
		}
		if hex < 0 || hex > 0xffffff {
			return "", errColorParse
		}
		return color[1:], nil
	}
	return "", errColorParse
}

var validateColorStringFluxFn = values.NewFunction(
	"validateColorString",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{"color": semantic.String},
		Required:   semantic.LabelSet{"color"},
		Return:     semantic.String,
	}),
	func(args values.Object) (values.Value, error) {
		v, ok := args.Get("color")

		if !ok {
			return nil, fmt.Errorf("missing argument: color")
		}

		if v.Type().Nature() == semantic.String {
			hex, err := validateColorString(v.Str())
			if err != nil {
				return nil, err
			}
			return values.NewString(hex), nil
		}

		return nil, fmt.Errorf("could not parse color string")
	},
	false,
)

//var print = values.NewFunction(
//	"print",
//	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
//		Parameters: map[string]semantic.PolyType{"color": semantic.String},
//		Required:   semantic.LabelSet{"x"},
//		Return:     semantic.String,
//	}),
//	func(args values.Object) (values.Value, error) {
//		v, ok := args.Get("color")
//
//		if !ok {
//			return nil, fmt.Errorf("missing argument: color")
//		}
//
//		if v.Type().Nature() == semantic.String {
//			hex, err := validateColorString(v.Str())
//			if err != nil {
//				return nil, err
//			}
//			return values.NewString(hex), nil
//		}
//
//		return nil, fmt.Errorf("could not parse color string")
//	},
//	false,
//)

func init() {
	flux.RegisterPackageValue("slack", "validateColorString", validateColorStringFluxFn)
}
