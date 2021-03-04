package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	quantileSignature := runtime.MustLookupBuiltinType("experimental", "quantile")
	runtime.RegisterPackageValue("experimental", "quantile", flux.MustValue(flux.FunctionValue("quantile", universe.CreateQuantileOpSpec, quantileSignature)))
}
