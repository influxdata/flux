package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	uniqueSignature := runtime.MustLookupBuiltinType("experimental", "unique")
	runtime.RegisterPackageValue("experimental", "unique", flux.MustValue(flux.FunctionValue("unique", universe.CreateUniqueOpSpec, uniqueSignature)))
}
