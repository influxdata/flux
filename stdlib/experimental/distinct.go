package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	distinctSignature := runtime.MustLookupBuiltinType("experimental", "distinct")
	runtime.RegisterPackageValue("experimental", "distinct", flux.MustValue(flux.FunctionValue("distinct", universe.CreateDistinctOpSpec, distinctSignature)))
}
