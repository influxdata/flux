package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	skewSignature := runtime.MustLookupBuiltinType("experimental", "skew")
	runtime.RegisterPackageValue("experimental", "skew", flux.MustValue(flux.FunctionValue("skew", universe.CreateSkewOpSpec, skewSignature)))
}
