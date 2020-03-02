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

const LastKind = "last"

type LastOpSpec struct {
	execute.SelectorConfig
}

func init() {
	lastSignature := runtime.MustLookupBuiltinType("universe", "last")

	runtime.RegisterPackageValue("universe", LastKind, flux.MustValue(flux.FunctionValue(LastKind, createLastOpSpec, lastSignature)))
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
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
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

// LastSelector selects the last row from a Flux table.
// Note that while 'last' and 'first' are conceptually similar, one is a
// row selector (last) while the other is an index selector (first). The
// reason for this is that it was easier to ensure a correct implementation
// of 'last' by defining it as a row selector when using multiple column
// readers to iterate over a Flux table.
type LastSelector struct {
	rows []execute.Row
}

func createLastTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*LastProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, new(LastSelector), ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

func (s *LastSelector) reset() {
	s.rows = nil
}
func (s *LastSelector) NewTimeSelector() execute.DoTimeRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewBoolSelector() execute.DoBoolRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewIntSelector() execute.DoIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewUIntSelector() execute.DoUIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewFloatSelector() execute.DoFloatRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewStringSelector() execute.DoStringRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) Rows() []execute.Row {
	return s.rows
}

func (s *LastSelector) selectLast(vs array.Interface, cr flux.ColReader) {
	for i := vs.Len() - 1; i >= 0; i-- {
		if !vs.IsNull(i) {
			s.rows = []execute.Row{execute.ReadRow(i, cr)}
			return
		}
	}
}

func (s *LastSelector) DoTime(vs *array.Int64, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
func (s *LastSelector) DoBool(vs *array.Boolean, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
func (s *LastSelector) DoInt(vs *array.Int64, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
func (s *LastSelector) DoUInt(vs *array.Uint64, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
func (s *LastSelector) DoFloat(vs *array.Float64, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
func (s *LastSelector) DoString(vs *array.Binary, cr flux.ColReader) {
	s.selectLast(vs, cr)
}
