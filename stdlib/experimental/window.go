package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	windowSignature := runtime.MustLookupBuiltinType("experimental", "_window")
	runtime.RegisterPackageValue("experimental", "_window", flux.MustValue(flux.FunctionValue("window", universe.CreateWindowOpSpec, windowSignature)))
}
