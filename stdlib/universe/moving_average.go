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

const MovingAverageKind = "movingAverage"

type MovingAverageOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	movingAverageSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":       semantic.Int,
			"columns": semantic.NewArrayPolyType(semantic.String),
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", MovingAverageKind, flux.FunctionValue(MovingAverageKind, createMovingAverageOpSpec, movingAverageSignature))
	flux.RegisterOpSpec(MovingAverageKind, newMovingAverageOp)
	plan.RegisterProcedureSpec(MovingAverageKind, newMovingAverageProcedure, MovingAverageKind)
	execute.RegisterTransformation(MovingAverageKind, createMovingAverageTransformation)
}

func createMovingAverageOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MovingAverageOpSpec)

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

func newMovingAverageOp() flux.OperationSpec {
	return new(MovingAverageOpSpec)
}

func (s *MovingAverageOpSpec) Kind() flux.OperationKind {
	return MovingAverageKind
}

type MovingAverageProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newMovingAverageProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MovingAverageOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &MovingAverageProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *MovingAverageProcedureSpec) Kind() plan.ProcedureKind {
	return MovingAverageKind
}

func (s *MovingAverageProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MovingAverageProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *MovingAverageProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createMovingAverageTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MovingAverageProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewMovingAverageTransformation(d, cache, s)
	return t, d, nil
}

type movingAverageTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n       int64
	columns []string

	i             []int
	sum           []interface{}
	count         []int
	window        [][]interface{}
	periodReached []bool
	lastVal       []interface{}
	notEmpty      []bool
}

func NewMovingAverageTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *MovingAverageProcedureSpec) *movingAverageTransformation {
	return &movingAverageTransformation{
		d:       d,
		cache:   cache,
		n:       spec.N,
		columns: spec.Columns,
	}
}

func (t *movingAverageTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *movingAverageTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("moving average found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doMovingAverage := make([]bool, len(cols))
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
			doMovingAverage[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	t.i = make([]int, len(cols))
	t.sum = make([]interface{}, len(cols))
	t.count = make([]int, len(cols))
	t.window = make([][]interface{}, len(cols))
	t.periodReached = make([]bool, len(cols))
	t.lastVal = make([]interface{}, len(cols))
	t.notEmpty = make([]bool, len(cols))

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
				err = t.doNumeric(&arrayContainer{cr.Ints(j)}, builder, j, doMovingAverage[j])
			case flux.TUInt:
				err = t.doNumeric(&arrayContainer{cr.UInts(j)}, builder, j, doMovingAverage[j])
			case flux.TFloat:
				err = t.doNumeric(&arrayContainer{cr.Floats(j)}, builder, j, doMovingAverage[j])
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
		if !t.periodReached[j] && t.notEmpty[j] {
			if !doMovingAverage[j] {
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
				average := *(t.sum[j].(*float64)) / float64(t.count[j])
				if err := builder.AppendFloat(j, average); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *movingAverageTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *movingAverageTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *movingAverageTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *movingAverageTransformation) passThrough(vs *arrayContainer, b execute.TableBuilder, bj int) error {
	t.notEmpty[bj] = true
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

func (t *movingAverageTransformation) doNumeric(vs *arrayContainer, b execute.TableBuilder, bj int, doMovingAverage bool) error {
	if !doMovingAverage {
		return t.passThrough(vs, b, bj)
	}

	t.notEmpty[bj] = true

	if t.window[bj] == nil {
		t.window[bj] = make([]interface{}, t.n)
	}
	if t.sum[bj] == nil {
		t.sum[bj] = new(float64)
	}
	sumPointer := &t.sum[bj]
	sum := (*sumPointer).(*float64)

	j := 0

	for ; int64(t.i[bj]) < t.n-1 && j < vs.Len(); t.i[bj]++ {
		if vs.IsValid(j) {
			*sum += vs.Value(j).Float()
			t.count[bj]++
			t.window[bj][int64(t.i[bj])%t.n] = vs.Value(j).Float()
		} else {
			t.window[bj][int64(t.i[bj])%t.n] = nil
		}
		j++
	}

	for ; j < vs.Len(); j++ {
		if vs.IsValid(j) {
			*sum += vs.Value(j).Float()
			t.count[bj]++
			t.window[bj][int64(t.i[bj])%t.n] = vs.Value(j).Float()
		} else {
			t.window[bj][int64(t.i[bj])%t.n] = nil
		}

		if int64(t.i[bj]) == t.n && !t.periodReached[bj] {
			t.periodReached[bj] = true
		}

		average := 0.0
		if t.count[bj] != 0 {
			average = float64(*sum) / float64(t.count[bj])
			if err := b.AppendFloat(bj, average); err != nil {
				return err
			}
		} else {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		}

		next := t.window[bj][int64(t.i[bj]+1)%t.n]
		if next != nil {
			*sum -= next.(float64)
			t.count[bj]--
		}

		t.i[bj]++
	}

	return nil
}

func (t *movingAverageTransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	t.notEmpty[bj] = true
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
