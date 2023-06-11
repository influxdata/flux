package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	histogramQuantileSignature := runtime.MustLookupBuiltinType("experimental", "histogramQuantile")
	runtime.RegisterPackageValue("experimental", "histogramQuantile", flux.MustValue(flux.FunctionValue("histogramQuantile", universe.CreateHistogramQuantileOpSpec, histogramQuantileSignature)))
}
