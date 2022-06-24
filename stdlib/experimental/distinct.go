package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	distinctSignature := runtime.MustLookupBuiltinType("experimental", "distinct")
	runtime.RegisterPackageValue("experimental", "distinct", flux.MustValue(flux.FunctionValue("distinct", universe.CreateDistinctOpSpec, distinctSignature)))
}
