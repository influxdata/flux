package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	quantileSignature := runtime.MustLookupBuiltinType("experimental", "quantile")
	runtime.RegisterPackageValue("experimental", "quantile", flux.MustValue(flux.FunctionValue("quantile", universe.CreateQuantileOpSpec, quantileSignature)))
}
