package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	kamaSignature := runtime.MustLookupBuiltinType("experimental", "kaufmansAMA")
	runtime.RegisterPackageValue("experimental", "kaufmansAMA", flux.MustValue(flux.FunctionValue("kaufmansAMA", universe.CreatekamaOpSpec, kamaSignature)))
}
