package universe

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe/moving_average"
)

const ExponentialMovingAverageKind = "exponentialMovingAverage"

type ExponentialMovingAverageOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	exponentialMovingAverageSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":       semantic.Int,
			"columns": semantic.NewArrayPolyType(semantic.String),
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", ExponentialMovingAverageKind, flux.FunctionValue(ExponentialMovingAverageKind, createExponentialMovingAverageOpSpec, exponentialMovingAverageSignature))
	flux.RegisterOpSpec(ExponentialMovingAverageKind, newExponentialMovingAverageOp)
	plan.RegisterProcedureSpec(ExponentialMovingAverageKind, newExponentialMovingAverageProcedure, ExponentialMovingAverageKind)
	execute.RegisterTransformation(ExponentialMovingAverageKind, createExponentialMovingAverageTransformation)
}

func createExponentialMovingAverageOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ExponentialMovingAverageOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}

	if cols, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	} else {
		spec.Columns = []string{execute.DefaultValueColLabel}
	}

	return spec, nil
}

func newExponentialMovingAverageOp() flux.OperationSpec {
	return new(ExponentialMovingAverageOpSpec)
}

func (s *ExponentialMovingAverageOpSpec) Kind() flux.OperationKind {
	return ExponentialMovingAverageKind
}

type ExponentialMovingAverageProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newExponentialMovingAverageProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ExponentialMovingAverageOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ExponentialMovingAverageProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *ExponentialMovingAverageProcedureSpec) Kind() plan.ProcedureKind {
	return ExponentialMovingAverageKind
}

func (s *ExponentialMovingAverageProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ExponentialMovingAverageProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ExponentialMovingAverageProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createExponentialMovingAverageTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ExponentialMovingAverageProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewExponentialMovingAverageTransformation(d, cache, s)
	return t, d, nil
}

type exponentialMovingAverageTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	ema *moving_average.ExponentialMovingAverage

	n       int64
	columns []string
}

func NewExponentialMovingAverageTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ExponentialMovingAverageProcedureSpec) *exponentialMovingAverageTransformation {
	return &exponentialMovingAverageTransformation{
		d:     d,
		cache: cache,

		n:       spec.N,
		columns: spec.Columns,
	}
}

func (t *exponentialMovingAverageTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *exponentialMovingAverageTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("moving average found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doExponentialMovingAverage := make([]bool, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
					return fmt.Errorf("cannot take moving average of column %s (type %s)", c.Label, c.Type.String())
				}
				found = true
				break
			}
		}

		if found {
			mac := c
			mac.Type = flux.TFloat
			_, err := builder.AddCol(mac)
			if err != nil {
				return err
			}
			doExponentialMovingAverage[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	t.ema = moving_average.New(int(t.n), len(cols))

	err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			// use ArrayContainer to avoid having a different function for each type, where almost all the code would be the same
			var err error
			switch c.Type {
			case flux.TBool:
				err = t.ema.PassThrough(&moving_average.ArrayContainer{Array: cr.Bools(j)}, builder, j)
			case flux.TInt:
				err = t.ema.DoNumeric(&moving_average.ArrayContainer{Array: cr.Ints(j)}, builder, j, doExponentialMovingAverage[j], true)
			case flux.TUInt:
				err = t.ema.DoNumeric(&moving_average.ArrayContainer{Array: cr.UInts(j)}, builder, j, doExponentialMovingAverage[j], true)
			case flux.TFloat:
				err = t.ema.DoNumeric(&moving_average.ArrayContainer{Array: cr.Floats(j)}, builder, j, doExponentialMovingAverage[j], true)
			case flux.TString:
				err = t.ema.PassThrough(&moving_average.ArrayContainer{Array: cr.Strings(j)}, builder, j)
			case flux.TTime:
				err = t.ema.PassThroughTime(cr.Times(j), builder, j)
			}

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return t.ema.Finish(tbl.Cols(), builder, doExponentialMovingAverage)
}

func (t *exponentialMovingAverageTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *exponentialMovingAverageTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *exponentialMovingAverageTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
