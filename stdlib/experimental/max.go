package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	maxSignature := runtime.MustLookupBuiltinType("experimental", "max")
	runtime.RegisterPackageValue("experimental", "max", flux.MustValue(flux.FunctionValue("max", universe.CreateMaxOpSpec, maxSignature)))
}
