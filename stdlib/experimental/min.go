package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	minSignature := runtime.MustLookupBuiltinType("experimental", "min")
	runtime.RegisterPackageValue("experimental", "min", flux.MustValue(flux.FunctionValue("min", universe.CreateMinOpSpec, minSignature)))
}
