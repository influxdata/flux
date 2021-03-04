package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	sumSignature := runtime.MustLookupBuiltinType("experimental", "sum")
	runtime.RegisterPackageValue("experimental", "sum", flux.MustValue(flux.FunctionValue("sum", universe.CreateSumOpSpec, sumSignature)))
}
