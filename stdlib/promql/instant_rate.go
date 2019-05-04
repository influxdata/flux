// Flux adaptation of PromQL's instantValue() helper function:
// https://github.com/prometheus/prometheus/blob/45506841e664665e8f0b1b59f416c91643913a3f/promql/functions.go#L167

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

const InstantRateKind = "instantRate"

type InstantRateOpSpec struct {
	IsRate bool `json:"isRate"`
}

func init() {
	instantRateSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"isRate": semantic.Bool,
		},
		nil,
	)

	flux.RegisterPackageValue("promql", InstantRateKind, flux.FunctionValue(InstantRateKind, createInstantRateOpSpec, instantRateSignature))
	flux.RegisterOpSpec(InstantRateKind, newInstantRateOp)
	plan.RegisterProcedureSpec(InstantRateKind, newInstantRateProcedure, InstantRateKind)
	execute.RegisterTransformation(InstantRateKind, createInstantRateTransformation)
}

func createInstantRateOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(InstantRateOpSpec)

	if ir, ok, err := args.GetBool("isRate"); err != nil {
		return nil, err
	} else if ok {
		spec.IsRate = ir
	}

	return spec, nil
}

func newInstantRateOp() flux.OperationSpec {
	return new(InstantRateOpSpec)
}

func (s *InstantRateOpSpec) Kind() flux.OperationKind {
	return InstantRateKind
}

type InstantRateProcedureSpec struct {
	plan.DefaultCost
	IsRate bool
}

func newInstantRateProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*InstantRateOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &InstantRateProcedureSpec{
		IsRate: spec.IsRate,
	}, nil
}

func (s *InstantRateProcedureSpec) Kind() plan.ProcedureKind {
	return InstantRateKind
}

func (s *InstantRateProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(InstantRateProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *InstantRateProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createInstantRateTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*InstantRateProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewInstantRateTransformation(d, cache, s)
	return t, d, nil
}

type instantRateTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	isRate bool
}

func NewInstantRateTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *InstantRateProcedureSpec) *instantRateTransformation {
	return &instantRateTransformation{
		d:      d,
		cache:  cache,
		isRate: spec.IsRate,
	}
}

func (t *instantRateTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *instantRateTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("instantRate found duplicate table with key: %v", tbl.Key())
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
		numVals   int
		lastValue float64
		lastTime  time.Time
		prevValue float64
		prevTime  time.Time
	)
	err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(valIdx)
		times := cr.Times(timeIdx)
		for i := 0; i < cr.Len(); i++ {
			if !vs.IsValid(i) || !times.IsValid(i) {
				continue
			}
			numVals++

			prevValue = lastValue
			prevTime = lastTime

			lastValue = vs.Value(i)
			lastTime = values.Time(times.Value(i)).Time()
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

	var resultValue float64
	if t.isRate && lastValue < prevValue {
		// Counter reset.
		resultValue = lastValue
	} else {
		resultValue = lastValue - prevValue
	}

	sampledInterval := lastTime.Sub(prevTime)
	if sampledInterval == 0 {
		// Avoid dividing by 0.
		return nil
	}

	if t.isRate {
		// Convert to per-second.
		resultValue /= sampledInterval.Seconds()
	}

	if err := builder.AppendTime(timeIdx, key.ValueTime(stopIdx)); err != nil {
		return err
	}
	if err := builder.AppendFloat(valIdx, resultValue); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

func (t *instantRateTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *instantRateTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *instantRateTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
