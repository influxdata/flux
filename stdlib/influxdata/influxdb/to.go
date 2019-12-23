package influxdb

import (
	"github.com/influxdata/flux"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

var ToSignature = flux.LookupBuiltInType("influxdata/influxdb", "to")

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", ToKind, flux.MustValue(flux.FunctionValueWithSideEffect(ToKind, nil, ToSignature)))
}
