package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	skewSignature := runtime.MustLookupBuiltinType("experimental", "skew")
	runtime.RegisterPackageValue("experimental", "skew", flux.MustValue(flux.FunctionValue("skew", universe.CreateSkewOpSpec, skewSignature)))
}
