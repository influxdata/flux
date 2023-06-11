package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	meanSignature := runtime.MustLookupBuiltinType("experimental", "mean")
	runtime.RegisterPackageValue("experimental", "mean", flux.MustValue(flux.FunctionValue("mean", universe.CreateMeanOpSpec, meanSignature)))
}
