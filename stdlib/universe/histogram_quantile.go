package universe

import (
	"math"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const (
	HistogramQuantileKind = "histogramQuantile"

	DefaultUpperBoundColumnLabel = "le"

	onNonmonotonicError = "error"
	onNonmonotonicDrop  = "drop"
	onNonmonotonicForce = "force"
)

type HistogramQuantileOpSpec struct {
	Quantile         float64 `json:"quantile"`
	CountColumn      string  `json:"countColumn"`
	UpperBoundColumn string  `json:"upperBoundColumn"`
	ValueColumn      string  `json:"valueColumn"`
	MinValue         float64 `json:"minValue"`
	OnNonmonotonic   string  `json:"onNonmonotonic"`
}

func init() {
	histogramQuantileSignature := runtime.MustLookupBuiltinType("universe", "histogramQuantile")

	runtime.RegisterPackageValue("universe", HistogramQuantileKind, flux.MustValue(flux.FunctionValue(HistogramQuantileKind, CreateHistogramQuantileOpSpec, histogramQuantileSignature)))
	plan.RegisterProcedureSpec(HistogramQuantileKind, newHistogramQuantileProcedure, HistogramQuantileKind)
	execute.RegisterTransformation(HistogramQuantileKind, createHistogramQuantileTransformation)
}
func CreateHistogramQuantileOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	s := new(HistogramQuantileOpSpec)
	q, err := args.GetRequiredFloat("quantile")
	if err != nil {
		return nil, err
	}
	s.Quantile = q

	if col, ok, err := args.GetString("countColumn"); err != nil {
		return nil, err
	} else if ok {
		s.CountColumn = col
	} else {
		s.CountColumn = execute.DefaultValueColLabel
	}

	if col, ok, err := args.GetString("upperBoundColumn"); err != nil {
		return nil, err
	} else if ok {
		s.UpperBoundColumn = col
	} else {
		s.UpperBoundColumn = DefaultUpperBoundColumnLabel
	}

	if col, ok, err := args.GetString("valueColumn"); err != nil {
		return nil, err
	} else if ok {
		s.ValueColumn = col
	} else {
		s.ValueColumn = execute.DefaultValueColLabel
	}

	if min, ok, err := args.GetFloat("minValue"); err != nil {
		return nil, err
	} else if ok {
		s.MinValue = min
	}

	if onNonmonotonic, ok, err := args.GetString("onNonmonotonic"); err != nil {
		return nil, err
	} else if ok {
		s.OnNonmonotonic = onNonmonotonic
	} else {
		s.OnNonmonotonic = onNonmonotonicError
	}

	if s.OnNonmonotonic != onNonmonotonicError && s.OnNonmonotonic != onNonmonotonicForce && s.OnNonmonotonic != onNonmonotonicDrop {
		return nil, errors.Newf(codes.Invalid, "value provided to histogramQuantile parameter onNonmonotonic is invalid; must be one of %q, %q or %q", onNonmonotonicError, onNonmonotonicForce, onNonmonotonicDrop)
	}

	return s, nil
}

func (s *HistogramQuantileOpSpec) Kind() flux.OperationKind {
	return HistogramQuantileKind
}

type HistogramQuantileProcedureSpec struct {
	plan.DefaultCost
	Quantile         float64 `json:"quantile"`
	CountColumn      string  `json:"countColumn"`
	UpperBoundColumn string  `json:"upperBoundColumn"`
	ValueColumn      string  `json:"valueColumn"`
	MinValue         float64 `json:"minValue"`
	OnNonmonotonic   string  `json:"onNonmonotonic"`
}

func newHistogramQuantileProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*HistogramQuantileOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &HistogramQuantileProcedureSpec{
		Quantile:         spec.Quantile,
		CountColumn:      spec.CountColumn,
		UpperBoundColumn: spec.UpperBoundColumn,
		ValueColumn:      spec.ValueColumn,
		MinValue:         spec.MinValue,
		OnNonmonotonic:   spec.OnNonmonotonic,
	}, nil
}

func (s *HistogramQuantileProcedureSpec) Kind() plan.ProcedureKind {
	return HistogramQuantileKind
}
func (s *HistogramQuantileProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(HistogramQuantileProcedureSpec)
	*ns = *s
	return ns
}

type histogramQuantileTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec HistogramQuantileProcedureSpec
}

type bucket struct {
	count      float64
	upperBound float64
}

func createHistogramQuantileTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*HistogramQuantileProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewHistorgramQuantileTransformation(d, cache, s)
	return t, d, nil
}

func NewHistorgramQuantileTransformation(
	d execute.Dataset,
	cache execute.TableBuilderCache,
	spec *HistogramQuantileProcedureSpec,
) execute.Transformation {
	return &histogramQuantileTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
	}
}

func (t histogramQuantileTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	// TODO
	return nil
}

func (t histogramQuantileTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "histogramQuantile found duplicate table with key: %v", tbl.Key())
	}

	if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}
	valueIdx, err := builder.AddCol(flux.ColMeta{
		Label: t.spec.ValueColumn,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}

	countIdx := execute.ColIdx(t.spec.CountColumn, tbl.Cols())
	if countIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "table is missing count column %q", t.spec.CountColumn)
	}
	if tbl.Cols()[countIdx].Type != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "count column %q must be of type float", t.spec.CountColumn)
	}
	upperBoundIdx := execute.ColIdx(t.spec.UpperBoundColumn, tbl.Cols())
	if upperBoundIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "table is missing upper bound column %q", t.spec.UpperBoundColumn)
	}
	if tbl.Cols()[upperBoundIdx].Type != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "upper bound column %q must be of type float", t.spec.UpperBoundColumn)
	}
	// Read buckets
	var cdf []bucket
	sorted := true // track if the cdf was naturally sorted
	if err := tbl.Do(func(cr flux.ColReader) error {
		offset := len(cdf)
		// Grow cdf by number of rows
		l := offset + cr.Len()
		if cap(cdf) < l {
			cpy := make([]bucket, l, l*2)
			// Copy existing buckets to new slice
			copy(cpy, cdf)
			cdf = cpy
		} else {
			cdf = cdf[:l]
		}
		for i := 0; i < cr.Len(); i++ {
			curr := i + offset
			prev := curr - 1

			b := bucket{}
			if vs := cr.Floats(countIdx); vs.IsValid(i) {
				b.count = vs.Value(i)
			} else {
				return errors.Newf(codes.FailedPrecondition, "unexpected null in the countColumn")
			}
			if vs := cr.Floats(upperBoundIdx); vs.IsValid(i) {
				b.upperBound = vs.Value(i)
			} else {
				return errors.Newf(codes.FailedPrecondition, "unexpected null in the upperBoundColumn")
			}
			cdf[curr] = b
			if prev >= 0 {
				sorted = sorted && cdf[prev].upperBound <= cdf[curr].upperBound
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if !sorted {
		sort.Slice(cdf, func(i, j int) bool {
			return cdf[i].upperBound < cdf[j].upperBound
		})
	}

	result, err := t.computeQuantile(cdf)
	if err != nil {
		return err
	}
	if result.action == drop {
		return nil
	}
	if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
		return err
	}
	if result.action == appendValue {
		if err := builder.AppendFloat(valueIdx, result.v); err != nil {
			return err
		}
	} else {
		// action is appendNil
		if err := builder.AppendNil(valueIdx); err != nil {
			return err
		}

	}
	return nil
}

type quantileAction int

const (
	appendValue quantileAction = iota
	appendNil
	drop
)

type quantileResult struct {
	action quantileAction
	v      float64
}

// isMonotonic will check if the buckets are monotonic and
// return true if so.
//
// If force is set, it will force them to be monotonic
// by assuming no increase from the previous bucket.
// When force is is set, this function will always return true.
func isMonotonic(force bool, cdf []bucket) bool {
	prevCount := 0.0
	for i := range cdf {
		if cdf[i].count < prevCount {
			if force {
				cdf[i].count = prevCount
			} else {
				return false
			}
		} else {
			prevCount = cdf[i].count
		}
	}
	return true
}

func (t *histogramQuantileTransformation) computeQuantile(cdf []bucket) (quantileResult, error) {
	if len(cdf) == 0 {
		return quantileResult{}, errors.New(codes.FailedPrecondition, "histogram is empty")
	}

	if !isMonotonic(t.spec.OnNonmonotonic == onNonmonotonicForce, cdf) {
		switch t.spec.OnNonmonotonic {
		case onNonmonotonicError:
			return quantileResult{}, errors.New(codes.FailedPrecondition, "histogram records counts are not monotonic")
		case onNonmonotonicDrop:
			return quantileResult{action: drop}, nil
		default:
			// "force" is not possible because isMonotonic will fix the buckets
			return quantileResult{}, errors.Newf(codes.Internal, "unknown or unexpected value for onNonmonotonic: %q", t.spec.OnNonmonotonic)
		}
	}

	// Find rank index and check counts are monotonic
	totalCount := cdf[len(cdf)-1].count
	if totalCount == 0 {
		// Produce a null value if there were no samples
		return quantileResult{action: appendNil}, nil
	}

	rank := t.spec.Quantile * totalCount
	rankIdx := -1
	for i, b := range cdf {
		if rank >= b.count {
			rankIdx = i
		}
	}

	var (
		lowerCount,
		lowerBound,
		upperCount,
		upperBound float64
	)
	switch rankIdx {
	case -1:
		// Quantile is below the lowest upper bound, interpolate using the min value
		lowerCount = 0
		lowerBound = t.spec.MinValue
		upperCount = cdf[0].count
		upperBound = cdf[0].upperBound
	case len(cdf) - 1:
		// Quantile is above the highest upper bound, simply return it as it must be finite
		return quantileResult{action: appendValue, v: cdf[len(cdf)-1].upperBound}, nil
	default:
		lowerCount = cdf[rankIdx].count
		lowerBound = cdf[rankIdx].upperBound
		upperCount = cdf[rankIdx+1].count
		upperBound = cdf[rankIdx+1].upperBound
	}
	if rank == lowerCount {
		// No need to interpolate
		return quantileResult{action: appendValue, v: lowerBound}, nil
	}
	if math.IsInf(lowerBound, -1) {
		// We cannot interpolate with infinity
		return quantileResult{action: appendValue, v: upperBound}, nil
	}
	if math.IsInf(upperBound, 1) {
		// We cannot interpolate with infinity
		return quantileResult{action: appendValue, v: lowerBound}, nil
	}
	// Compute quantile using linear interpolation
	scale := (rank - lowerCount) / (upperCount - lowerCount)
	return quantileResult{action: appendValue, v: lowerBound + (upperBound-lowerBound)*scale}, nil
}

func (t histogramQuantileTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t histogramQuantileTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t histogramQuantileTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
