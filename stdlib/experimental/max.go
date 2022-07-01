package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	maxSignature := runtime.MustLookupBuiltinType("experimental", "max")
	runtime.RegisterPackageValue("experimental", "max", flux.MustValue(flux.FunctionValue("max", universe.CreateMaxOpSpec, maxSignature)))
}
