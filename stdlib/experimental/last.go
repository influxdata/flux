package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	lastSignature := runtime.MustLookupBuiltinType("experimental", "last")
	runtime.RegisterPackageValue("experimental", "last", flux.MustValue(flux.FunctionValue("last", universe.CreateLastOpSpec, lastSignature)))
}
