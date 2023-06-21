package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	windowSignature := runtime.MustLookupBuiltinType("experimental", "_window")
	runtime.RegisterPackageValue("experimental", "_window", flux.MustValue(flux.FunctionValue("window", universe.CreateWindowOpSpec, windowSignature)))
}
