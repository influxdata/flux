package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	spreadSignature := runtime.MustLookupBuiltinType("experimental", "spread")
	runtime.RegisterPackageValue("experimental", "spread", flux.MustValue(flux.FunctionValue("spread", universe.CreateSpreadOpSpec, spreadSignature)))
}
