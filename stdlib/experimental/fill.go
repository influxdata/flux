package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	fillSignature := runtime.MustLookupBuiltinType("experimental", "fill")
	runtime.RegisterPackageValue("experimental", "fill", flux.MustValue(flux.FunctionValue("fill", universe.CreateFillOpSpec, fillSignature)))
}
