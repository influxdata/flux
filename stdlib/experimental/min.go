package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	minSignature := runtime.MustLookupBuiltinType("experimental", "min")
	runtime.RegisterPackageValue("experimental", "min", flux.MustValue(flux.FunctionValue("min", universe.CreateMinOpSpec, minSignature)))
}
