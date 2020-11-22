package events

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const pkgPath = "events"

const StateChangesKind = "stateChanges"

const stateColumn = "state"

type StateChangesOpSpec struct{}

func init() {
	sig := runtime.MustLookupBuiltinType(pkgPath, StateChangesKind)
	runtime.RegisterPackageValue(pkgPath, StateChangesKind, flux.MustValue(flux.FunctionValue(StateChangesKind, createStateChangesOpSpec, sig)))
	flux.RegisterOpSpec(StateChangesKind, newStateChangesOp)
	plan.RegisterProcedureSpec(StateChangesKind, newStateChangesProcedure, StateChangesKind)
	execute.RegisterTransformation(StateChangesKind, createStateChangesTransformation)
}

func createStateChangesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(StateChangesOpSpec)

	return spec, nil
}

func newStateChangesOp() flux.OperationSpec {
	return new(StateChangesOpSpec)
}

func (s *StateChangesOpSpec) Kind() flux.OperationKind {
	return StateChangesKind
}

type StateChangesProcedureSpec struct {
	plan.DefaultCost
}

func newStateChangesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*StateChangesOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &StateChangesProcedureSpec{}, nil
}

func (s *StateChangesProcedureSpec) Kind() plan.ProcedureKind {
	return StateChangesKind
}

func (s *StateChangesProcedureSpec) Copy() plan.ProcedureSpec {
	return &StateChangesProcedureSpec{}
}

func createStateChangesTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*StateChangesProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewStateChangesTransformation(d, cache, s)
	return t, d, nil
}

type stateChangesTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewStateChangesTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *StateChangesProcedureSpec) *stateChangesTransformation {
	return &stateChangesTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *stateChangesTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *stateChangesTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *stateChangesTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *stateChangesTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *stateChangesTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()

	err := execute.AddTableCols(tbl, builder)
	if err != nil {
		return err
	}

	stateIdx := execute.ColIdx(stateColumn, cols)
	if stateIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "column %q does not exist", stateColumn)
	}

	colMap := execute.ColMap([]int{0}, builder, tbl.Cols())

	var prev values.Value
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			curr := execute.ValueForRow(cr, i, stateIdx)
			if prev == nil || !curr.Equal(prev) {
				if err := execute.AppendMappedRecordExplicit(i, cr, builder, colMap); err != nil {
					return err
				}
			}
			prev = curr
		}
		return nil
	})
}
