package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	fillSignature := runtime.MustLookupBuiltinType("experimental", "fill")
	runtime.RegisterPackageValue("experimental", "fill", flux.MustValue(flux.FunctionValue("fill", universe.CreateFillOpSpec, fillSignature)))
}
