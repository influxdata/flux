package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	skewSignature := runtime.MustLookupBuiltinType("experimental", "skew")
	runtime.RegisterPackageValue("experimental", "skew", flux.MustValue(flux.FunctionValue("skew", universe.CreateSkewOpSpec, skewSignature)))
}
