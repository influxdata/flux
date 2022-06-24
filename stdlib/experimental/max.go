package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	maxSignature := runtime.MustLookupBuiltinType("experimental", "max")
	runtime.RegisterPackageValue("experimental", "max", flux.MustValue(flux.FunctionValue("max", universe.CreateMaxOpSpec, maxSignature)))
}
