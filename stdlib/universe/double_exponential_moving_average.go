package universe

import (
	"fmt"
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe/moving_average"
	"github.com/influxdata/flux/values"
)

const DEMAKind = "doubleExponentialMovingAverage"

type DEMAOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	DEMASignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":       semantic.Int,
			"columns": semantic.NewArrayPolyType(semantic.String),
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", DEMAKind, flux.FunctionValue(DEMAKind, createDEMAOpSpec, DEMASignature))
	flux.RegisterOpSpec(DEMAKind, newDEMAOp)
	plan.RegisterProcedureSpec(DEMAKind, newDEMAProcedure, DEMAKind)
	execute.RegisterTransformation(DEMAKind, createDEMATransformation)
}

func createDEMAOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DEMAOpSpec)

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

func newDEMAOp() flux.OperationSpec {
	return new(DEMAOpSpec)
}

func (s *DEMAOpSpec) Kind() flux.OperationKind {
	return DEMAKind
}

type DEMAProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newDEMAProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DEMAOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &DEMAProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *DEMAProcedureSpec) Kind() plan.ProcedureKind {
	return DEMAKind
}

func (s *DEMAProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DEMAProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *DEMAProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createDEMATransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DEMAProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	alloc := a.Allocator()
	cache := execute.NewTableBuilderCache(alloc)

	d := execute.NewDataset(id, mode, cache)
	t := NewDEMATransformation(d, cache, alloc, s)
	return t, d, nil
}

type DEMATransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	alloc *memory.Allocator

	n       int64
	columns []string

	ema1    *moving_average.ExponentialMovingAverage
	ema2    *moving_average.ExponentialMovingAverage
	i       []int
	lastVal []interface{}
}

func NewDEMATransformation(d execute.Dataset, cache execute.TableBuilderCache, alloc *memory.Allocator, spec *DEMAProcedureSpec) *DEMATransformation {
	return &DEMATransformation{
		d:     d,
		cache: cache,
		alloc: alloc,

		n:       spec.N,
		columns: spec.Columns,
	}
}

func (t *DEMATransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *DEMATransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("double exponential moving average found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doDEMA := make([]bool, len(cols))
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
			doDEMA[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	t.ema1 = moving_average.New(int(t.n), len(cols))
	t.ema2 = moving_average.New(int(t.n), len(cols))

	// Keeps track of current position and last value looked at for columns that are passed through
	// Faster than calling ema.PassThrough twice
	t.i = make([]int, len(cols))
	t.lastVal = make([]interface{}, len(cols))

	err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		// Start by doing the first EMA - we can't do both at once because of chunking
		for j, c := range cols {

			var err error
			switch c.Type {
			case flux.TBool:
				err = t.passThrough(&moving_average.ArrayContainer{Array: cr.Bools(j)}, builder, j)
			case flux.TInt:
				err = t.doFirstEMA(&moving_average.ArrayContainer{Array: cr.Ints(j)}, builder, j, doDEMA[j])
			case flux.TUInt:
				err = t.doFirstEMA(&moving_average.ArrayContainer{Array: cr.UInts(j)}, builder, j, doDEMA[j])
			case flux.TFloat:
				err = t.doFirstEMA(&moving_average.ArrayContainer{Array: cr.Floats(j)}, builder, j, doDEMA[j])
			case flux.TString:
				err = t.passThrough(&moving_average.ArrayContainer{Array: cr.Strings(j)}, builder, j)
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

	for j := range cols {
		if doDEMA[j] {
			// Do second EMA and calculate the DEMA
			if err := t.doSecondEMA(builder, j); err != nil {
				return err
			}
		} else {
			// Check for any incomplete DEMAs, append the last value
			if int64(t.i[j]) < 2*t.n-1 {
				val := values.New(t.lastVal[j])
				if err := builder.AppendValue(j, val); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *DEMATransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *DEMATransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *DEMATransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *DEMATransformation) passThrough(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int) error {
	// This is faster than passing through EMA twice
	j := 0

	for ; int64(t.i[bj]) < 2*t.n-1 && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			t.lastVal[bj] = nil
		} else {
			t.lastVal[bj] = vs.OrigValue(j)
		}
		if int64(t.i[bj]) == 2*t.n-2 {
			if vs.IsNull(j) {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendValue(bj, values.New(vs.OrigValue(j))); err != nil {
					return err
				}
			}
		}
		j++
	}

	for ; int64(t.i[bj]) >= 2*t.n-1 && j < vs.Len(); t.i[bj]++ {
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

func (t *DEMATransformation) doFirstEMA(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int, doDEMA bool) error {
	// if !doDEMA, append the last 2n - 1 values
	if !doDEMA {
		return t.passThrough(vs, b, bj)
	}

	return t.ema1.DoNumeric(vs, b, bj, doDEMA, false)
}

func (t *DEMATransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	j := 0

	for ; int64(t.i[bj]) < 2*t.n-1 && j < vs.Len(); t.i[bj]++ {
		if vs.IsNull(j) {
			t.lastVal[bj] = nil
		} else {
			t.lastVal[bj] = execute.Time(vs.Value(j))
		}
		if int64(t.i[bj]) == 2*t.n-2 {
			if vs.IsNull(j) {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendTime(bj, execute.Time(vs.Value(j))); err != nil {
					return err
				}
			}
		}

		j++
	}

	for ; int64(t.i[bj]) >= 2*t.n-1 && j < vs.Len(); t.i[bj]++ {
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

func (t *DEMATransformation) doSecondEMA(b execute.TableBuilder, bj int) error {
	// Get the first EMA
	firstEMA := t.ema1.GetEMA(bj)

	// Convert firstEMA to *array.Float64
	bld := arrow.NewFloatBuilder(t.alloc)
	defer bld.Release()

	for _, val := range firstEMA {
		if val != nil {
			bld.Append(val.(float64))
		} else {
			bld.AppendNull()
		}
	}

	arr := bld.NewFloat64Array()
	defer arr.Release()

	// Get the second EMA
	if err := t.ema2.DoNumeric(&moving_average.ArrayContainer{Array: arr}, b, bj, true, false); err != nil {
		return err
	}

	secondEMA := t.ema2.GetEMA(bj)

	// If there weren't enough values to take 2 EMAs, append a math.NaN
	if secondEMA == nil {
		if err := b.AppendFloat(bj, math.NaN()); err != nil {
			return nil
		}
	} else {
		for i := range firstEMA[(t.n - 1):] {
			if secondEMA[i] == nil {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				val := 2*firstEMA[t.n-1+int64(i)].(float64) - secondEMA[i].(float64)
				if err := b.AppendFloat(bj, val); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
