package universe

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	fluxarrow "github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	fluxmemory "github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe/holt_winters"
	"github.com/influxdata/flux/values"
)

const HoltWintersKind = "holtWinters"

type HoltWintersOpSpec struct {
	WithFit    bool          `json:"with_fit"`
	Column     string        `json:"column"`
	TimeColumn string        `json:"time_column"`
	N          int64         `json:"n"`
	S          int64         `json:"s"`
	Interval   flux.Duration `json:"interval"`
}

func init() {
	hwSignature := semantic.LookupBuiltInType("univser", "holtWinter")
	flux.RegisterPackageValue("universe", HoltWintersKind, flux.MustValue(flux.FunctionValue(HoltWintersKind, createHoltWintersOpSpec, hwSignature)))
	flux.RegisterOpSpec(HoltWintersKind, newHoltWintersOp)
	plan.RegisterProcedureSpec(HoltWintersKind, newHoltWintersProcedure, HoltWintersKind)
	execute.RegisterTransformation(HoltWintersKind, createHoltWintersTransformation)
}

func createHoltWintersOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	spec := new(HoltWintersOpSpec)
	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}
	if i, err := args.GetRequiredDuration("interval"); err != nil {
		return nil, err
	} else {
		spec.Interval = i
	}
	if wf, ok, err := args.GetBool("withFit"); err != nil {
		return nil, err
	} else if ok {
		spec.WithFit = wf
	}
	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}
	if col, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = col
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}
	if s, ok, err := args.GetInt("seasonality"); err != nil {
		return nil, err
	} else if ok {
		spec.S = s
	} else {
		spec.S = 0
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
	WithFit    bool
	Column     string
	TimeColumn string
	N          int64
	S          int64
	Interval   flux.Duration
}

func newHoltWintersProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*HoltWintersOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &HoltWintersProcedureSpec{
		WithFit:    spec.WithFit,
		Column:     spec.Column,
		TimeColumn: spec.TimeColumn,
		N:          spec.N,
		S:          spec.S,
		Interval:   spec.Interval,
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
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewHoltWintersTransformation(d, cache, a.Allocator(), s)
	return t, d, nil
}

type holtWintersTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	alloc *fluxmemory.Allocator

	withFit    bool
	column     string
	timeColumn string
	n          int64
	s          int64
	interval   values.Duration
}

func NewHoltWintersTransformation(d execute.Dataset, cache execute.TableBuilderCache, alloc *fluxmemory.Allocator, spec *HoltWintersProcedureSpec) *holtWintersTransformation {
	return &holtWintersTransformation{
		d:          d,
		cache:      cache,
		alloc:      alloc,
		withFit:    spec.WithFit,
		column:     spec.Column,
		timeColumn: spec.TimeColumn,
		n:          spec.N,
		s:          spec.S,
		interval:   values.Duration(spec.Interval),
	}
}

func (hwt *holtWintersTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Sanity checks.
	builder, created := hwt.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "holtWinters found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	timeIdx := execute.ColIdx(hwt.timeColumn, cols)
	if timeIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "cannot find time column %s", hwt.timeColumn)
	}
	colIdx := execute.ColIdx(hwt.column, cols)
	if colIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "cannot find column %s", hwt.column)
	}
	typ := cols[colIdx].Type
	if typ != flux.TInt &&
		typ != flux.TUInt &&
		typ != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "holtWinters can work only on numerical types, got %s", typ.String())
	}

	// Building schema.
	if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}
	newTimeIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultTimeColLabel,
		Type:  flux.TTime,
	})
	if err != nil {
		return err
	}
	newValueIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}

	// Cleaning data for HoltWinters input.
	vs, start, stop, err := hwt.getCleanData(tbl, colIdx, timeIdx)
	if err != nil {
		return err
	}

	// Holt Winters.
	hw := holt_winters.New(int(hwt.n), int(hwt.s), hwt.withFit, fluxarrow.NewAllocator(hwt.alloc))
	newVs := hw.Do(vs)
	// don't need vs anymore
	vs.Release()

	// Crafting timestamps.
	// Timestamps are deduced by summing the interval to the first/last valid timestamp.
	tsb := array.NewInt64Builder(fluxarrow.NewAllocator(hwt.alloc))
	s := stop.Add(hwt.interval)
	if hwt.withFit {
		s = start
	}
	for i := 0; i < newVs.Len(); i++ {
		tsb.Append(int64(s))
		s = s.Add(hwt.interval)
	}
	newTs := tsb.NewInt64Array()
	defer func() {
		newVs.Release()
		newTs.Release()
	}()

	// Appending columns.
	if err := builder.AppendTimes(newTimeIdx, newTs); err != nil {
		return err
	}
	if err := builder.AppendFloats(newValueIdx, newVs); err != nil {
		return err
	}
	if err := execute.AppendKeyValuesN(tbl.Key(), builder, newVs.Len()); err != nil {
		return err
	}
	return nil
}

// getCleanData returns cleaned data (using the value and time column), and the first and last valid timestamps.
// Below are the cleaning criteria.
// Rows that have a null timestamp get discarded.
// Rows that have a null value are considered invalid, but used by the algorithm.
// HoltWinters supposes to work with evenly spaced values in time, so:
//  - the Interval passed to the transformation is used to divide the data in time buckets;
//  - if many values are in the same bucket, the first one is selected, the others are skipped;
//  - if no value is present for a bucket, that is considered as an invalid value (treated like null values).
// HoltWinters will only be provided with the values returned.
// Timestamps can be deduced by summing interval to the first/last valid timestamp.
func (hwt *holtWintersTransformation) getCleanData(tbl flux.Table, colIdx, timeIdx int) (*array.Float64, values.Time, values.Time, error) {
	vs := array.NewFloat64Builder(fluxarrow.NewAllocator(hwt.alloc))
	var start, stop int64
	bucketEnd := int64(-1)
	bucketFilled := false
	roundTime := func(t int64) int64 {
		return int64(values.Time(t).Round(hwt.interval))
	}
	nextBucket := func() {
		bucketEnd += int64(hwt.interval.Duration())
		bucketFilled = false
	}
	appendV := func(cr flux.ColReader, i int) {
		switch typ := tbl.Cols()[colIdx].Type; typ {
		case flux.TInt:
			c := cr.Ints(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		case flux.TUInt:
			c := cr.UInts(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		case flux.TFloat:
			c := cr.Floats(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		default:
			panic(fmt.Sprintf("cannot append non-numerical type %s", typ.String()))
		}
		bucketFilled = true
	}
	isNull := func(cr flux.ColReader, i int) bool {
		switch typ := tbl.Cols()[colIdx].Type; typ {
		case flux.TInt:
			return cr.Ints(colIdx).IsNull(i)
		case flux.TUInt:
			return cr.UInts(colIdx).IsNull(i)
		case flux.TFloat:
			return cr.Floats(colIdx).IsNull(i)
		default:
			panic(fmt.Sprintf("cannot check non-numerical type %s", typ.String()))
		}
	}
	isFirst := func() bool {
		return bucketEnd == -1
	}
	if err := tbl.Do(func(cr flux.ColReader) error {
		// we work row-wise
		for i := 0; i < cr.Len(); i++ {
			// drop values with invalid timestamp
			if cts := cr.Times(timeIdx); cts.IsValid(i) {
				// the first value must be valid, skip it if it isn't so
				if isFirst() && isNull(cr, i) {
					continue
				}
				trueT := cts.Value(i)
				roundT := roundTime(trueT)
				// if this is the first valid ts, directly append the value and continue
				if isFirst() {
					start = trueT
					bucketEnd = roundT
					appendV(cr, i)
					continue
				}
				if roundT <= bucketEnd && bucketFilled {
					// drop values that occur for the same time bucket
					continue
				}
				// ok, this value is for a new bucket
				nextBucket()
				// append null for each empty bucket found
				for roundT > bucketEnd {
					vs.AppendNull()
					nextBucket()
				}
				// this is the first value for the bucket
				appendV(cr, i)
				stop = trueT
			}
		}
		return nil
	}); err != nil {
		return nil, 0, 0, err
	}
	return vs.NewFloat64Array(), values.Time(start), values.Time(stop), nil
}

func (hwt *holtWintersTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return hwt.d.RetractTable(key)
}

func (hwt *holtWintersTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return hwt.d.UpdateWatermark(mark)
}
func (hwt *holtWintersTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return hwt.d.UpdateProcessingTime(pt)
}
func (hwt *holtWintersTransformation) Finish(id execute.DatasetID, err error) {
	hwt.d.Finish(err)
}
