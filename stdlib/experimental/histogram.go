package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	histogramSignature := runtime.MustLookupBuiltinType("experimental", "histogram")
	runtime.RegisterPackageValue("experimental", "histogram", flux.MustValue(flux.FunctionValue("histogram", universe.CreateHistogramOpSpec, histogramSignature)))
}
