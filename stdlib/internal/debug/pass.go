package debug

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const PassKind = "internal/debug.pass"

type PassOpSpec struct{}

func init() {
	passSignature := runtime.MustLookupBuiltinType("internal/debug", "pass")

	runtime.RegisterPackageValue("internal/debug", "pass", flux.MustValue(flux.FunctionValue(PassKind, createPassOpSpec, passSignature)))
	// both pass and opaque use the procedure spec and transformation; only their type signatures differ.
	plan.RegisterProcedureSpec(PassKind, newPassProcedure, PassKind, OpaqueKind)
	execute.RegisterTransformation(PassKind, createPassTransformation)
}

func createPassOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(PassOpSpec), nil
}

func (s *PassOpSpec) Kind() flux.OperationKind {
	return PassKind
}

type PassProcedureSpec struct {
	plan.DefaultCost
}

func newPassProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*PassOpSpec)
	if !ok {
		_, ok = qs.(*OpaqueOpSpec)
		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", qs)
		}
	}

	return new(PassProcedureSpec), nil
}

func (s *PassProcedureSpec) Kind() plan.ProcedureKind {
	return PassKind
}

func (s *PassProcedureSpec) Copy() plan.ProcedureSpec {
	return new(PassProcedureSpec)
}

// PassThroughAttribute implements the PassThroughAttributer interface used
// by the planner. This implementation says that any attributes provided by
// input to this transformation are also propagated to its output.
func (s *PassProcedureSpec) PassThroughAttribute(attrKey string) bool {
	return true
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *PassProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createPassTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*PassProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	t, d := NewPassTransformation(id, s)
	return t, d, nil
}

type passTransformation struct {
	execute.ExecutionNode
	d *execute.PassthroughDataset
}

func NewPassTransformation(id execute.DatasetID, spec *PassProcedureSpec) (execute.Transformation, execute.Dataset) {
	t := &passTransformation{
		d: execute.NewPassthroughDataset(id),
	}
	return t, t.d
}

func (t *passTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *passTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	return t.d.Process(tbl)
}

func (t *passTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *passTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *passTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
