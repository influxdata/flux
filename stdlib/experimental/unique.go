package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	uniqueSignature := runtime.MustLookupBuiltinType("experimental", "unique")
	runtime.RegisterPackageValue("experimental", "unique", flux.MustValue(flux.FunctionValue("unique", universe.CreateUniqueOpSpec, uniqueSignature)))
}
