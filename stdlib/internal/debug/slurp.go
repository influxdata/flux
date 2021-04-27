package debug

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const SlurpKind = "internal/debug.slurp"

type SlurpOpSpec struct{}

func init() {
	slurpSignature := runtime.MustLookupBuiltinType("internal/debug", "slurp")

	runtime.RegisterPackageValue("internal/debug", "slurp", flux.MustValue(flux.FunctionValue(SlurpKind, createSlurpOpSpec, slurpSignature)))
	flux.RegisterOpSpec(SlurpKind, newSlurpOp)
	plan.RegisterProcedureSpec(SlurpKind, newSlurpProcedure, SlurpKind)
	execute.RegisterTransformation(SlurpKind, createSlurpTransformation)
}

func createSlurpOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(SlurpOpSpec), nil
}

func newSlurpOp() flux.OperationSpec {
	return new(SlurpOpSpec)
}

func (s *SlurpOpSpec) Kind() flux.OperationKind {
	return SlurpKind
}

type SlurpProcedureSpec struct {
	plan.DefaultCost
}

func newSlurpProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*SlurpOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(SlurpProcedureSpec), nil
}

func (s *SlurpProcedureSpec) Kind() plan.ProcedureKind {
	return SlurpKind
}

func (s *SlurpProcedureSpec) Copy() plan.ProcedureSpec {
	return new(SlurpProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SlurpProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createSlurpTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SlurpProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	t, d := NewSlurpTransformation(id, s, a)
	return t, d, nil
}

type slurpTransformation struct {
	execute.ExecutionNode
	d   *execute.PassthroughDataset
	mem memory.Allocator
}

func NewSlurpTransformation(id execute.DatasetID, spec *SlurpProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset) {
	t := &slurpTransformation{
		d:   execute.NewPassthroughDataset(id),
		mem: a.Allocator(),
	}
	return t, t.d
}

func (t *slurpTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *slurpTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	b := table.NewArrowBuilder(tbl.Key(), t.mem)
	b.Init(tbl.Cols())

	if err := tbl.Do(func(cr flux.ColReader) error {
		for i := range cr.Cols() {
			arrowutil.CopyTo(b.Builders[i], table.Values(cr, i))
		}
		return nil
	}); err != nil {
		return err
	}

	out, err := b.Table()
	if err != nil {
		return err
	}
	return t.d.Process(out)
}

func (t *slurpTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *slurpTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *slurpTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
