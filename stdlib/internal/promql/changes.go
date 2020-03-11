package promql

import (
	"fmt"
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const ChangesKind = "changes"

type ChangesOpSpec struct{}

func init() {
	changesSignature := runtime.MustLookupBuiltinType("internal/promql", "changes")

	runtime.RegisterPackageValue("internal/promql", ChangesKind, flux.MustValue(flux.FunctionValue(ChangesKind, createChangesOpSpec, changesSignature)))
	flux.RegisterOpSpec(ChangesKind, newChangesOp)
	plan.RegisterProcedureSpec(ChangesKind, newChangesProcedure, ChangesKind)
	execute.RegisterTransformation(ChangesKind, createChangesTransformation)
}

func createChangesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(ChangesOpSpec), nil
}

func newChangesOp() flux.OperationSpec {
	return new(ChangesOpSpec)
}

func (s *ChangesOpSpec) Kind() flux.OperationKind {
	return ChangesKind
}

type ChangesProcedureSpec struct {
	plan.DefaultCost
}

func newChangesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*ChangesOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(ChangesProcedureSpec), nil
}

func (s *ChangesProcedureSpec) Kind() plan.ProcedureKind {
	return ChangesKind
}

func (s *ChangesProcedureSpec) Copy() plan.ProcedureSpec {
	return new(ChangesProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ChangesProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createChangesTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ChangesProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewChangesTransformation(d, cache, s)
	return t, d, nil
}

type changesTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewChangesTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ChangesProcedureSpec) *changesTransformation {
	return &changesTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *changesTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *changesTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("changes found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableKeyCols(key, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found (cols: %v): %s", cols, execute.DefaultValueColLabel)
	}

	var (
		numVals int
		changes int
		prev    float64
	)
	err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(valIdx)
		for i := 0; i < cr.Len(); i++ {
			if !vs.IsValid(i) {
				continue
			}

			numVals++
			if numVals == 1 {
				prev = vs.Value(i)
				continue
			}

			current := vs.Value(i)
			if current != prev && !(math.IsNaN(current) && math.IsNaN(prev)) {
				changes++
			}
			prev = current
		}
		return nil
	})
	if err != nil {
		return err
	}

	// No output for empty input range vectors.
	if numVals < 1 {
		return nil
	}

	outValIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat})
	if err != nil {
		return fmt.Errorf("error appending value column: %s", err)
	}

	if err := builder.AppendFloat(outValIdx, float64(changes)); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

func (t *changesTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *changesTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *changesTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
