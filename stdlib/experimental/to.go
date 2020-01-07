package experimental

import (
	"github.com/influxdata/flux"
)

// ToKind is the kind for the experimental `to` flux function
const ExperimentalToKind = "experimental-to"

var ToSignature = semantic.LookupBuiltInType("experimental", "to")

func init() {
	flux.RegisterPackageValue("experimental", "to", flux.MustValue(flux.FunctionValueWithSideEffect("to", nil, ToSignature)))
}
