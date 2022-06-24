package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	histogramQuantileSignature := runtime.MustLookupBuiltinType("experimental", "histogramQuantile")
	runtime.RegisterPackageValue("experimental", "histogramQuantile", flux.MustValue(flux.FunctionValue("histogramQuantile", universe.CreateHistogramQuantileOpSpec, histogramQuantileSignature)))
}
