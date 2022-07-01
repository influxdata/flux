package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	meanSignature := runtime.MustLookupBuiltinType("experimental", "mean")
	runtime.RegisterPackageValue("experimental", "mean", flux.MustValue(flux.FunctionValue("mean", universe.CreateMeanOpSpec, meanSignature)))
}
