package universe

import (
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/moving_average"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const TripleExponentialDerivativeKind = "tripleExponentialDerivative"

type TripleExponentialDerivativeOpSpec struct {
	N int64 `json:"n"`
}

func init() {
	tripleExponentialDerivativeSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n": semantic.Int,
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", TripleExponentialDerivativeKind, flux.FunctionValue(TripleExponentialDerivativeKind, createTripleExponentialDerivativeOpSpec, tripleExponentialDerivativeSignature))
	flux.RegisterOpSpec(TripleExponentialDerivativeKind, newTripleExponentialDerivativeOp)
	plan.RegisterProcedureSpec(TripleExponentialDerivativeKind, newTripleExponentialDerivativeProcedure, TripleExponentialDerivativeKind)
	execute.RegisterTransformation(TripleExponentialDerivativeKind, createTripleExponentialDerivativeTransformation)
}

func createTripleExponentialDerivativeOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(TripleExponentialDerivativeOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else if n <= 0 {
		return nil, errors.Newf(codes.Internal, "cannot take triple exponential derivative with a period of %v (must be greater than 0)", n)
	} else {
		spec.N = n
	}

	return spec, nil
}

func newTripleExponentialDerivativeOp() flux.OperationSpec {
	return new(TripleExponentialDerivativeOpSpec)
}

func (s *TripleExponentialDerivativeOpSpec) Kind() flux.OperationKind {
	return TripleExponentialDerivativeKind
}

type TripleExponentialDerivativeProcedureSpec struct {
	plan.DefaultCost
	N int64 `json:"n"`
}

func newTripleExponentialDerivativeProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TripleExponentialDerivativeOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &TripleExponentialDerivativeProcedureSpec{
		N: spec.N,
	}, nil
}

func (s *TripleExponentialDerivativeProcedureSpec) Kind() plan.ProcedureKind {
	return TripleExponentialDerivativeKind
}

func (s *TripleExponentialDerivativeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(TripleExponentialDerivativeProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *TripleExponentialDerivativeProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createTripleExponentialDerivativeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*TripleExponentialDerivativeProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	alloc := a.Allocator()
	cache := execute.NewTableBuilderCache(alloc)

	d := execute.NewDataset(id, mode, cache)
	t := NewTripleExponentialDerivativeTransformation(d, cache, alloc, s)
	return t, d, nil
}

type tripleExponentialDerivativeTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	alloc *memory.Allocator

	i                []int
	lastVal          []interface{}
	ema1, ema2, ema3 *moving_average.ExponentialMovingAverage

	n int64
}

func NewTripleExponentialDerivativeTransformation(d execute.Dataset, cache execute.TableBuilderCache, alloc *memory.Allocator, spec *TripleExponentialDerivativeProcedureSpec) *tripleExponentialDerivativeTransformation {
	return &tripleExponentialDerivativeTransformation{
		d:     d,
		cache: cache,
		alloc: alloc,

		n: spec.N,
	}
}

func (t *tripleExponentialDerivativeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *tripleExponentialDerivativeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "triple exponential derivative found duplicate table with key: %v", tbl.Key())
	}

	cols := tbl.Cols()
	valueIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valueIdx == -1 {
		return errors.New(codes.FailedPrecondition, "cannot find _value column")
	}
	valueCol := cols[valueIdx]
	if valueCol.Type != flux.TInt && valueCol.Type != flux.TUInt && valueCol.Type != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "cannot take exponential moving average of column %s (type %s)", valueCol.Label, valueCol.Type.String())
	}
	for j, c := range cols {
		if j == valueIdx {
			_, err := builder.AddCol(flux.ColMeta{Label: c.Label, Type: flux.TFloat})
			if err != nil {
				return err
			}
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	// Keeps track of current position and last value looked at for columns that are passed through
	// Faster than calling ema.PassThrough three times
	t.i = make([]int, len(cols))
	t.lastVal = make([]interface{}, len(cols))

	t.ema1 = moving_average.New(int(t.n), len(cols))
	t.ema2 = moving_average.New(int(t.n), len(cols))
	t.ema3 = moving_average.New(int(t.n), len(cols))

	if err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			isValueCol := valueIdx == j
			var err error
			switch c.Type {
			case flux.TBool:
				// We can pass through values using one of the EMAs, since the same number of values have to be appended
				err = t.passThrough(moving_average.NewArrayContainer(cr.Bools(j)), builder, j)
			case flux.TInt:
				err = t.doFirstEMA(moving_average.NewArrayContainer(cr.Ints(j)), builder, j, isValueCol)
			case flux.TUInt:
				err = t.doFirstEMA(moving_average.NewArrayContainer(cr.UInts(j)), builder, j, isValueCol)
			case flux.TFloat:
				err = t.doFirstEMA(moving_average.NewArrayContainer(cr.Floats(j)), builder, j, isValueCol)
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
	}); err != nil {
		return err
	}

	for j := range cols {
		if j == valueIdx {
			if err := t.doRest(builder, j); err != nil {
				return err
			}
		} else {
			// Check for any incomplete TRIXs, append the last value
			if int64(t.i[j]) < 3*(t.n-1) {
				val := values.New(t.lastVal[j])
				if err := builder.AppendValue(j, val); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *tripleExponentialDerivativeTransformation) passThrough(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int) error {

	// Skip all values which Triple Exponential Derivative won't output
	//     - math.Min decides whether to skip all the values in the slice
	//     - math.Max leaves the index unchanged if we want to pass through all the values in the slice (since math.Min would return a negative)
	j := int(math.Max(0, math.Min(float64(vs.Len()), float64(3*(t.n-1))-float64(t.i[bj])+1)))
	t.i[bj] += int(math.Max(0, math.Min(float64(vs.Len()), float64(3*(t.n-1))-float64(t.i[bj])+1)))

	if j < vs.Len() {
		slice := vs.Slice(j, vs.Len()).Array()
		defer slice.Release()

		switch s := slice.(type) {
		case *array.Boolean:
			if err := b.AppendBools(bj, s); err != nil {
				return err
			}
		case *array.Binary:
			if err := b.AppendStrings(bj, s); err != nil {
				return err
			}
		}
	}

	t.lastVal[bj] = vs.Value(vs.Len() - 1)

	return nil
}

func (t *tripleExponentialDerivativeTransformation) passThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {

	// Skip all values which Triple Exponential Derivative won't output
	//     - math.Min decides whether to skip all the values in the slice
	//     - math.Max leaves the index unchanged if we want to pass through all the values in the slice (since math.Min would return a negative)
	j := int(math.Max(0, math.Min(float64(vs.Len()), float64(3*(t.n-1))-float64(t.i[bj])+1)))
	t.i[bj] += int(math.Max(0, math.Min(float64(vs.Len()), float64(3*(t.n-1))-float64(t.i[bj])+1)))

	if j < vs.Len() {
		slice := arrow.IntSlice(vs, j, vs.Len())
		defer slice.Release()
		if err := b.AppendTimes(bj, slice); err != nil {
			return err
		}
	}

	t.lastVal[bj] = execute.Time(vs.Value(vs.Len() - 1))

	return nil
}

func (t *tripleExponentialDerivativeTransformation) doFirstEMA(vs *moving_average.ArrayContainer, b execute.TableBuilder, bj int, doDEMA bool) error {
	// if !doDEMA, append the last 2n - 1 values
	if !doDEMA {
		return t.passThrough(vs, b, bj)
	}

	return t.ema1.DoNumeric(vs, b, bj, doDEMA, false)
}

func (t *tripleExponentialDerivativeTransformation) doRest(b execute.TableBuilder, bj int) error {
	firstEMA := t.ema1.GetEMA(bj)

	// Convert firstEMA to *array.Float64
	arr1 := arrayToFloatArrow(firstEMA, t.alloc)
	defer arr1.Release()

	// Do the second EMA
	if err := t.ema2.DoNumeric(moving_average.NewArrayContainer(arr1), b, bj, true, false); err != nil {
		return err
	}

	// Get the second EMA
	secondEMA := t.ema2.GetEMA(bj)
	arr2 := arrayToFloatArrow(secondEMA, t.alloc)
	defer arr2.Release()

	// Get the third EMA
	if err := t.ema3.DoNumeric(moving_average.NewArrayContainer(arr2), b, bj, true, false); err != nil {
		return err
	}

	thirdEMA := t.ema3.GetEMA(bj)

	// If there weren't enough values to take 2 EMAs, append a null
	if thirdEMA == nil {
		if err := b.AppendNil(bj); err != nil {
			return nil
		}
	} else {
		for i := 1; i < len(thirdEMA); i++ {
			if thirdEMA[i] == nil || thirdEMA[i-1] == nil {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				val := ((thirdEMA[i].(float64) / thirdEMA[i-1].(float64)) - 1) * 100
				if err := b.AppendFloat(bj, val); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *tripleExponentialDerivativeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *tripleExponentialDerivativeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *tripleExponentialDerivativeTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func arrayToFloatArrow(a []interface{}, alloc *memory.Allocator) *array.Float64 {
	bld := arrow.NewFloatBuilder(alloc)
	defer bld.Release()

	for _, val := range a {
		if val != nil {
			bld.Append(val.(float64))
		} else {
			bld.AppendNull()
		}
	}

	return bld.NewFloat64Array()
}
