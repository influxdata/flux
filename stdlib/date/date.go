package date

import (
	"context"
	"math"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/date"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {
	SpecialFns = map[string]values.Function{
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
				location, offset, err := date.GetLocationFromObjArgs(args)
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
				location, offset, err := date.GetLocationFromObjArgs(args)
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
				location, offset, err := date.GetLocationFromObjArgs(args)
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
				location, offset, err := date.GetLocationFromObjArgs(args)
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
				location, offset, err := date.GetLocationFromObjArgs(args)
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
		"week": values.NewFunction(
			"week",
			runtime.MustLookupBuiltinType("date", "_week"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := date.GetLocationFromObjArgs(args)
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
		"month": values.NewFunction(
			"month",
			runtime.MustLookupBuiltinType("date", "_month"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := date.GetLocationFromObjArgs(args)
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
		"quarter": values.NewFunction(
			"quarter",
			runtime.MustLookupBuiltinType("date", "_quarter"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := date.GetLocationFromObjArgs(args)
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
		"year": values.NewFunction(
			"year",
			runtime.MustLookupBuiltinType("date", "_year"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := date.GetLocationFromObjArgs(args)
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
		"truncate": values.NewFunction(
			"truncate",
			runtime.MustLookupBuiltinType("date", "_truncate"),
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
				location, offset, err := date.GetLocationFromObjArgs(args)
				if err != nil {
					return nil, err
				}
				intervalLocation, err := interval.LoadLocation(location)
				// TODO: offset previously ignored. Is this enough of a fix?
				//  Need to check to see if setting it here is a sufficient fix. Follow-up in #5013
				intervalLocation.Offset = offset
				if err != nil {
					return nil, err
				}
				w, err := interval.NewWindowInLocation(u.Duration(), u.Duration(), values.Duration{}, intervalLocation)
				if err != nil {
					return nil, err
				}
				b := w.GetLatestBounds(t)
				return values.NewTime(b.Start()), nil
			}, false,
		),
		"time": values.NewFunction(
			"time",
			runtime.MustLookupBuiltinType("date", "_time"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				tm, err := getTimeableTime(ctx, args)
				if err != nil {
					return nil, err
				}
				location, offset, err := date.GetLocationFromObjArgs(args)
				if err != nil {
					return nil, err
				}
				lTime, err := date.GetTimeInLocation(tm, location, offset)
				if err != nil {
					return nil, err
				}
				return values.NewTime(lTime.Time()), nil
			}, false,
		),
	}

	runtime.RegisterPackageValue("date", "nanosecond", SpecialFns["nanosecond"])
	runtime.RegisterPackageValue("date", "microsecond", SpecialFns["microsecond"])
	runtime.RegisterPackageValue("date", "millisecond", SpecialFns["millisecond"])
	runtime.RegisterPackageValue("date", "second", SpecialFns["second"])
	runtime.RegisterPackageValue("date", "_minute", SpecialFns["minute"])
	runtime.RegisterPackageValue("date", "_hour", SpecialFns["hour"])
	runtime.RegisterPackageValue("date", "_weekDay", SpecialFns["weekDay"])
	runtime.RegisterPackageValue("date", "_week", SpecialFns["week"])
	runtime.RegisterPackageValue("date", "_monthDay", SpecialFns["monthDay"])
	runtime.RegisterPackageValue("date", "_yearDay", SpecialFns["yearDay"])
	runtime.RegisterPackageValue("date", "_month", SpecialFns["month"])
	runtime.RegisterPackageValue("date", "_quarter", SpecialFns["quarter"])
	runtime.RegisterPackageValue("date", "_year", SpecialFns["year"])
	runtime.RegisterPackageValue("date", "_truncate", SpecialFns["truncate"])
	runtime.RegisterPackageValue("date", "_time", SpecialFns["time"])

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
