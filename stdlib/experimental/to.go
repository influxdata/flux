package experimental

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

// ToKind is the kind for the experimental `to` flux function
const ExperimentalToKind = "experimental-to"

var ToSignature = semantic.MustLookupBuiltinType("experimental", "to")

func init() {
	runtime.RegisterPackageValue("experimental", "to", flux.MustValue(flux.FunctionValueWithSideEffect("to", nil, ToSignature)))
}
