package universe

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const LastKind = "last"

type LastOpSpec struct {
	execute.SelectorConfig
}

func init() {
	lastSignature := execute.SelectorSignature(nil, nil)

	flux.RegisterPackageValue("universe", LastKind, flux.FunctionValue(LastKind, createLastOpSpec, lastSignature))
	flux.RegisterOpSpec(LastKind, newLastOp)
	plan.RegisterProcedureSpec(LastKind, newLastProcedure, LastKind)
	execute.RegisterTransformation(LastKind, createLastTransformation)
}

func createLastOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LastOpSpec)
	if err := spec.SelectorConfig.ReadArgs(args); err != nil {
		return nil, err
	}
	return spec, nil
}

func newLastOp() flux.OperationSpec {
	return new(LastOpSpec)
}

func (s *LastOpSpec) Kind() flux.OperationKind {
	return LastKind
}

type LastProcedureSpec struct {
	execute.SelectorConfig
}

func newLastProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LastOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &LastProcedureSpec{
		SelectorConfig: spec.SelectorConfig,
	}, nil
}

func (s *LastProcedureSpec) Kind() plan.ProcedureKind {
	return LastKind
}

func (s *LastProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LastProcedureSpec)
	ns.SelectorConfig = s.SelectorConfig
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LastProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type LastSelector struct {
	selected bool
}

func createLastTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*LastProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, new(LastSelector), ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

func (s *LastSelector) reset() {
	s.selected = false
}
func (s *LastSelector) NewBoolSelector() execute.DoBoolIndexSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewIntSelector() execute.DoIntIndexSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewUIntSelector() execute.DoUIntIndexSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewFloatSelector() execute.DoFloatIndexSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewStringSelector() execute.DoStringIndexSelector {
	s.reset()
	return s
}

func (s *LastSelector) selectLast(vs array.Interface) []int {
	if !s.selected {
		sz := vs.Len()
		for i := sz - 1; i >= 0; i-- {
			if !vs.IsNull(i) {
				s.selected = true
				return []int{i}
			}
		}
	}
	return nil
}

func (s *LastSelector) DoBool(vs *array.Boolean) []int {
	return s.selectLast(vs)
}
func (s *LastSelector) DoInt(vs *array.Int64) []int {
	return s.selectLast(vs)
}
func (s *LastSelector) DoUInt(vs *array.Uint64) []int {
	return s.selectLast(vs)
}
func (s *LastSelector) DoFloat(vs *array.Float64) []int {
	return s.selectLast(vs)
}
func (s *LastSelector) DoString(vs *array.Binary) []int {
	return s.selectLast(vs)
}
