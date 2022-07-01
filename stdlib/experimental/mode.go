package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	modeSignature := runtime.MustLookupBuiltinType("experimental", "mode")
	runtime.RegisterPackageValue("experimental", "mode", flux.MustValue(flux.FunctionValue("mode", universe.CreateModeOpSpec, modeSignature)))
}
