package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	quantileSignature := runtime.MustLookupBuiltinType("experimental", "quantile")
	runtime.RegisterPackageValue("experimental", "quantile", flux.MustValue(flux.FunctionValue("quantile", universe.CreateQuantileOpSpec, quantileSignature)))
}
