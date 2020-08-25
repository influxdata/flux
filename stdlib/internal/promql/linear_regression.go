package promql

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const LinearRegressionKind = "linearRegression"

type LinearRegressionOpSpec struct {
	Predict bool `json:"predict"`
	// Stored as seconds in float64 to avoid back-and-forth duration conversions from PromQL.
	FromNow float64 `json:"fromNow"`
}

func init() {
	linearRegressionSignature := runtime.MustLookupBuiltinType("internal/promql", LinearRegressionKind)

	runtime.RegisterPackageValue("internal/promql", LinearRegressionKind, flux.MustValue(flux.FunctionValue(LinearRegressionKind, createLinearRegressionOpSpec, linearRegressionSignature)))
	flux.RegisterOpSpec(LinearRegressionKind, newLinearRegressionOp)
	plan.RegisterProcedureSpec(LinearRegressionKind, newLinearRegressionProcedure, LinearRegressionKind)
	execute.RegisterTransformation(LinearRegressionKind, createLinearRegressionTransformation)
}

func createLinearRegressionOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LinearRegressionOpSpec)

	if p, ok, err := args.GetBool("predict"); err != nil {
		return nil, err
	} else if ok {
		spec.Predict = p
	}

	if d, ok, err := args.GetFloat("fromNow"); err != nil {
		return nil, err
	} else if ok {
		spec.FromNow = d
	}
	return spec, nil
}

func newLinearRegressionOp() flux.OperationSpec {
	return new(LinearRegressionOpSpec)
}

func (s *LinearRegressionOpSpec) Kind() flux.OperationKind {
	return LinearRegressionKind
}

type LinearRegressionProcedureSpec struct {
	plan.DefaultCost
	Predict bool
	FromNow float64
}

func newLinearRegressionProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LinearRegressionOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &LinearRegressionProcedureSpec{
		Predict: spec.Predict,
		FromNow: spec.FromNow,
	}, nil
}

func (s *LinearRegressionProcedureSpec) Kind() plan.ProcedureKind {
	return LinearRegressionKind
}

func (s *LinearRegressionProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LinearRegressionProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LinearRegressionProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createLinearRegressionTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LinearRegressionProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewLinearRegressionTransformation(d, cache, s)
	return t, d, nil
}

type linearRegressionTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	predict bool
	fromNow float64
}

func NewLinearRegressionTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *LinearRegressionProcedureSpec) *linearRegressionTransformation {
	return &linearRegressionTransformation{
		d:       d,
		cache:   cache,
		predict: spec.Predict,
		fromNow: spec.FromNow,
	}
}

func (t *linearRegressionTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *linearRegressionTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("linearRegression found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableKeyCols(key, builder); err != nil {
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

	if key.Value(stopIdx).Type().Nature() != semantic.Time {
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
				// Subtle difference between deriv() and predict_linear() intercept time.
				if t.predict {
					firstTime = key.ValueTime(stopIdx).Time()
				} else {
					firstTime = ts
				}
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

	resultValue := slope
	if t.predict {
		intercept := sumY/n - slope*sumX/n
		resultValue = slope*t.fromNow + intercept
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

func (t *linearRegressionTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *linearRegressionTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *linearRegressionTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
