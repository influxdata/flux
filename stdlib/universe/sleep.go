package universe

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("universe", "sleep", sleep)
}

const (
	vArg        = "v"
	durationArg = "duration"
)

var sleep = values.NewFunction(
	"sleep",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			vArg:        semantic.Tvar(1),
			durationArg: semantic.Duration,
		},
		PipeArgument: vArg,
		Required:     semantic.LabelSet{vArg, durationArg},
		Return:       semantic.Tvar(1),
	}),
	func(args values.Object) (values.Value, error) {
		v, ok := args.Get(vArg)
		if !ok {
			return nil, fmt.Errorf("missing argument %q", vArg)
		}
		d, ok := args.Get(durationArg)
		if !ok {
			return nil, fmt.Errorf("missing argument %q", durationArg)
		}

		if d.Type().Nature() == semantic.Duration {
			dur := d.Duration()
			time.Sleep(time.Duration(dur))
		}
		return v, nil
	},
	// sleeping is a side effect
	true,
)
