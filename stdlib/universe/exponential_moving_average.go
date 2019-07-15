package universe

import (
	"fmt"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
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

	n       int64
	columns []string

	i             []int
	count         []float64
	pValue        []float64
	periodReached []bool
	lastVal       []interface{}
}

func NewExponentialMovingAverageTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ExponentialMovingAverageProcedureSpec) *exponentialMovingAverageTransformation {
	return &exponentialMovingAverageTransformation{
		d:       d,
		cache:   cache,
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

	t.i = make([]int, len(cols))
	t.count = make([]float64, len(cols))
	t.pValue = make([]float64, len(cols))
	t.periodReached = make([]bool, len(cols))
	t.lastVal = make([]interface{}, len(cols))

	err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			var err error
			switch c.Type {
			case flux.TBool:
				err = t.passThrough(&arrayContainer{cr.Bools(j)}, builder, j)
			case flux.TInt:
				err = t.doNumeric(&arrayContainer{cr.Ints(j)}, builder, j, doExponentialMovingAverage[j])
			case flux.TUInt:
				err = t.doNumeric(&arrayContainer{cr.UInts(j)}, builder, j, doExponentialMovingAverage[j])
			case flux.TFloat:
				err = t.doNumeric(&arrayContainer{cr.Floats(j)}, builder, j, doExponentialMovingAverage[j])
			case flux.TString:
				err = t.passThrough(&arrayContainer{cr.Strings(j)}, builder, j)
			case flux.TTime:
				err = t.passThroughTime(cr.Times(j), builder, j)
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

	for j := range tbl.Cols() {
		if !t.periodReached[j] {
			if !doExponentialMovingAverage[j] {
				if t.lastVal[j] == nil {
					if err := builder.AppendNil(j); err != nil {
						return err
					}
				} else {
					if err := builder.AppendValue(j, values.New(t.lastVal[j])); err != nil {
						return err
					}
				}
			} else {
				if t.count[j] != 0 {
					average := t.pValue[j] / t.count[j]
					if err := builder.AppendFloat(j, average); err != nil {
						return err
					}
				} else {
					if err := builder.AppendNil(j); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
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

func (t *exponentialMovingAverageTransformation) passThrough(vs *arrayContainer, b execute.TableBuilder, bj int) error {
	j := 0

	for ; int64(t.i[bj]) < t.n && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			t.lastVal[bj] = nil
		} else {
			t.lastVal[bj] = vs.OrigValue(j)
		}
		j++
	}

	if int64(t.i[bj]) == t.n && !t.periodReached[bj] {
		if vs.IsNull(j - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendValue(bj, values.New(vs.OrigValue(j-1))); err != nil {
				return err
			}
		}
		t.periodReached[bj] = true
	}

	for ; int64(t.i[bj]) >= t.n && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendValue(bj, values.New(vs.OrigValue(j))); err != nil {
				return err
			}
		}
		j++
	}
	return nil
}

func (t *exponentialMovingAverageTransformation) doNumeric(vs *arrayContainer, b execute.TableBuilder, bj int, doExponentialMovingAverage bool) error {
	if !doExponentialMovingAverage {
		return t.passThrough(vs, b, bj)
	}

	mult := 2.0 / (float64(t.n) + 1)
	j := 0

	if t.i[bj] == 0 {
		if vs.IsValid(j) {
			t.pValue[bj] = vs.Value(j).Float()
			t.count[bj]++
			t.lastVal[bj] = vs.OrigValue(j)
		} else {
			t.lastVal[bj] = nil
		}

		t.i[bj]++
		j++
	}

	for ; int64(t.i[bj]) < t.n && j < vs.Len(); t.i[bj]++ {
		if !vs.IsNull(j) {
			t.pValue[bj] += vs.Value(j).Float()
			t.count[bj]++
			t.lastVal[bj] = vs.OrigValue(j)
		} else {
			t.lastVal[bj] = nil
		}
		j++
	}

	if int64(t.i[bj]) == t.n && !t.periodReached[bj] {
		if t.count[bj] != 0 {
			t.pValue[bj] = t.pValue[bj] / t.count[bj]
			if err := b.AppendFloat(bj, t.pValue[bj]); err != nil {
				return err
			}
		} else {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		}
		t.periodReached[bj] = true
	}

	l := vs.Len()
	for ; j < l; j++ {
		if vs.IsNull(j) {
			if t.count[bj] == 0 {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendFloat(bj, t.pValue[bj]); err != nil {
					return err
				}
			}
		} else {
			cValue := vs.Value(j).Float()
			var ema float64
			if t.count[bj] == 0 {
				ema = cValue
				t.count[bj]++
			} else {
				ema = (cValue * mult) + (t.pValue[bj] * (1.0 - mult))
			}
			if err := b.AppendFloat(bj, ema); err != nil {
				return err
			}
			t.pValue[bj] = ema
		}
		t.i[bj]++
	}
	return nil
}

func (t *exponentialMovingAverageTransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	j := 0

	for ; int64(t.i[bj]) < t.n && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			t.lastVal[bj] = nil
		} else {
			t.lastVal[bj] = execute.Time(vs.Value(j))
		}
		j++
	}

	if int64(t.i[bj]) == t.n && !t.periodReached[bj] {
		if vs.IsNull(j - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendTime(bj, execute.Time(vs.Value(j-1))); err != nil {
				return err
			}
		}
		t.periodReached[bj] = true
	}

	for ; int64(t.i[bj]) >= t.n && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendTime(bj, execute.Time(vs.Value(j))); err != nil {
				return err
			}
		}
		j++
	}
	return nil
}
