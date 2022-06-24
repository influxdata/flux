package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	histogramSignature := runtime.MustLookupBuiltinType("experimental", "histogram")
	runtime.RegisterPackageValue("experimental", "histogram", flux.MustValue(flux.FunctionValue("histogram", universe.CreateHistogramOpSpec, histogramSignature)))
}
