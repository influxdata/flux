// Flux adaptation of PromQL's holtWinters() helper function:
// https://github.com/prometheus/prometheus/blob/f04b1b5559a80a4fd1745cf891ce392a056460c9/promql/functions.go#L65

package promql

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const HoltWintersKind = "promHoltWinters"

type HoltWintersOpSpec struct {
	SmoothingFactor float64 `json:"smoothingFactor"`
	TrendFactor     float64 `json:"trendFactor"`
}

func init() {
	holtWintersSignature := runtime.MustLookupBuiltinType("internal/promql", "holtWinters")

	runtime.RegisterPackageValue("internal/promql", "holtWinters", flux.MustValue(flux.FunctionValue(HoltWintersKind, createHoltWintersOpSpec, holtWintersSignature)))
	flux.RegisterOpSpec(HoltWintersKind, newHoltWintersOp)
	plan.RegisterProcedureSpec(HoltWintersKind, newHoltWintersProcedure, HoltWintersKind)
	execute.RegisterTransformation(HoltWintersKind, createHoltWintersTransformation)
}

func createHoltWintersOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(HoltWintersOpSpec)

	if sm, ok, err := args.GetFloat("smoothingFactor"); err != nil {
		return nil, err
	} else if ok {
		spec.SmoothingFactor = sm
	}

	if tf, ok, err := args.GetFloat("trendFactor"); err != nil {
		return nil, err
	} else if ok {
		spec.TrendFactor = tf
	}

	return spec, nil
}

func newHoltWintersOp() flux.OperationSpec {
	return new(HoltWintersOpSpec)
}

func (s *HoltWintersOpSpec) Kind() flux.OperationKind {
	return HoltWintersKind
}

type HoltWintersProcedureSpec struct {
	plan.DefaultCost
	SmoothingFactor float64
	TrendFactor     float64
}

func newHoltWintersProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*HoltWintersOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &HoltWintersProcedureSpec{
		SmoothingFactor: spec.SmoothingFactor,
		TrendFactor:     spec.TrendFactor,
	}, nil
}

func (s *HoltWintersProcedureSpec) Kind() plan.ProcedureKind {
	return HoltWintersKind
}

func (s *HoltWintersProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(HoltWintersProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *HoltWintersProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createHoltWintersTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*HoltWintersProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewHoltWintersTransformation(d, cache, s)
	return t, d, nil
}

type holtWintersTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	smoothingFactor float64
	trendFactor     float64
}

func NewHoltWintersTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *HoltWintersProcedureSpec) *holtWintersTransformation {
	return &holtWintersTransformation{
		d:               d,
		cache:           cache,
		smoothingFactor: spec.SmoothingFactor,
		trendFactor:     spec.TrendFactor,
	}
}

func (t *holtWintersTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *holtWintersTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// TODO: Check that all columns are part of the key, except _value and _time.
	if t.smoothingFactor <= 0 || t.smoothingFactor >= 1 {
		return fmt.Errorf("invalid smoothing factor. Expected: 0 < sf < 1, got: %f", t.smoothingFactor)
	}
	if t.trendFactor <= 0 || t.trendFactor >= 1 {
		return fmt.Errorf("invalid trend factor. Expected: 0 < tf < 1, got: %f", t.trendFactor)
	}

	key := tbl.Key()
	builder, created := t.cache.TableBuilder(key)
	if !created {
		return fmt.Errorf("holtWinters found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableKeyCols(key, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, cols)
	if timeIdx < 0 {
		return fmt.Errorf("time column not found (cols: %v): %s", cols, execute.DefaultTimeColLabel)
	}
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found (cols: %v): %s", cols, execute.DefaultValueColLabel)
	}

	var (
		numVals   int
		s0, s1, b float64
		x, y      float64
	)
	err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(valIdx)
		times := cr.Times(timeIdx)
		for i := 0; i < cr.Len(); i++ {
			if !vs.IsValid(i) || !times.IsValid(i) {
				continue
			}

			v := vs.Value(i)

			switch numVals {
			// Set initial values.
			case 0:
				s1 = v
			case 1:
				b = v - s1
				fallthrough
				// Run the smoothing operation.
			default:
				// Scale the raw value against the smoothing factor.
				x = t.smoothingFactor * v

				// Scale the last smoothed value with the trend at this point.
				b = calcTrendValue(numVals-1, t.smoothingFactor, t.trendFactor, s0, s1, b)
				y = (1 - t.smoothingFactor) * (s1 + b)

				s0, s1 = s1, x+y
			}

			numVals++
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Omit output table if there are not at least two samples to compute a smoothing from.
	if numVals < 2 {
		return nil
	}

	outValIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat})
	if err != nil {
		return fmt.Errorf("error appending value column: %s", err)
	}

	if err := builder.AppendFloat(outValIdx, s1); err != nil {
		return err
	}
	return execute.AppendKeyValues(key, builder)
}

// Calculate the trend value at the given index i in raw data d.
// This is somewhat analogous to the slope of the trend at the given index.
// The argument "s" is the set of computed smoothed values.
// The argument "b" is the set of computed trend factors.
// The argument "d" is the set of raw input values.
func calcTrendValue(i int, sf, tf, s0, s1, b float64) float64 {
	if i == 0 {
		return b
	}

	x := tf * (s1 - s0)
	y := (1 - tf) * b

	return x + y
}

func (t *holtWintersTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *holtWintersTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *holtWintersTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
