package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	histogramQuantileSignature := runtime.MustLookupBuiltinType("experimental", "histogramQuantile")
	runtime.RegisterPackageValue("experimental", "histogramQuantile", flux.MustValue(flux.FunctionValue("histogramQuantile", universe.CreateHistogramQuantileOpSpec, histogramQuantileSignature)))
}
