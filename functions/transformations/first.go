package transformations

import (
	"fmt"
	"github.com/influxdata/flux/functions/inputs"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const FirstKind = "first"

type FirstOpSpec struct {
	execute.SelectorConfig
}

var firstSignature = execute.DefaultSelectorSignature()

func init() {
	flux.RegisterFunction(FirstKind, createFirstOpSpec, firstSignature)
	flux.RegisterOpSpec(FirstKind, newFirstOp)
	plan.RegisterProcedureSpec(FirstKind, newFirstProcedure, FirstKind)
	execute.RegisterTransformation(FirstKind, createFirstTransformation)
}

func createFirstOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(FirstOpSpec)
	if err := spec.SelectorConfig.ReadArgs(args); err != nil {
		return nil, err
	}

	return spec, nil
}

func newFirstOp() flux.OperationSpec {
	return new(FirstOpSpec)
}

func (s *FirstOpSpec) Kind() flux.OperationKind {
	return FirstKind
}

type FirstProcedureSpec struct {
	execute.SelectorConfig
}

func newFirstProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FirstOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &FirstProcedureSpec{
		SelectorConfig: spec.SelectorConfig,
	}, nil
}

func (s *FirstProcedureSpec) Kind() plan.ProcedureKind {
	return FirstKind
}
func (s *FirstProcedureSpec) PushDownRules() []plan.PushDownRule {
	return []plan.PushDownRule{{
		Root:    inputs.FromKind,
		Through: []plan.ProcedureKind{GroupKind, LimitKind, FilterKind},
		Match: func(spec plan.ProcedureSpec) bool {
			selectSpec := spec.(*inputs.FromProcedureSpec)
			return !selectSpec.AggregateSet
		},
	}}
}

func (s *FirstProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*inputs.FromProcedureSpec)
	if selectSpec.BoundsSet || selectSpec.LimitSet || selectSpec.DescendingSet {
		root = dup()
		selectSpec = root.Spec.(*inputs.FromProcedureSpec)
		selectSpec.BoundsSet = false
		selectSpec.Bounds = flux.Bounds{}
		selectSpec.LimitSet = false
		selectSpec.PointsLimit = 0
		selectSpec.SeriesLimit = 0
		selectSpec.SeriesOffset = 0
		selectSpec.DescendingSet = false
		selectSpec.Descending = false
		return
	}
	selectSpec.BoundsSet = true
	selectSpec.Bounds = flux.Bounds{
		Start: flux.MinTime,
		Stop:  flux.Now,
	}
	selectSpec.LimitSet = true
	selectSpec.PointsLimit = 1
	selectSpec.DescendingSet = true
	selectSpec.Descending = false
}
func (s *FirstProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FirstProcedureSpec)
	*ns = *s
	ns.SelectorConfig = s.SelectorConfig
	return ns
}

type FirstSelector struct {
	selected bool
}

func createFirstTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*FirstProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, new(FirstSelector), ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

func (s *FirstSelector) reset() {
	s.selected = false
}

func (s *FirstSelector) NewBoolSelector() execute.DoBoolIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewIntSelector() execute.DoIntIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewUIntSelector() execute.DoUIntIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewFloatSelector() execute.DoFloatIndexSelector {
	s.reset()
	return s
}
func (s *FirstSelector) NewStringSelector() execute.DoStringIndexSelector {
	s.reset()
	return s
}

func (s *FirstSelector) selectFirst(l int) []int {
	if !s.selected && l > 0 {
		s.selected = true
		return []int{0}
	}
	return nil
}
func (s *FirstSelector) DoBool(vs []bool) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoInt(vs []int64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoUInt(vs []uint64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoFloat(vs []float64) []int {
	return s.selectFirst(len(vs))
}
func (s *FirstSelector) DoString(vs []string) []int {
	return s.selectFirst(len(vs))
}
