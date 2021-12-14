package experimental

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/codes"
<<<<<<< HEAD
	"github.com/influxdata/flux/execute"
=======
	"github.com/influxdata/flux/internal/date"
>>>>>>> a170c552 (feat(date): incorporated review comments)
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	addDurationTo        = "_addDuration"
	subtractDurationFrom = "_subDuration"
)

func init() {
	runtime.RegisterPackageValue("experimental", addDurationTo, addDuration(addDurationTo))
	runtime.RegisterPackageValue("experimental", subtractDurationFrom, subDuration(subtractDurationFrom))
}

func addDuration(name string) values.Value {
	tp := runtime.MustLookupBuiltinType("experimental", "addDuration")
	fn := func(ctx context.Context, args values.Object) (values.Value, error) {
		d, ok := args.Get("d")
		if !ok {
			return nil, fmt.Errorf("%s requires 'd' parameter", name)
		}
		t, ok := args.Get("to")
		if !ok {
			return nil, fmt.Errorf("%s requires 'to' parameter", name)
		}
<<<<<<< HEAD
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		return values.NewTime(time.Add(d.Duration())), nil
		location, err := getLocation(args)
=======
		location, offset, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime, err := date.GetTimeInLocation(t.Time(), location, offset)
>>>>>>> a170c552 (feat(date): incorporated review comments)
		if err != nil {
			return nil, err
		}
		return values.NewTime(lTime.Time().Add(d.Duration())), nil
	}
	return values.NewFunction(name, tp, fn, false)
}

func subDuration(name string) values.Value {
	tp := runtime.MustLookupBuiltinType("experimental", "subDuration")
	fn := func(ctx context.Context, args values.Object) (values.Value, error) {
		d, ok := args.Get("d")
		if !ok {
			return nil, fmt.Errorf("%s requires 'd' parameter", name)
		}
		t, ok := args.Get("from")
		if !ok {
			return nil, fmt.Errorf("%s requires 'from' parameter", name)
		}
<<<<<<< HEAD
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		return values.NewTime(time.Add(d.Duration().Mul(-1))), nil
		location, err := getLocation(args)
=======
		location, offset, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime, err := date.GetTimeInLocation(t.Time(), location, offset)
>>>>>>> a170c552 (feat(date): incorporated review comments)
		if err != nil {
			return nil, err
		}
		return values.NewTime(lTime.Time().Add(d.Duration().Mul(-1))), nil
	}
	return values.NewFunction(name, tp, fn, false)
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
