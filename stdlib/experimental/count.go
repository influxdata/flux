package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	countSignature := runtime.MustLookupBuiltinType("experimental", "count")
	runtime.RegisterPackageValue("experimental", "count", flux.MustValue(flux.FunctionValue("count", universe.CreateCountOpSpec, countSignature)))
}
