package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/moving_average"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const RelativeStrengthIndexKind = "relativeStrengthIndex"

type RelativeStrengthIndexOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	relativeStrengthIndexSignature := semantic.LookupBuiltInType("universe", "relativeStrenthIndex")
	flux.RegisterPackageValue("universe", RelativeStrengthIndexKind, flux.MustValue(flux.FunctionValue(RelativeStrengthIndexKind, createRelativeStrengthIndexOpSpec, relativeStrengthIndexSignature)))
	flux.RegisterOpSpec(RelativeStrengthIndexKind, newRelativeStrengthIndexOp)
	plan.RegisterProcedureSpec(RelativeStrengthIndexKind, newRelativeStrengthIndexProcedure, RelativeStrengthIndexKind)
	execute.RegisterTransformation(RelativeStrengthIndexKind, createRelativeStrengthIndexTransformation)
}

func createRelativeStrengthIndexOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(RelativeStrengthIndexOpSpec)

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

func newRelativeStrengthIndexOp() flux.OperationSpec {
	return new(RelativeStrengthIndexOpSpec)
}

func (s *RelativeStrengthIndexOpSpec) Kind() flux.OperationKind {
	return RelativeStrengthIndexKind
}

type RelativeStrengthIndexProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newRelativeStrengthIndexProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*RelativeStrengthIndexOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &RelativeStrengthIndexProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *RelativeStrengthIndexProcedureSpec) Kind() plan.ProcedureKind {
	return RelativeStrengthIndexKind
}

func (s *RelativeStrengthIndexProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RelativeStrengthIndexProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *RelativeStrengthIndexProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createRelativeStrengthIndexTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*RelativeStrengthIndexProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewRelativeStrengthIndexTransformation(d, cache, s)
	return t, d, nil
}

type relativeStrengthIndexTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n       int64
	columns []string

	i       []int64
	emaUp   moving_average.ExponentialMovingAverage
	emaDown moving_average.ExponentialMovingAverage
	lastVal []interface{}
}

func NewRelativeStrengthIndexTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *RelativeStrengthIndexProcedureSpec) *relativeStrengthIndexTransformation {
	return &relativeStrengthIndexTransformation{
		d:       d,
		cache:   cache,
		n:       spec.N,
		columns: spec.Columns,
	}
}

func (t *relativeStrengthIndexTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *relativeStrengthIndexTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "moving average found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doRelativeStrengthIndex := make([]bool, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
					return errors.Newf(codes.FailedPrecondition, "cannot take relative strength index of column %s (type %s)", c.Label, c.Type.String())
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
			doRelativeStrengthIndex[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	t.i = make([]int64, len(cols))
	t.emaUp = *moving_average.New(int(t.n), len(cols))
	t.emaDown = *moving_average.New(int(t.n), len(cols))

	t.emaUp.Multiplier = float64(1) / float64(t.n)
	t.emaDown.Multiplier = float64(1) / float64(t.n)

	t.lastVal = make([]interface{}, len(cols))

	err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cols {
			var err error
			switch c.Type {
			case flux.TBool:
				// We can pass through values using one of the EMAs, since the same number of values have to be appended
				err = t.passThrough(moving_average.NewArrayContainer(cr.Bools(j)), builder, j)
			case flux.TInt:
				err = t.doNumeric(moving_average.NewArrayContainer(cr.Ints(j)), builder, j, doRelativeStrengthIndex[j])
			case flux.TUInt:
				err = t.doNumeric(moving_average.NewArrayContainer(cr.UInts(j)), builder, j, doRelativeStrengthIndex[j])
			case flux.TFloat:
				err = t.doNumeric(moving_average.NewArrayContainer(cr.Floats(j)), builder, j, doRelativeStrengthIndex[j])
			case flux.TString:
				err = t.passThrough(moving_average.NewArrayContainer(cr.Strings(j)), builder, j)
			case flux.TTime:
				err = t.passThroughTime(cr.Times(j), builder, j)
			}

			if err != nil {
				return err
			}
		}
		return nil
	})

	for j := range cols {
		if t.i[j] <= t.n {
			if doRelativeStrengthIndex[j] {
				// If we don't have enough values for a complete period, we compute the RSI using the averages of values encountered so far (no smoothing)
				rsi := float64(100) - (float64(100) / (float64(1) + t.emaUp.Value(j)/t.emaDown.Value(j)))
				if err := builder.AppendFloat(j, rsi); err != nil {
					return err
				}
			} else {
				if t.emaUp.LastVal(j) == nil {
					if err := builder.AppendNil(j); err != nil {
						return err
					}
				} else {
					if err := builder.AppendValue(j, values.New(t.emaUp.LastVal(j))); err != nil {
						return err
					}
				}
			}
		}
	}

	return err
}

func (t *relativeStrengthIndexTransformation) passThrough(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int) error {
	// We can use EMA's PassThrough, but we need to get rid of the first value
	slice := vs
	if t.i[bj] == 0 {
		if vs.Len() == 1 {
			t.i[bj] += int64(vs.Len())
			return nil
		} else {
			slice = vs.Slice(1, vs.Len())
			defer slice.Release()
		}
	}
	t.i[bj] += int64(vs.Len())
	return t.emaUp.PassThrough(slice, b, bj)
}

func (t *relativeStrengthIndexTransformation) doNumeric(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int, doRSI bool) error {
	if !doRSI {
		return t.passThrough(vs, b, bj)
	}

	j := 0

	for ; j < vs.Len(); j++ {
		if !vs.IsNull(j) {
			var up float64
			var down float64
			v := vs.Value(j).Float()
			if t.lastVal[bj] == nil {
				t.lastVal[bj] = float64(0)
			}
			if v > t.lastVal[bj].(float64) {
				up = v - t.lastVal[bj].(float64)
			} else if v < t.lastVal[bj].(float64) {
				down = t.lastVal[bj].(float64) - v
			}
			t.emaUp.Add(up, bj)
			t.emaDown.Add(down, bj)
			t.lastVal[bj] = v
		} else {
			// Skip nulls
			t.emaUp.AddNull(bj)
			t.emaDown.AddNull(bj)
		}
		if t.i[bj] >= t.n {
			if t.lastVal[bj] == nil {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				rsi := float64(100) - (float64(100) / (float64(1) + t.emaUp.Value(bj)/t.emaDown.Value(bj)))
				if err := b.AppendFloat(bj, rsi); err != nil {
					return err
				}
			}
		}
		t.i[bj]++
	}

	return nil
}

func (t *relativeStrengthIndexTransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	// We can use EMA's PassThroughTime, but we need to get rid of the first value
	slice := vs
	if t.i[bj] == 0 {
		if vs.Len() == 1 {
			t.i[bj] += int64(vs.Len())
			return nil
		} else {
			slice = arrow.IntSlice(vs, 1, vs.Len())
			defer slice.Release()
		}
	}
	t.i[bj] += int64(vs.Len())
	return t.emaUp.PassThroughTime(slice, b, bj)
}

func (t *relativeStrengthIndexTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *relativeStrengthIndexTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *relativeStrengthIndexTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
