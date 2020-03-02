package experimental

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const (
	addDurationTo        = "addDuration"
	subtractDurationFrom = "subDuration"
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
		return values.NewTime(t.Time().Add(d.Duration())), nil
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
		return values.NewTime(t.Time().Add(d.Duration().Mul(-1))), nil
	}
	return values.NewFunction(name, tp, fn, false)
}
