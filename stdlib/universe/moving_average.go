package universe

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
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

	i      []int
	sum    []interface{}
	count  []int
	window [][]interface{}
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

	return tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			var err error
			switch c.Type {
			case flux.TBool:
				err = t.passThroughBool(cr.Bools(j), builder, j)
			case flux.TInt:
				err = t.doInt(cr.Ints(j), builder, j, doMovingAverage[j])
			case flux.TUInt:
				err = t.doUInt(cr.UInts(j), builder, j, doMovingAverage[j])
			case flux.TFloat:
				err = t.doFloat(cr.Floats(j), builder, j, doMovingAverage[j])
			case flux.TString:
				err = t.passThroughString(cr.Strings(j), builder, j)
			case flux.TTime:
				err = t.passThroughTime(cr.Times(j), builder, j)
			}

			if err != nil {
				return err
			}
		}
		return nil
	})
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

func (t *movingAverageTransformation) passThroughBool(vs *array.Boolean, b execute.TableBuilder, bj int) error {
	if int64(vs.Len()) < t.n {
		if vs.IsNull(vs.Len() - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendBool(bj, vs.Value(vs.Len()-1)); err != nil {
				return nil
			}
		}
		return nil
	} else {
		s := arrow.BoolSlice(vs, int(t.n-1), vs.Len())
		defer s.Release()
		return b.AppendBools(bj, s)
	}
}

func (t *movingAverageTransformation) doInt(vs *array.Int64, b execute.TableBuilder, bj int, doMovingAverage bool) error {
	if t.window[bj] == nil {
		t.window[bj] = make([]interface{}, t.n)
	}
	if t.sum[bj] == nil {
		t.sum[bj] = new(int64)
	}
	sumPointer := &t.sum[bj]
	sum := (*sumPointer).(*int64)

	j := 0

	if vs.IsValid(j) {
		*sum += vs.Value(j)
		t.count[bj]++
		t.window[bj][0] = vs.Value(j)
	} else {
		t.window[bj][0] = nil
	}
	j++
	t.i[bj]++

	l := vs.Len()
	for ; j < l; j++ {

		if !vs.IsNull(j) {
			t.count[bj]++
			t.window[bj][int64(t.i[bj])%t.n] = vs.Value(j)
		} else {
			t.window[bj][int64(t.i[bj])%t.n] = nil
		}
		*sum += vs.Value(j)

		if int64(t.i[bj]) < t.n-1 {
			t.i[bj]++
			continue
		}

		if !doMovingAverage {
			if vs.IsValid(j) {
				if err := b.AppendInt(bj, vs.Value(j)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else {
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
				*sum -= next.(int64)
				t.count[bj]--
			}

		}
		t.i[bj]++
	}
	if int64(t.i[bj]) < t.n-1 {
		if !doMovingAverage {
			if vs.IsNull(j - 1) {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendInt(bj, vs.Value(j-1)); err != nil {
					return err
				}
			}
		} else {
			average := float64(*sum) / float64(t.count[bj])
			if err := b.AppendFloat(bj, average); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *movingAverageTransformation) doUInt(vs *array.Uint64, b execute.TableBuilder, bj int, doMovingAverage bool) error {
	if t.window[bj] == nil {
		t.window[bj] = make([]interface{}, t.n)
	}
	if t.sum[bj] == nil {
		t.sum[bj] = new(uint64)
	}
	sumPointer := &t.sum[bj]
	sum := (*sumPointer).(*uint64)

	j := 0

	if vs.IsValid(j) {
		*sum += vs.Value(j)
		t.count[bj]++
		t.window[bj][0] = vs.Value(j)
	} else {
		t.window[bj][0] = nil
	}
	j++
	t.i[bj]++

	l := vs.Len()
	for ; j < l; j++ {

		if !vs.IsNull(j) {
			t.count[bj]++
			t.window[bj][int64(t.i[bj])%t.n] = vs.Value(j)
		} else {
			t.window[bj][int64(t.i[bj])%t.n] = nil
		}
		*sum += vs.Value(j)

		if int64(t.i[bj]) < t.n-1 {
			t.i[bj]++
			continue
		}

		if !doMovingAverage {
			if vs.IsValid(j) {
				if err := b.AppendUInt(bj, vs.Value(j)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else {
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
				*sum -= next.(uint64)
				t.count[bj]--
			}

		}
		t.i[bj]++
	}
	if int64(t.i[bj]) < t.n-1 {
		if !doMovingAverage {
			if vs.IsNull(j - 1) {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendUInt(bj, vs.Value(j-1)); err != nil {
					return err
				}
			}
		} else {
			average := float64(*sum) / float64(t.count[bj])
			if err := b.AppendFloat(bj, average); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *movingAverageTransformation) doFloat(vs *array.Float64, b execute.TableBuilder, bj int, doMovingAverage bool) error {
	if t.window[bj] == nil {
		t.window[bj] = make([]interface{}, t.n)
	}
	if t.sum[bj] == nil {
		t.sum[bj] = new(float64)
	}
	sumPointer := &t.sum[bj]
	sum := (*sumPointer).(*float64)

	j := 0

	if vs.IsValid(j) {
		*sum += vs.Value(j)
		t.count[bj]++
		t.window[bj][0] = vs.Value(j)
	} else {
		t.window[bj][0] = nil
	}
	j++
	t.i[bj]++

	l := vs.Len()
	for ; j < l; j++ {

		if !vs.IsNull(j) {
			t.count[bj]++
			t.window[bj][int64(t.i[bj])%t.n] = vs.Value(j)
		} else {
			t.window[bj][int64(t.i[bj])%t.n] = nil
		}
		*sum += vs.Value(j)

		if int64(t.i[bj]) < t.n-1 {
			t.i[bj]++
			continue
		}

		if !doMovingAverage {
			if vs.IsValid(j) {
				if err := b.AppendFloat(bj, vs.Value(j)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else {
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

		}
		t.i[bj]++
	}
	if int64(t.i[bj]) < t.n-1 {
		if !doMovingAverage {
			if vs.IsNull(j - 1) {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				if err := b.AppendFloat(bj, vs.Value(j-1)); err != nil {
					return err
				}
			}
		} else {
			average := float64(*sum) / float64(t.count[bj])
			if err := b.AppendFloat(bj, average); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *movingAverageTransformation) passThroughString(vs *array.Binary, b execute.TableBuilder, bj int) error {
	if int64(vs.Len()) < t.n {
		if vs.IsNull(vs.Len() - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendString(bj, string(vs.Value(vs.Len()-1))); err != nil {
				return nil
			}
		}
		return nil
	} else {
		s := arrow.StringSlice(vs, int(t.n-1), vs.Len())
		defer s.Release()
		return b.AppendStrings(bj, s)
	}
}

func (t *movingAverageTransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	if int64(vs.Len()) < t.n {
		if vs.IsNull(vs.Len() - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendTime(bj, execute.Time(vs.Value(vs.Len()-1))); err != nil {
				return nil
			}
		}
		return nil
	} else {
		s := arrow.IntSlice(vs, int(t.n-1), vs.Len())
		defer s.Release()
		return b.AppendTimes(bj, s)
	}
}
