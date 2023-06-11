package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	countSignature := runtime.MustLookupBuiltinType("experimental", "count")
	runtime.RegisterPackageValue("experimental", "count", flux.MustValue(flux.FunctionValue("count", universe.CreateCountOpSpec, countSignature)))
}
