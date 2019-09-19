package date

import (
	"context"
	"fmt"
	"math"
	"time"

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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
			func(ctx context.Context, args values.Object) (values.Value, error) {
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
		"year": values.NewFunction(
			"year",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Year())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"week": values.NewFunction(
			"week",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					_, week := v1.Time().Time().ISOWeek()
					return values.NewInt(int64(week)), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"quarter": values.NewFunction(
			"quarter",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					month := v1.Time().Time().Month()
					return values.NewInt(int64(math.Ceil(float64(month) / 3.0))), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"millisecond": values.NewFunction(
			"millisecond",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					millisecond := int64(time.Nanosecond) * int64(v1.Time().Time().Nanosecond()) / int64(time.Millisecond)
					return values.NewInt(millisecond), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"microsecond": values.NewFunction(
			"microsecond",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					microsecond := int64(time.Nanosecond) * int64(v1.Time().Time().Nanosecond()) / int64(time.Microsecond)
					return values.NewInt(microsecond), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"nanosecond": values.NewFunction(
			"nanosecond",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time},
				Required:   semantic.LabelSet{"t"},
				Return:     semantic.Int,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1.Type().Nature() == semantic.Time {
					return values.NewInt(int64(v1.Time().Time().Nanosecond())), nil
				}
				return nil, fmt.Errorf("cannot convert argument t of type %v to time", v1.Type().Nature())
			}, false,
		),
		"truncate": values.NewFunction(
			"truncate",
			semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{"t": semantic.Time, "unit": semantic.Duration},
				Required:   semantic.LabelSet{"t", "unit"},
				Return:     semantic.Time,
			}),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				u, unitOk := args.Get("unit")
				if !unitOk {
					return nil, errors.New(codes.Invalid, "missing argument unit")
				}

				if v.Type().Nature() == semantic.Time && u.Type().Nature() == semantic.Duration {
					return values.NewTime(v.Time().Truncate(u.Duration())), nil
				}
				return nil, fmt.Errorf("cannot truncate argument t of type %v to unit %v", v.Type().Nature(), u)
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
	flux.RegisterPackageValue("date", "year", SpecialFns["year"])
	flux.RegisterPackageValue("date", "week", SpecialFns["week"])
	flux.RegisterPackageValue("date", "quarter", SpecialFns["quarter"])
	flux.RegisterPackageValue("date", "millisecond", SpecialFns["millisecond"])
	flux.RegisterPackageValue("date", "microsecond", SpecialFns["microsecond"])
	flux.RegisterPackageValue("date", "nanosecond", SpecialFns["nanosecond"])
	flux.RegisterPackageValue("date", "truncate", SpecialFns["truncate"])
}
