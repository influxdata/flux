package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	histogramSignature := runtime.MustLookupBuiltinType("experimental", "histogram")
	runtime.RegisterPackageValue("experimental", "histogram", flux.MustValue(flux.FunctionValue("histogram", universe.CreateHistogramOpSpec, histogramSignature)))
}
