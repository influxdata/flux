package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	integralSignature := runtime.MustLookupBuiltinType("experimental", "integral")
	runtime.RegisterPackageValue("experimental", "integral", flux.MustValue(flux.FunctionValue("integral", universe.CreateIntegralOpSpec, integralSignature)))
}
