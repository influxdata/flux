package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	stddevSignature := runtime.MustLookupBuiltinType("experimental", "stddev")
	runtime.RegisterPackageValue("experimental", "stddev", flux.MustValue(flux.FunctionValue("stddev", universe.CreateStddevOpSpec, stddevSignature)))
}
