package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	stddevSignature := runtime.MustLookupBuiltinType("experimental", "stddev")
	runtime.RegisterPackageValue("experimental", "stddev", flux.MustValue(flux.FunctionValue("stddev", universe.CreateStddevOpSpec, stddevSignature)))
}
