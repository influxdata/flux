package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	countSignature := runtime.MustLookupBuiltinType("experimental", "count")
	runtime.RegisterPackageValue("experimental", "count", flux.MustValue(flux.FunctionValue("count", universe.CreateCountOpSpec, countSignature)))
}
