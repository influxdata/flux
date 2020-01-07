package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

const FirstKind = "first"

type FirstOpSpec struct {
	execute.SelectorConfig
}

func init() {
	firstSignature := semantic.LookupBuiltInType("universe", "first")

	flux.RegisterPackageValue("universe", FirstKind, flux.MustValue(flux.FunctionValue(FirstKind, createFirstOpSpec, firstSignature)))
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
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &FirstProcedureSpec{
		SelectorConfig: spec.SelectorConfig,
	}, nil
}

func (s *FirstProcedureSpec) Kind() plan.ProcedureKind {
	return FirstKind
}

func (s *FirstProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FirstProcedureSpec)
	*ns = *s
	ns.SelectorConfig = s.SelectorConfig
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *FirstProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type FirstSelector struct {
	selected bool
}

func createFirstTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*FirstProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", ps)
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, new(FirstSelector), ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

func (s *FirstSelector) reset() {
	s.selected = false
}

func (s *FirstSelector) NewTimeSelector() execute.DoTimeIndexSelector {
	s.reset()
	return s
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

func (s *FirstSelector) selectFirst(vs array.Interface) []int {
	if !s.selected {
		sz := vs.Len()
		for i := 0; i < sz; i++ {
			if !vs.IsNull(i) {
				s.selected = true
				return []int{i}
			}
		}
	}
	return nil
}

func (s *FirstSelector) DoTime(vs *array.Int64) []int {
	return s.selectFirst(vs)
}
func (s *FirstSelector) DoBool(vs *array.Boolean) []int {
	return s.selectFirst(vs)
}
func (s *FirstSelector) DoInt(vs *array.Int64) []int {
	return s.selectFirst(vs)
}
func (s *FirstSelector) DoUInt(vs *array.Uint64) []int {
	return s.selectFirst(vs)
}
func (s *FirstSelector) DoFloat(vs *array.Float64) []int {
	return s.selectFirst(vs)
}
func (s *FirstSelector) DoString(vs *array.Binary) []int {
	return s.selectFirst(vs)
}
