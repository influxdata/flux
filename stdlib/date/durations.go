package date

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/date"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const (
	addDurationTo        = "_add"
	subtractDurationFrom = "_sub"
)

func init() {
	runtime.RegisterPackageValue("date", addDurationTo, addDuration(addDurationTo))
	runtime.RegisterPackageValue("date", subtractDurationFrom, subDuration(subtractDurationFrom))
}

func addDuration(name string) values.Value {
	tp := runtime.MustLookupBuiltinType("date", "add")
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
