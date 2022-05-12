package date

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/date"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	addDurationTo        = "_add"
	subtractDurationFrom = "_sub"
)

func init() {
	runtime.RegisterPackageValue("date", addDurationTo, addDuration(addDurationTo))
	runtime.RegisterPackageValue("date", subtractDurationFrom, subDuration(subtractDurationFrom))
	runtime.RegisterPackageValue("date", "scale", scaleDuration())
}

func addDuration(name string) values.Value {
	tp := runtime.MustLookupBuiltinType("date", "add")
	fn := func(ctx context.Context, args values.Object) (values.Value, error) {
		d, ok := args.Get("d")
		if !ok {
			return nil, errors.Newf(codes.Invalid, "%s requires 'd' parameter", name)
		}
		t, ok := args.Get("to")
		if !ok {
			return nil, errors.Newf(codes.Invalid, "%s requires 'to' parameter", name)
		}
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		location, offset, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime, err := date.GetTimeInLocation(time, location, offset)
		if err != nil {
			return nil, err
		}
		return values.NewTime(lTime.Time().Add(d.Duration())), nil
	}
	return values.NewFunction(name, tp, fn, false)
}

func subDuration(name string) values.Value {
	tp := runtime.MustLookupBuiltinType("date", "sub")
	fn := func(ctx context.Context, args values.Object) (values.Value, error) {
		d, ok := args.Get("d")
		if !ok {
			return nil, fmt.Errorf("%s requires 'd' parameter", name)
		}
		t, ok := args.Get("from")
		if !ok {
			return nil, fmt.Errorf("%s requires 'from' parameter", name)
		}
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		location, offset, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime, err := date.GetTimeInLocation(time, location, offset)
		if err != nil {
			return nil, err
		}
		return values.NewTime(lTime.Time().Add(d.Duration().Mul(-1))), nil
	}
	return values.NewFunction(name, tp, fn, false)
}

func scaleDuration() values.Value {
	tp := runtime.MustLookupBuiltinType("date", "scale")
	return values.NewFunction("scale", tp, func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(func(args interpreter.Arguments) (values.Value, error) {
			d, err := args.GetRequired("d")
			if err != nil {
				return nil, err
			} else if d.Type().Nature() != semantic.Duration {
				return nil, errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", "scale", semantic.Duration, d.Type().Nature())
			}

			n, err := args.GetRequiredInt("n")
			if err != nil {
				return nil, err
			}

			if n == 1 {
				return d, nil
			}
			return values.NewDuration(d.Duration().Mul(int(n))), nil
		}, args)
	}, false)
}
