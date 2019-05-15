package promql

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const ResetsKind = "resets"

type ResetsOpSpec struct{}

func init() {
	resetsSignature := flux.FunctionSignature(nil, nil)

	flux.RegisterPackageValue("promql", ResetsKind, flux.FunctionValue(ResetsKind, createResetsOpSpec, resetsSignature))
	flux.RegisterOpSpec(ResetsKind, newResetsOp)
	plan.RegisterProcedureSpec(ResetsKind, newResetsProcedure, ResetsKind)
	execute.RegisterTransformation(ResetsKind, createResetsTransformation)
}

func createResetsOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(ResetsOpSpec), nil
}

func newResetsOp() flux.OperationSpec {
	return new(ResetsOpSpec)
}

func (s *ResetsOpSpec) Kind() flux.OperationKind {
	return ResetsKind
}

type ResetsProcedureSpec struct {
	plan.DefaultCost
}

func newResetsProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*ResetsOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(ResetsProcedureSpec), nil
}

func (s *ResetsProcedureSpec) Kind() plan.ProcedureKind {
	return ResetsKind
}

func (s *ResetsProcedureSpec) Copy() plan.ProcedureSpec {
	return new(ResetsProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ResetsProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createResetsTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ResetsProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewResetsTransformation(d, cache, s)
	return t, d, nil
}

type resetsTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewResetsTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ResetsProcedureSpec) *resetsTransformation {
	return &resetsTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *resetsTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *resetsTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("resets found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, cols)
	if timeIdx < 0 {
		return fmt.Errorf("time column not found (cols: %v): %s", cols, execute.DefaultTimeColLabel)
	}
	stopIdx := execute.ColIdx(execute.DefaultStopColLabel, cols)
	if stopIdx < 0 {
		return fmt.Errorf("stop column not found (cols: %v): %s", cols, execute.DefaultStopColLabel)
	}
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found (cols: %v): %s", cols, execute.DefaultValueColLabel)
	}

	if key.Value(stopIdx).Type() != semantic.Time {
		return fmt.Errorf("stop column is not of time type")
	}

	var (
		numVals int
		resets  int
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
			if current < prev {
				resets++
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

	if err := builder.AppendTime(timeIdx, key.ValueTime(stopIdx)); err != nil {
		return err
	}
	if err := builder.AppendFloat(valIdx, float64(resets)); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

func (t *resetsTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *resetsTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *resetsTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
