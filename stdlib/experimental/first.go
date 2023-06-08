package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	firstSignature := runtime.MustLookupBuiltinType("experimental", "first")
	runtime.RegisterPackageValue("experimental", "first", flux.MustValue(flux.FunctionValue("first", universe.CreateFirstOpSpec, firstSignature)))
}
