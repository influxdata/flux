package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	windowSignature := runtime.MustLookupBuiltinType("experimental", "window")
	runtime.RegisterPackageValue("experimental", "window", flux.MustValue(flux.FunctionValue("window", universe.CreateWindowOpSpec, windowSignature)))
}
