package date

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {
	SpecialFns = map[string]values.Function{
		"second": values.NewFunction(
			"second",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Second())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"minute": values.NewFunction(
			"minute",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Minute())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"hour": values.NewFunction(
			"hour",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Hour())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"weekDay": values.NewFunction(
			"weekDay",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Weekday())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"monthDay": values.NewFunction(
			"monthDay",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Day())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"yearDay": values.NewFunction(
			"yearDay",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().YearDay())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"month": values.NewFunction(
			"month",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Month())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
	}

	flux.RegisterPackageValue("date", "second", SpecialFns["second"])
	flux.RegisterPackageValue("date", "minute", SpecialFns["minute"])
	flux.RegisterPackageValue("date", "hour", SpecialFns["hour"])
	flux.RegisterPackageValue("date", "weekDay", SpecialFns["weekDay"])
	flux.RegisterPackageValue("date", "monthDay", SpecialFns["monthDay"])
	flux.RegisterPackageValue("date", "yearDay", SpecialFns["yearDay"])
	flux.RegisterPackageValue("date", "month", SpecialFns["month"])
}
