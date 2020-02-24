package date

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {
	SpecialFns = map[string]values.Function{
		"second": values.NewFunction(
			"second",
			semantic.MustLookupBuiltinType("date", "second"),
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
			semantic.MustLookupBuiltinType("date", "minute"),
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
			semantic.MustLookupBuiltinType("date", "hour"),
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
			semantic.MustLookupBuiltinType("date", "weekDay"),
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
			semantic.MustLookupBuiltinType("date", "monthDay"),
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
			semantic.MustLookupBuiltinType("date", "yearDay"),
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
			semantic.MustLookupBuiltinType("date", "month"),
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
			semantic.MustLookupBuiltinType("date", "year"),
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
			semantic.MustLookupBuiltinType("date", "week"),
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
			semantic.MustLookupBuiltinType("date", "quarter"),
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
			semantic.MustLookupBuiltinType("date", "millisecond"),
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
			semantic.MustLookupBuiltinType("date", "microsecond"),
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
			semantic.MustLookupBuiltinType("date", "nanosecond"),
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
			semantic.MustLookupBuiltinType("date", "truncate"),
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
					w, err := execute.NewWindow(u.Duration(), u.Duration(), execute.Duration{})
					if err != nil {
						return nil, err
					}
					b := w.GetEarliestBounds(v.Time())
					return values.NewTime(b.Start), nil
				}
				return nil, fmt.Errorf("cannot truncate argument t of type %v to unit %v", v.Type().Nature(), u)
			}, false,
		),
	}

	runtime.RegisterPackageValue("date", "second", SpecialFns["second"])
	runtime.RegisterPackageValue("date", "minute", SpecialFns["minute"])
	runtime.RegisterPackageValue("date", "hour", SpecialFns["hour"])
	runtime.RegisterPackageValue("date", "weekDay", SpecialFns["weekDay"])
	runtime.RegisterPackageValue("date", "monthDay", SpecialFns["monthDay"])
	runtime.RegisterPackageValue("date", "yearDay", SpecialFns["yearDay"])
	runtime.RegisterPackageValue("date", "month", SpecialFns["month"])
	runtime.RegisterPackageValue("date", "year", SpecialFns["year"])
	runtime.RegisterPackageValue("date", "week", SpecialFns["week"])
	runtime.RegisterPackageValue("date", "quarter", SpecialFns["quarter"])
	runtime.RegisterPackageValue("date", "millisecond", SpecialFns["millisecond"])
	runtime.RegisterPackageValue("date", "microsecond", SpecialFns["microsecond"])
	runtime.RegisterPackageValue("date", "nanosecond", SpecialFns["nanosecond"])
	runtime.RegisterPackageValue("date", "truncate", SpecialFns["truncate"])
}
