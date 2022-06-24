package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	spreadSignature := runtime.MustLookupBuiltinType("experimental", "spread")
	runtime.RegisterPackageValue("experimental", "spread", flux.MustValue(flux.FunctionValue("spread", universe.CreateSpreadOpSpec, spreadSignature)))
}
