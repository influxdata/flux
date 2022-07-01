package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	integralSignature := runtime.MustLookupBuiltinType("experimental", "integral")
	runtime.RegisterPackageValue("experimental", "integral", flux.MustValue(flux.FunctionValue("integral", universe.CreateIntegralOpSpec, integralSignature)))
}
