package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	lastSignature := runtime.MustLookupBuiltinType("experimental", "last")
	runtime.RegisterPackageValue("experimental", "last", flux.MustValue(flux.FunctionValue("last", universe.CreateLastOpSpec, lastSignature)))
}
