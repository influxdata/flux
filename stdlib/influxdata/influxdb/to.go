package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

func init() {
	toSignature := flux.FunctionSignature(
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

	flux.RegisterPackageValue("influxdata/influxdb", "to", flux.FunctionValueWithSideEffect(ToKind, createToOpSpec, toSignature))
}
func createToOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	return nil, errors.New("`to` is not implemented in native Flux")
}
