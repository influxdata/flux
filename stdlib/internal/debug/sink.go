package debug

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const SinkKind = "internal/debug.sink"

type SinkOpSpec struct{}

func init() {
	sinkSignature := runtime.MustLookupBuiltinType("internal/debug", "sink")

	runtime.RegisterPackageValue("internal/debug", "sink", flux.MustValue(flux.FunctionValue(SinkKind, createSinkOpSpec, sinkSignature)))
	flux.RegisterOpSpec(SinkKind, newSinkOp)
	plan.RegisterProcedureSpec(SinkKind, newSinkProcedure, SinkKind)
	execute.RegisterTransformation(SinkKind, createSinkTransformation)
}

func createSinkOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(SinkOpSpec), nil
}

func newSinkOp() flux.OperationSpec {
	return new(SinkOpSpec)
}

func (s *SinkOpSpec) Kind() flux.OperationKind {
	return SinkKind
}

type SinkProcedureSpec struct {
	plan.DefaultCost
}

func newSinkProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*SinkOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(SinkProcedureSpec), nil
}

func (s *SinkProcedureSpec) Kind() plan.ProcedureKind {
	return SinkKind
}

func (s *SinkProcedureSpec) Copy() plan.ProcedureSpec {
	return new(SinkProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SinkProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createSinkTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SinkProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	t, d := NewSinkTransformation(id, s)
	return t, d, nil
}

type sinkTransformation struct {
	execute.ExecutionNode
	d *execute.PassthroughDataset
}

func NewSinkTransformation(id execute.DatasetID, spec *SinkProcedureSpec) (execute.Transformation, execute.Dataset) {
	t := &sinkTransformation{
		d: execute.NewPassthroughDataset(id),
	}
	return t, t.d
}

func (t *sinkTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *sinkTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	return tbl.Do(func(cr flux.ColReader) error {
		return nil
	})
}

func (t *sinkTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *sinkTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *sinkTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
