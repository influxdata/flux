package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	uniqueSignature := runtime.MustLookupBuiltinType("experimental", "unique")
	runtime.RegisterPackageValue("experimental", "unique", flux.MustValue(flux.FunctionValue("unique", universe.CreateUniqueOpSpec, uniqueSignature)))
}
