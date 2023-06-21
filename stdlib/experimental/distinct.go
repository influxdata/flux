package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	distinctSignature := runtime.MustLookupBuiltinType("experimental", "distinct")
	runtime.RegisterPackageValue("experimental", "distinct", flux.MustValue(flux.FunctionValue("distinct", universe.CreateDistinctOpSpec, distinctSignature)))
}
