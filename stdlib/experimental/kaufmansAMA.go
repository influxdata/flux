package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	kamaSignature := runtime.MustLookupBuiltinType("experimental", "kaufmansAMA")
	runtime.RegisterPackageValue("experimental", "kaufmansAMA", flux.MustValue(flux.FunctionValue("kaufmansAMA", universe.CreatekamaOpSpec, kamaSignature)))
}
