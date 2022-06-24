package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	modeSignature := runtime.MustLookupBuiltinType("experimental", "mode")
	runtime.RegisterPackageValue("experimental", "mode", flux.MustValue(flux.FunctionValue("mode", universe.CreateModeOpSpec, modeSignature)))
}
