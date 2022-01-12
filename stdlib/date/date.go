package date

import (
	"context"
	"math"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/date"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {
	SpecialFns = map[string]values.Function{
		"second": values.NewFunction(
			"second",
			runtime.MustLookupBuiltinType("date", "second"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(tm.Time().Second())), nil
			}, false,
		),
		"minute": values.NewFunction(
			"minute",
			runtime.MustLookupBuiltinType("date", "_minute"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Minute())), nil
			}, false,
		),
		"hour": values.NewFunction(
			"hour",
			runtime.MustLookupBuiltinType("date", "_hour"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Hour())), nil
			}, false,
		),
		"weekDay": values.NewFunction(
			"weekDay",
			runtime.MustLookupBuiltinType("date", "_weekDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Weekday())), nil
			}, false,
		),
		"monthDay": values.NewFunction(
			"monthDay",
			runtime.MustLookupBuiltinType("date", "_monthDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Day())), nil
			}, false,
		),
		"yearDay": values.NewFunction(
			"yearDay",
			runtime.MustLookupBuiltinType("date", "_yearDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().YearDay())), nil
			}, false,
		),
		"month": values.NewFunction(
			"month",
			runtime.MustLookupBuiltinType("date", "_month"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Month())), nil
			}, false,
		),
		"year": values.NewFunction(
			"year",
			runtime.MustLookupBuiltinType("date", "_year"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(lTime.Time().Time().Year())), nil
			}, false,
		),
		"week": values.NewFunction(
			"week",
			runtime.MustLookupBuiltinType("date", "_week"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				_, week := lTime.Time().Time().ISOWeek()
				return values.NewInt(int64(week)), nil
			}, false,
		),
		"quarter": values.NewFunction(
			"quarter",
			runtime.MustLookupBuiltinType("date", "_quarter"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := getLocation(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				month := lTime.Time().Time().Month()
				return values.NewInt(int64(math.Ceil(float64(month) / 3.0))), nil
			}, false,
		),
		"millisecond": values.NewFunction(
			"millisecond",
			runtime.MustLookupBuiltinType("date", "millisecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				millisecond := int64(time.Nanosecond) * int64(tm.Time().Nanosecond()) / int64(time.Millisecond)
				return values.NewInt(millisecond), nil
			}, false,
		),
		"microsecond": values.NewFunction(
			"microsecond",
			runtime.MustLookupBuiltinType("date", "microsecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				microsecond := int64(time.Nanosecond) * int64(tm.Time().Nanosecond()) / int64(time.Microsecond)
				return values.NewInt(microsecond), nil
			}, false,
		),
		"nanosecond": values.NewFunction(
			"nanosecond",
			runtime.MustLookupBuiltinType("date", "nanosecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(tm.Time().Nanosecond())), nil
			}, false,
		),
		"truncate": values.NewFunction(
			"truncate",
			runtime.MustLookupBuiltinType("date", "truncate"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				u, unitOk := args.Get("unit")
				if !unitOk {
					return nil, errors.New(codes.Invalid, "missing argument unit")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v)
				if err != nil {
					return nil, err
				}
				w, err := execute.NewWindow(u.Duration(), u.Duration(), execute.Duration{})
				if err != nil {
					return nil, err
				}
				b := w.GetEarliestBounds(t)
				return values.NewTime(b.Start), nil
			}, false,
		),
	}

	runtime.RegisterPackageValue("date", "second", SpecialFns["second"])
	runtime.RegisterPackageValue("date", "_minute", SpecialFns["minute"])
	runtime.RegisterPackageValue("date", "_hour", SpecialFns["hour"])
	runtime.RegisterPackageValue("date", "_weekDay", SpecialFns["weekDay"])
	runtime.RegisterPackageValue("date", "_monthDay", SpecialFns["monthDay"])
	runtime.RegisterPackageValue("date", "_yearDay", SpecialFns["yearDay"])
	runtime.RegisterPackageValue("date", "_month", SpecialFns["month"])
	runtime.RegisterPackageValue("date", "_year", SpecialFns["year"])
	runtime.RegisterPackageValue("date", "_week", SpecialFns["week"])
	runtime.RegisterPackageValue("date", "_quarter", SpecialFns["quarter"])
	runtime.RegisterPackageValue("date", "millisecond", SpecialFns["millisecond"])
	runtime.RegisterPackageValue("date", "microsecond", SpecialFns["microsecond"])
	runtime.RegisterPackageValue("date", "nanosecond", SpecialFns["nanosecond"])
	runtime.RegisterPackageValue("date", "truncate", SpecialFns["truncate"])
}

func getTime(args values.Object) (values.Value, error) {
	tArg, ok := args.Get("t")
	if !ok {
		return nil, errors.New(codes.Invalid, "missing argument t")
	}
	if tArg == nil {
		return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
	}
	return tArg, nil
}

func getLocation(args values.Object) (string, values.Duration, error) {
	var name, offset values.Value
	var ok bool
	a := interpreter.NewArguments(args)
	if location, err := a.GetRequiredObject("location"); err != nil {
		return "UTC", values.ConvertDurationNsecs(0), err
	} else {
		name, ok = location.Get("zone")
		if !ok {
			return "UTC", values.ConvertDurationNsecs(0), errors.New(codes.Invalid, "zone property missing from location record")
		} else if got := name.Type().Nature(); got != semantic.String {
			return "UTC", values.ConvertDurationNsecs(0), errors.Newf(codes.Invalid, "zone property for location must be of type %s, got %s", semantic.String, got)
		}

		if offset, ok = location.Get("offset"); ok {
			if got := offset.Type().Nature(); got != semantic.Duration {
				return "UTC", values.ConvertDurationNsecs(0), errors.Newf(codes.Invalid, "offset property for location must be of type %s, got %s", semantic.Duration, got)
			}
		}
	}
	if name.IsNull() {
		name = values.NewString("UTC")
	}

	return name.Str(), offset.Duration(), nil
}

func getTimeableTime(ctx context.Context, args values.Object) (values.Time, error) {
	var tm values.Time
	t, err := getTime(args)
	if err != nil {
		return tm, err
	}
	deps := execute.GetExecutionDependencies(ctx)
	tm, err = deps.ResolveTimeable(t)
	if err != nil {
		return tm, err
	}
	return tm, nil
}
