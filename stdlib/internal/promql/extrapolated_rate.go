// Flux adaptation of PromQL's extrapolatedRate() helper function:
// https://github.com/prometheus/prometheus/blob/f04b1b5559a80a4fd1745cf891ce392a056460c9/promql/functions.go#L65

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

const ExtrapolatedRateKind = "extrapolatedRate"

type ExtrapolatedRateOpSpec struct {
	IsCounter bool `json:"isCounter"`
	IsRate    bool `json:"isRate"`
}

func init() {
	extrapolatedRateSignature := semantic.MustLookupBuiltinType("internal/promql", ExtrapolatedRateKind)

	flux.RegisterPackageValue("internal/promql", ExtrapolatedRateKind, flux.MustValue(flux.FunctionValue(ExtrapolatedRateKind, createExtrapolatedRateOpSpec, extrapolatedRateSignature)))
	flux.RegisterOpSpec(ExtrapolatedRateKind, newExtrapolatedRateOp)
	plan.RegisterProcedureSpec(ExtrapolatedRateKind, newExtrapolatedRateProcedure, ExtrapolatedRateKind)
	execute.RegisterTransformation(ExtrapolatedRateKind, createExtrapolatedRateTransformation)
}

func createExtrapolatedRateOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ExtrapolatedRateOpSpec)

	if ic, ok, err := args.GetBool("isCounter"); err != nil {
		return nil, err
	} else if ok {
		spec.IsCounter = ic
	}

	if ir, ok, err := args.GetBool("isRate"); err != nil {
		return nil, err
	} else if ok {
		spec.IsRate = ir
	}

	return spec, nil
}

func newExtrapolatedRateOp() flux.OperationSpec {
	return new(ExtrapolatedRateOpSpec)
}

func (s *ExtrapolatedRateOpSpec) Kind() flux.OperationKind {
	return ExtrapolatedRateKind
}

type ExtrapolatedRateProcedureSpec struct {
	plan.DefaultCost
	IsCounter bool
	IsRate    bool
}

func newExtrapolatedRateProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ExtrapolatedRateOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ExtrapolatedRateProcedureSpec{
		IsCounter: spec.IsCounter,
		IsRate:    spec.IsRate,
	}, nil
}

func (s *ExtrapolatedRateProcedureSpec) Kind() plan.ProcedureKind {
	return ExtrapolatedRateKind
}

func (s *ExtrapolatedRateProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ExtrapolatedRateProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ExtrapolatedRateProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createExtrapolatedRateTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ExtrapolatedRateProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewExtrapolatedRateTransformation(d, cache, s)
	return t, d, nil
}

type extrapolatedRateTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	isCounter bool
	isRate    bool
}

func NewExtrapolatedRateTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ExtrapolatedRateProcedureSpec) *extrapolatedRateTransformation {
	return &extrapolatedRateTransformation{
		d:         d,
		cache:     cache,
		isCounter: spec.IsCounter,
		isRate:    spec.IsRate,
	}
}

func (t *extrapolatedRateTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *extrapolatedRateTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("extrapolatedRate found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableKeyCols(key, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, cols)
	if timeIdx < 0 {
		return fmt.Errorf("time column not found (cols: %v): %s", cols, execute.DefaultTimeColLabel)
	}
	startIdx := execute.ColIdx(execute.DefaultStartColLabel, cols)
	if startIdx < 0 {
		return fmt.Errorf("start column not found (cols: %v): %s", cols, execute.DefaultStartColLabel)
	}
	stopIdx := execute.ColIdx(execute.DefaultStopColLabel, cols)
	if stopIdx < 0 {
		return fmt.Errorf("stop column not found (cols: %v): %s", cols, execute.DefaultStopColLabel)
	}
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found (cols: %v): %s", cols, execute.DefaultValueColLabel)
	}

	if key.Value(startIdx).Type().Nature() != semantic.Time {
		return fmt.Errorf("start column is not of time type")
	}
	if key.Value(stopIdx).Type().Nature() != semantic.Time {
		return fmt.Errorf("stop column is not of time type")
	}
	rangeStart := key.ValueTime(startIdx).Time()
	rangeEnd := key.ValueTime(stopIdx).Time()

	var (
		numVals           int
		counterCorrection float64
		firstValue        float64
		firstTime         time.Time
		lastValue         float64
		lastTime          time.Time
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

			if t.isCounter && v < lastValue {
				counterCorrection += lastValue
			}

			if numVals == 0 {
				firstValue = v
				firstTime = ts
			}
			lastValue = v
			lastTime = ts

			numVals++
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

	resultValue := lastValue - firstValue + counterCorrection

	// Duration between first/last samples and boundary of range.
	durationToStart := float64(firstTime.Sub(rangeStart))
	durationToEnd := float64(rangeEnd.Sub(lastTime))

	sampledInterval := float64(lastTime.Sub(firstTime))
	averageDurationBetweenSamples := sampledInterval / float64(numVals-1)

	if t.isCounter && resultValue > 0 && firstValue >= 0 {
		// Counters cannot be negative. If we have any slope at
		// all (i.e. resultValue went up), we can extrapolate
		// the zero point of the counter. If the duration to the
		// zero point is shorter than the durationToStart, we
		// take the zero point as the start of the series,
		// thereby avoiding extrapolation to negative counter
		// values.
		durationToZero := sampledInterval * (firstValue / resultValue)
		if durationToZero < durationToStart {
			durationToStart = durationToZero
		}
	}

	// If the first/last samples are close to the boundaries of the range,
	// extrapolate the result. This is as we expect that another sample
	// will exist given the spacing between samples we've seen thus far,
	// with an allowance for noise.
	extrapolationThreshold := averageDurationBetweenSamples * 1.1
	extrapolateToInterval := sampledInterval

	if durationToStart < extrapolationThreshold {
		extrapolateToInterval += durationToStart
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	if durationToEnd < extrapolationThreshold {
		extrapolateToInterval += durationToEnd
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	resultValue = resultValue * (extrapolateToInterval / sampledInterval)
	if t.isRate {
		resultValue = resultValue / rangeEnd.Sub(rangeStart).Seconds()
	}

	outValIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat})
	if err != nil {
		return fmt.Errorf("error appending value column: %s", err)
	}

	if err := builder.AppendFloat(outValIdx, resultValue); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

func (t *extrapolatedRateTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *extrapolatedRateTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *extrapolatedRateTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
