// From is an operation that mocks the real implementation of InfluxDB's from.
// It is used in Flux to compile queries that resemble real queries issued against InfluxDB.
// Implementors of the real from are expected to replace its implementation via flux.ReplacePackageValue.
package influxdb

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const FromKind = "from"

type FromOpSpec struct {
	Bucket string
}

func init() {
	fromSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"bucket": semantic.String,
		},
		Required: nil,
		Return:   flux.TableObjectType,
	}

	flux.RegisterPackageValue("influxdata/influxdb", FromKind, flux.FunctionValue(FromKind, createFromOpSpec, fromSignature))
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)
	if b, _, e := args.GetString("bucket"); e != nil {
		return nil, e
	} else {
		spec.Bucket = b
	}
	return spec, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	Bucket string
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Bucket: spec.Bucket,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	*ns = *s
	return ns
}
