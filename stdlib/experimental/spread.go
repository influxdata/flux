package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	spreadSignature := runtime.MustLookupBuiltinType("experimental", "spread")
	runtime.RegisterPackageValue("experimental", "spread", flux.MustValue(flux.FunctionValue("spread", universe.CreateSpreadOpSpec, spreadSignature)))
}
