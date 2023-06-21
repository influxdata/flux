package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	minSignature := runtime.MustLookupBuiltinType("experimental", "min")
	runtime.RegisterPackageValue("experimental", "min", flux.MustValue(flux.FunctionValue("min", universe.CreateMinOpSpec, minSignature)))
}
