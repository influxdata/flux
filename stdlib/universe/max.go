package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const MaxKind = "max"

type MaxOpSpec struct {
	execute.SelectorConfig
}

func init() {
	maxSignature := runtime.MustLookupBuiltinType("universe", "max")

	runtime.RegisterPackageValue("universe", MaxKind, flux.MustValue(flux.FunctionValue(MaxKind, createMaxOpSpec, maxSignature)))
	flux.RegisterOpSpec(MaxKind, newMaxOp)
	plan.RegisterProcedureSpec(MaxKind, newMaxProcedure, MaxKind)
	execute.RegisterTransformation(MaxKind, createMaxTransformation)
}

func createMaxOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MaxOpSpec)
	if err := spec.SelectorConfig.ReadArgs(args); err != nil {
		return nil, err
	}

	return spec, nil
}

func newMaxOp() flux.OperationSpec {
	return new(MaxOpSpec)
}

func (s *MaxOpSpec) Kind() flux.OperationKind {
	return MaxKind
}

type MaxProcedureSpec struct {
	execute.SelectorConfig
}

func newMaxProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MaxOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &MaxProcedureSpec{
		SelectorConfig: spec.SelectorConfig,
	}, nil
}

func (s *MaxProcedureSpec) Kind() plan.ProcedureKind {
	return MaxKind
}
func (s *MaxProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MaxProcedureSpec)
	ns.SelectorConfig = s.SelectorConfig
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *MaxProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type MaxSelector struct {
	set  bool
	rows []execute.Row
}

func createMaxTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*MaxProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, new(MaxSelector), ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

type MaxIntSelector struct {
	MaxSelector
	max int64
}
type MaxUIntSelector struct {
	MaxSelector
	max uint64
}
type MaxFloatSelector struct {
	MaxSelector
	max float64
}
type MaxTimeSelector struct {
	MaxIntSelector
}

func (s *MaxSelector) NewTimeSelector() execute.DoTimeRowSelector {
	return new(MaxTimeSelector)
}

func (s *MaxSelector) NewBoolSelector() execute.DoBoolRowSelector {
	return nil
}

func (s *MaxSelector) NewIntSelector() execute.DoIntRowSelector {
	return new(MaxIntSelector)
}

func (s *MaxSelector) NewUIntSelector() execute.DoUIntRowSelector {
	return new(MaxUIntSelector)
}

func (s *MaxSelector) NewFloatSelector() execute.DoFloatRowSelector {
	return new(MaxFloatSelector)
}

func (s *MaxSelector) NewStringSelector() execute.DoStringRowSelector {
	return nil
}

func (s *MaxSelector) Rows() []execute.Row {
	if !s.set {
		return nil
	}
	return s.rows
}

func (s *MaxSelector) selectRow(idx int, cr flux.ColReader) {
	// Capture row
	if idx >= 0 {
		s.rows = []execute.Row{execute.ReadRow(idx, cr)}
	}
}

func (s *MaxTimeSelector) DoTime(vs *array.Int64, cr flux.ColReader) {
	s.MaxIntSelector.DoInt(vs, cr)
}
func (s *MaxIntSelector) DoInt(vs *array.Int64, cr flux.ColReader) {
	maxIdx := -1
	for i := 0; i < vs.Len(); i++ {
		if vs.IsValid(i) {
			if v := vs.Value(i); !s.set || v > s.max {
				s.set = true
				s.max = v
				maxIdx = i
			}
		}
	}
	s.selectRow(maxIdx, cr)
}
func (s *MaxUIntSelector) DoUInt(vs *array.Uint64, cr flux.ColReader) {
	maxIdx := -1
	for i := 0; i < vs.Len(); i++ {
		if vs.IsValid(i) {
			if v := vs.Value(i); !s.set || v > s.max {
				s.set = true
				s.max = v
				maxIdx = i
			}
		}
	}
	s.selectRow(maxIdx, cr)
}
func (s *MaxFloatSelector) DoFloat(vs *array.Float64, cr flux.ColReader) {
	maxIdx := -1
	for i := 0; i < vs.Len(); i++ {
		if vs.IsValid(i) {
			if v := vs.Value(i); !s.set || v > s.max {
				s.set = true
				s.max = v
				maxIdx = i
			}
		}
	}
	s.selectRow(maxIdx, cr)
}
