package experimental

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	addDurationTo        = "addDuration"
	subtractDurationFrom = "subDuration"
)

func init() {
	flux.RegisterPackageValue("experimental", addDurationTo, addDuration(addDurationTo))
	flux.RegisterPackageValue("experimental", subtractDurationFrom, subDuration(subtractDurationFrom))
}

func addDuration(name string) values.Value {
	tp := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"d":  semantic.Duration,
			"to": semantic.Time,
		},
		Required: semantic.LabelSet{"d", "to"},
		Return:   semantic.Time,
	})
	fn := func(args values.Object) (values.Value, error) {
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
	tp := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"d":    semantic.Duration,
			"from": semantic.Time,
		},
		Required: semantic.LabelSet{"d", "from"},
		Return:   semantic.Time,
	})
	fn := func(args values.Object) (values.Value, error) {
		d, ok := args.Get("d")
		if !ok {
			return nil, fmt.Errorf("%s requires 'd' parameter", name)
		}
		t, ok := args.Get("from")
		if !ok {
			return nil, fmt.Errorf("%s requires 'from' parameter", name)
		}
		return values.NewTime(t.Time().Add(-d.Duration())), nil
	}
	return values.NewFunction(name, tp, fn, false)
}
