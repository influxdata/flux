package experimental

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/locationutil"
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
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		return values.NewTime(time.Add(d.Duration())), nil
		location, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime := location.ToUTCTime(t.Time())
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
		deps := execute.GetExecutionDependencies(ctx)
		time, err := deps.ResolveTimeable(t)
		if err != nil {
			return nil, err
		}
		return values.NewTime(time.Add(d.Duration().Mul(-1))), nil
		location, err := getLocation(args)
		if err != nil {
			return nil, err
		}
		lTime := location.ToUTCTime(t.Time())
		return values.NewTime(lTime.Time().Add(d.Duration().Mul(-1))), nil
	}
	return values.NewFunction(name, tp, fn, false)
}

func getLocation(args values.Object) (locationutil.Location, error) {
	var name, offset values.Value
	var ok bool
	a := interpreter.NewArguments(args)
	if location, err := a.GetRequiredObject("location"); err != nil {
		return locationutil.Location{}, err
	} else {
		name, ok = location.Get("zone")
		if !ok {
			return locationutil.Location{}, errors.New(codes.Invalid, "zone property missing from location record")
		} else if got := name.Type().Nature(); got != semantic.String {
			return locationutil.Location{}, errors.Newf(codes.Invalid, "zone property for location must be of type %s, got %s", semantic.String, got)
		}

		if offset, ok = location.Get("offset"); ok {
			if got := offset.Type().Nature(); got != semantic.Duration {
				return locationutil.Location{}, errors.Newf(codes.Invalid, "offset property for location must be of type %s, got %s", semantic.Duration, got)
			}
		}
	}
	if name.IsNull() {
		name = values.NewString("UTC")
	}
	loc := locationutil.Location{
		Name:   name.Str(),
		Offset: offset.Duration(),
	}
	return loc, nil
}
