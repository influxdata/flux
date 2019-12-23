package promql

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

// TODO: Added "prom" prefix to avoid duplicate registration error. Decide whether to
// remove universe version of this function.
const HistogramQuantileKind = "promHistogramQuantile"

const DefaultUpperBoundColumnLabel = "le"

type HistogramQuantileOpSpec struct {
	Quantile         float64 `json:"quantile"`
	CountColumn      string  `json:"countColumn"`
	UpperBoundColumn string  `json:"upperBoundColumn"`
	ValueColumn      string  `json:"valueColumn"`
}

func init() {
	histogramQuantileSignature := flux.LookupBuiltInType("internal/promql", HistogramQuantileKind)

	flux.RegisterPackageValue("internal/promql", HistogramQuantileKind, flux.MustValue(flux.FunctionValue(HistogramQuantileKind, createHistogramQuantileOpSpec, histogramQuantileSignature)))
	flux.RegisterOpSpec(HistogramQuantileKind, newHistogramQuantileOp)
	plan.RegisterProcedureSpec(HistogramQuantileKind, newHistogramQuantileProcedure, HistogramQuantileKind)
	execute.RegisterTransformation(HistogramQuantileKind, createHistogramQuantileTransformation)
}
func createHistogramQuantileOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
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

	return s, nil
}

func newHistogramQuantileOp() flux.OperationSpec {
	return new(HistogramQuantileOpSpec)
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
}

func newHistogramQuantileProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*HistogramQuantileOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &HistogramQuantileProcedureSpec{
		Quantile:         spec.Quantile,
		CountColumn:      spec.CountColumn,
		UpperBoundColumn: spec.UpperBoundColumn,
		ValueColumn:      spec.ValueColumn,
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
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewHistogramQuantileTransformation(d, cache, s)
	return t, d, nil
}

func NewHistogramQuantileTransformation(
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
		return fmt.Errorf("histogramQuantile found duplicate table with key: %v", tbl.Key())
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
		return fmt.Errorf("table is missing count column %q", t.spec.CountColumn)
	}
	if tbl.Cols()[countIdx].Type != flux.TFloat {
		return fmt.Errorf("count column %q must be of type float", t.spec.CountColumn)
	}
	upperBoundIdx := execute.ColIdx(t.spec.UpperBoundColumn, tbl.Cols())
	if upperBoundIdx < 0 {
		// No "le" labels present at all, return empty result.
		return nil
	}
	if tbl.Cols()[upperBoundIdx].Type != flux.TString {
		return fmt.Errorf("upper bound column %q must be of type string", t.spec.UpperBoundColumn)
	}
	// Read buckets
	var cdf []bucket
	sorted := true //track if the cdf was naturally sorted
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
				return fmt.Errorf("unexpected null in the countColumn")
			}
			if vs := cr.Strings(upperBoundIdx); vs.IsValid(i) {
				upperBound, err := strconv.ParseFloat(string(vs.Value(i)), 64)
				if err != nil {
					// "le" label value invalid, skip.
					continue
				}
				b.upperBound = upperBound
			} else {
				// "le" label missing, skip.
				continue
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

	q := bucketQuantile(t.spec.Quantile, cdf)
	if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
		return err
	}
	if err := builder.AppendFloat(valueIdx, q); err != nil {
		return err
	}
	return nil
}

// The following functions (bucketQuantile(), coalesceBuckets(), and
// ensureMonotonic()) have been taken verbatim from Prometheus. The original
// copyright notice is as follows:
//
// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// bucketQuantile calculates the quantile 'q' based on the given buckets. The
// buckets will be sorted by upperBound by this function (i.e. no sorting
// needed before calling this function). The quantile value is interpolated
// assuming a linear distribution within a bucket. However, if the quantile
// falls into the highest bucket, the upper bound of the 2nd highest bucket is
// returned. A natural lower bound of 0 is assumed if the upper bound of the
// lowest bucket is greater 0. In that case, interpolation in the lowest bucket
// happens linearly between 0 and the upper bound of the lowest bucket.
// However, if the lowest bucket has an upper bound less or equal 0, this upper
// bound is returned if the quantile falls into the lowest bucket.
//
// There are a number of special cases (once we have a way to report errors
// happening during evaluations of AST functions, we should report those
// explicitly):
//
// If 'buckets' has fewer than 2 elements, NaN is returned.
//
// If the highest bucket is not +Inf, NaN is returned.
//
// If q<0, -Inf is returned.
//
// If q>1, +Inf is returned.
func bucketQuantile(q float64, buckets []bucket) float64 {
	if q < 0 {
		return math.Inf(-1)
	}
	if q > 1 {
		return math.Inf(+1)
	}
	if !math.IsInf(buckets[len(buckets)-1].upperBound, +1) {
		return math.NaN()
	}

	buckets = coalesceBuckets(buckets)
	ensureMonotonic(buckets)

	if len(buckets) < 2 {
		return math.NaN()
	}

	rank := q * buckets[len(buckets)-1].count
	b := sort.Search(len(buckets)-1, func(i int) bool { return buckets[i].count >= rank })

	if b == len(buckets)-1 {
		return buckets[len(buckets)-2].upperBound
	}
	if b == 0 && buckets[0].upperBound <= 0 {
		return buckets[0].upperBound
	}
	var (
		bucketStart float64
		bucketEnd   = buckets[b].upperBound
		count       = buckets[b].count
	)
	if b > 0 {
		bucketStart = buckets[b-1].upperBound
		count -= buckets[b-1].count
		rank -= buckets[b-1].count
	}
	return bucketStart + (bucketEnd-bucketStart)*(rank/count)
}

// coalesceBuckets merges buckets with the same upper bound.
//
// The input buckets must be sorted.
func coalesceBuckets(buckets []bucket) []bucket {
	last := buckets[0]
	i := 0
	for _, b := range buckets[1:] {
		if b.upperBound == last.upperBound {
			last.count += b.count
		} else {
			buckets[i] = last
			last = b
			i++
		}
	}
	buckets[i] = last
	return buckets[:i+1]
}

// The assumption that bucket counts increase monotonically with increasing
// upperBound may be violated during:
//
//   * Recording rule evaluation of histogram_quantile, especially when rate()
//      has been applied to the underlying bucket timeseries.
//   * Evaluation of histogram_quantile computed over federated bucket
//      timeseries, especially when rate() has been applied.
//
// This is because scraped data is not made available to rule evaluation or
// federation atomically, so some buckets are computed with data from the
// most recent scrapes, but the other buckets are missing data from the most
// recent scrape.
//
// Monotonicity is usually guaranteed because if a bucket with upper bound
// u1 has count c1, then any bucket with a higher upper bound u > u1 must
// have counted all c1 observations and perhaps more, so that c  >= c1.
//
// Randomly interspersed partial sampling breaks that guarantee, and rate()
// exacerbates it. Specifically, suppose bucket le=1000 has a count of 10 from
// 4 samples but the bucket with le=2000 has a count of 7 from 3 samples. The
// monotonicity is broken. It is exacerbated by rate() because under normal
// operation, cumulative counting of buckets will cause the bucket counts to
// diverge such that small differences from missing samples are not a problem.
// rate() removes this divergence.)
//
// bucketQuantile depends on that monotonicity to do a binary search for the
// bucket with the φ-quantile count, so breaking the monotonicity
// guarantee causes bucketQuantile() to return undefined (nonsense) results.
//
// As a somewhat hacky solution until ingestion is atomic per scrape, we
// calculate the "envelope" of the histogram buckets, essentially removing
// any decreases in the count between successive buckets.
func ensureMonotonic(buckets []bucket) {
	max := buckets[0].count
	for i := range buckets[1:] {
		switch {
		case buckets[i].count > max:
			max = buckets[i].count
		case buckets[i].count < max:
			buckets[i].count = max
		}
	}
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
