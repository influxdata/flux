package experimental

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func init() {
	meanSignature := runtime.MustLookupBuiltinType("experimental", "mean")
	runtime.RegisterPackageValue("experimental", "mean", flux.MustValue(flux.FunctionValue("mean", universe.CreateMeanOpSpec, meanSignature)))
}
