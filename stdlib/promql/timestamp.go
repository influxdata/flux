package promql

import (
	"fmt"

	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const (
	// TimestampKind is the Kind for the Timestamp Flux function
	TimestampKind = "timestamp"
)

type TimestampOpSpec struct{}

func init() {
	timestampSignature := flux.FunctionSignature(nil, nil)
	flux.RegisterPackageValue("promql", "timestamp", flux.FunctionValue(TimestampKind, createTimestampOpSpec, timestampSignature))
	flux.RegisterOpSpec(TimestampKind, func() flux.OperationSpec { return &TimestampOpSpec{} })
	plan.RegisterProcedureSpec(TimestampKind, newTimestampProcedure, TimestampKind)
	execute.RegisterTransformation(TimestampKind, createTimestampTransformation)
}

func createTimestampOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(TimestampOpSpec), nil
}

func (s *TimestampOpSpec) Kind() flux.OperationKind {
	return TimestampKind
}

func newTimestampProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*TimestampOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &TimestampProcedureSpec{}, nil
}

type TimestampProcedureSpec struct {
	plan.DefaultCost
}

func (s *TimestampProcedureSpec) Kind() plan.ProcedureKind {
	return TimestampKind
}

func (s *TimestampProcedureSpec) Copy() plan.ProcedureSpec {
	return &TimestampProcedureSpec{}
}

func createTimestampTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*TimestampProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewTimestampTransformation(d, cache, s)
	return t, d, nil
}

type timestampTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewTimestampTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *TimestampProcedureSpec) *timestampTransformation {
	return &timestampTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *timestampTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *timestampTransformation) Process(id execute.DatasetID, tbl flux.Table) (err error) {
	key := tbl.Key()
	// TODO: Update key if _value is part of key? (should never be the case for PromQL, but still...)
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("timestamp found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, cols)
	if timeIdx < 0 {
		timeIdx = execute.ColIdx(execute.DefaultStopColLabel, cols)
		if timeIdx < 0 {
			return fmt.Errorf("neither %q nor %q column not found (cols: %v)", execute.DefaultTimeColLabel, execute.DefaultStopColLabel, cols)
		}
	}
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found: %s", execute.DefaultValueColLabel)
	}

	return tbl.Do(func(cr flux.ColReader) error {
		// Copy over all non-"_value" columns.
		for j := range cr.Cols() {
			if j == valIdx {
				continue
			}
			if err := execute.AppendCol(j, j, cr, builder); err != nil {
				return err
			}
		}

		// Get the "_time" (or "_stop") of the current row as a Unix timestamp.
		for i := 0; i < cr.Len(); i++ {
			v := execute.ValueForRow(cr, i, timeIdx)
			ts := float64(v.Time().Time().UnixNano()) / 1e9
			if err := builder.AppendValue(valIdx, values.NewFloat(ts)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *timestampTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *timestampTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *timestampTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
