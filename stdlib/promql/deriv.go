package promql

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const DerivKind = "deriv"

type DerivOpSpec struct{}

func init() {
	derivSignature := flux.FunctionSignature(nil, nil)

	flux.RegisterPackageValue("promql", DerivKind, flux.FunctionValue(DerivKind, createDerivOpSpec, derivSignature))
	flux.RegisterOpSpec(DerivKind, newDerivOp)
	plan.RegisterProcedureSpec(DerivKind, newDerivProcedure, DerivKind)
	execute.RegisterTransformation(DerivKind, createDerivTransformation)
}

func createDerivOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(DerivOpSpec), nil
}

func newDerivOp() flux.OperationSpec {
	return new(DerivOpSpec)
}

func (s *DerivOpSpec) Kind() flux.OperationKind {
	return DerivKind
}

type DerivProcedureSpec struct {
	plan.DefaultCost
}

func newDerivProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*DerivOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(DerivProcedureSpec), nil
}

func (s *DerivProcedureSpec) Kind() plan.ProcedureKind {
	return DerivKind
}

func (s *DerivProcedureSpec) Copy() plan.ProcedureSpec {
	return new(DerivProcedureSpec)
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *DerivProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createDerivTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DerivProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDerivTransformation(d, cache, s)
	return t, d, nil
}

type derivTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewDerivTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *DerivProcedureSpec) *derivTransformation {
	return &derivTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *derivTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *derivTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("deriv found duplicate table with key: %v", tbl.Key())
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
		return fmt.Errorf("start column not found (cols: %v): %s", cols, execute.DefaultStopColLabel)
	}
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found (cols: %v): %s", cols, execute.DefaultValueColLabel)
	}

	if key.Value(stopIdx).Type() != semantic.Time {
		return fmt.Errorf("stop column is not of time type")
	}

	var (
		numVals      int
		sumX, sumY   float64
		sumXY, sumX2 float64
		firstTime    time.Time
	)
	err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(valIdx)
		times := cr.Times(timeIdx)
		for i := 0; i < cr.Len(); i++ {
			if !vs.IsValid(i) || !times.IsValid(i) {
				continue
			}

			v := vs.Value(i)
			ts := values.Time(times.Value(i)).Time()

			if numVals == 0 {
				firstTime = ts
			}

			x := float64(ts.Sub(firstTime).Seconds())
			numVals++
			sumY += v
			sumX += x
			sumXY += x * v
			sumX2 += x * x
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Omit output table if there are not at least two samples to compute a rate from.
	if numVals < 2 {
		return nil
	}

	n := float64(numVals)
	covXY := sumXY - sumX*sumY/n
	varX := sumX2 - sumX*sumX/n

	slope := covXY / varX
	// TODO: only needed for predict_linear()
	//intercept := sumY/n - slope*sumX/n

	if err := builder.AppendTime(timeIdx, key.ValueTime(stopIdx)); err != nil {
		return err
	}
	if err := builder.AppendFloat(valIdx, slope); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

func (t *derivTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *derivTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *derivTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
