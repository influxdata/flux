package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	fillSignature := runtime.MustLookupBuiltinType("experimental", "fill")
	runtime.RegisterPackageValue("experimental", "fill", flux.MustValue(flux.FunctionValue("fill", universe.CreateFillOpSpec, fillSignature)))
}
