package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

var ToSignature = flux.FunctionSignature(
	map[string]semantic.PolyType{
		"bucket":            semantic.String,
		"bucketID":          semantic.String,
		"org":               semantic.String,
		"orgID":             semantic.String,
		"host":              semantic.String,
		"token":             semantic.String,
		"timeColumn":        semantic.String,
		"measurementColumn": semantic.String,
		"tagColumns":        semantic.Array,
		"fieldFn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"r": semantic.Tvar(1),
			},
			Required: semantic.LabelSet{"r"},
			Return:   semantic.Tvar(2),
		}),
	},
	[]string{},
)

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", ToKind, flux.FunctionValueWithSideEffect(ToKind, nil, ToSignature))
}
