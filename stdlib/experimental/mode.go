package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	modeSignature := runtime.MustLookupBuiltinType("experimental", "mode")
	runtime.RegisterPackageValue("experimental", "mode", flux.MustValue(flux.FunctionValue("mode", universe.CreateModeOpSpec, modeSignature)))
}
