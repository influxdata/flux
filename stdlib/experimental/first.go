package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	firstSignature := runtime.MustLookupBuiltinType("experimental", "first")
	runtime.RegisterPackageValue("experimental", "first", flux.MustValue(flux.FunctionValue("first", universe.CreateFirstOpSpec, firstSignature)))
}
