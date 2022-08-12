package debug

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
)

const OpaqueKind = "internal/debug.opaque"

type OpaqueOpSpec struct{}

func init() {
	opaqueSig := runtime.MustLookupBuiltinType("internal/debug", "opaque")

	runtime.RegisterPackageValue("internal/debug", "opaque", flux.MustValue(flux.FunctionValue(OpaqueKind, createOpaqueOpSpec, opaqueSig)))
	// opaque uses the same procedure spec and transformation as pass, so we only need to
	// create and register the op spec here.
}

func createOpaqueOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(OpaqueOpSpec), nil
}

func (s *OpaqueOpSpec) Kind() flux.OperationKind {
	return OpaqueKind
}
