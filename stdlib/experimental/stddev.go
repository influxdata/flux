package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	stddevSignature := runtime.MustLookupBuiltinType("experimental", "stddev")
	runtime.RegisterPackageValue("experimental", "stddev", flux.MustValue(flux.FunctionValue("stddev", universe.CreateStddevOpSpec, stddevSignature)))
}
