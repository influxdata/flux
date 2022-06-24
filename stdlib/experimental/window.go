package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	windowSignature := runtime.MustLookupBuiltinType("experimental", "_window")
	runtime.RegisterPackageValue("experimental", "_window", flux.MustValue(flux.FunctionValue("window", universe.CreateWindowOpSpec, windowSignature)))
}
