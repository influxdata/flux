package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	sumSignature := runtime.MustLookupBuiltinType("experimental", "sum")
	runtime.RegisterPackageValue("experimental", "sum", flux.MustValue(flux.FunctionValue("sum", universe.CreateSumOpSpec, sumSignature)))
}
