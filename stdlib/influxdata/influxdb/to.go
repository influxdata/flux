package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

var ToSignature = semantic.MustLookupBuiltinType("influxdata/influxdb", "to")

func init() {
	runtime.RegisterPackageValue("influxdata/influxdb", ToKind, flux.MustValue(flux.FunctionValueWithSideEffect(ToKind, nil, ToSignature)))
}
