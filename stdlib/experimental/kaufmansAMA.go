package experimental

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func init() {
	kamaSignature := runtime.MustLookupBuiltinType("experimental", "kaufmansAMA")
	runtime.RegisterPackageValue("experimental", "kaufmansAMA", flux.MustValue(flux.FunctionValue("kaufmansAMA", universe.CreatekamaOpSpec, kamaSignature)))
}
