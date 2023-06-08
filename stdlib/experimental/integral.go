package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	integralSignature := runtime.MustLookupBuiltinType("experimental", "integral")
	runtime.RegisterPackageValue("experimental", "integral", flux.MustValue(flux.FunctionValue("integral", universe.CreateIntegralOpSpec, integralSignature)))
}
