package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	firstSignature := runtime.MustLookupBuiltinType("experimental", "first")
	runtime.RegisterPackageValue("experimental", "first", flux.MustValue(flux.FunctionValue("first", universe.CreateFirstOpSpec, firstSignature)))
}
