package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	lastSignature := runtime.MustLookupBuiltinType("experimental", "last")
	runtime.RegisterPackageValue("experimental", "last", flux.MustValue(flux.FunctionValue("last", universe.CreateLastOpSpec, lastSignature)))
}
