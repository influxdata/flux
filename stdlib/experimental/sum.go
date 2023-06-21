package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	sumSignature := runtime.MustLookupBuiltinType("experimental", "sum")
	runtime.RegisterPackageValue("experimental", "sum", flux.MustValue(flux.FunctionValue("sum", universe.CreateSumOpSpec, sumSignature)))
}
