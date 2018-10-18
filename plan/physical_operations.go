package plan

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/platform"
)

const FromRangeKind = "FromRange"

// FromRangeProcedureSpec represents a sequential scanning operation
type FromRangeProcedureSpec struct {
	Bucket   string
	BucketID platform.ID
	Bounds   flux.Bounds
	TimeCol  string
	StartCol string
	StopCol  string
}

func (spec *FromRangeProcedureSpec) Kind() ProcedureKind {
	return FromRangeKind
}
func (spec *FromRangeProcedureSpec) Copy() ProcedureSpec {
	return &FromRangeProcedureSpec{
		Bucket:   spec.Bucket,
		BucketID: spec.BucketID,
		Bounds:   spec.Bounds,
		TimeCol:  spec.TimeCol,
		StartCol: spec.StartCol,
		StopCol:  spec.StopCol,
	}
}
func (spec *FromRangeProcedureSpec) Cost(inStats []Statistics) (Cost, Statistics) {
	return Cost{}, Statistics{}
}

const FromTagFilterKind = "FromTagFilter"

// FromTagFilterProcedureSpec represents an index scanning operation
type FromTagFilterProcedureSpec struct {
	Bucket   string
	BucketID platform.ID

	// A conjunctive predicate of tag, value equalities.
	// For example, if the predicate is: r.tag1 = "A" AND r.tag1 == "B",
	// then TagEqualityFilters is: map[string]string{"tag0": "A", "tag1": "B"}.
	TagEqualityFilters map[string]string
	PredicateFunction  *semantic.FunctionExpression
}

func (spec *FromTagFilterProcedureSpec) Kind() ProcedureKind {
	return FromTagFilterKind
}
func (spec *FromTagFilterProcedureSpec) Copy() ProcedureSpec {
	tagPreds := make(map[string]string, len(spec.TagEqualityFilters))
	for k, v := range spec.TagEqualityFilters {
		tagPreds[k] = v
	}
	return &FromTagFilterProcedureSpec{
		Bucket:   spec.Bucket,
		BucketID: spec.BucketID,

		TagEqualityFilters: tagPreds,
	}
}
func (spec *FromTagFilterProcedureSpec) Cost(inStats []Statistics) (Cost, Statistics) {
	return Cost{}, Statistics{}
}

const FromRangeFieldFilterKind = "FromRangeFieldFilter"

// FromRangeFieldFilterProcedureSpec represents a sequential scanning operation
type FromRangeFieldFilterProcedureSpec struct {
	Bucket   string
	BucketID platform.ID
	Bounds   flux.Bounds
	TimeCol  string
	StartCol string
	StopCol  string

	// predicate not involving any tags
	// For example, (r) => r._value > 0
	PredicateFn *semantic.FunctionExpression
}

func (spec *FromRangeFieldFilterProcedureSpec) Kind() ProcedureKind {
	return FromRangeFieldFilterKind
}
func (spec *FromRangeFieldFilterProcedureSpec) Copy() ProcedureSpec {
	return &FromRangeFieldFilterProcedureSpec{
		Bucket:      spec.Bucket,
		BucketID:    spec.BucketID,
		Bounds:      spec.Bounds,
		TimeCol:     spec.TimeCol,
		StartCol:    spec.StartCol,
		StopCol:     spec.StopCol,
		PredicateFn: spec.PredicateFn.Copy().(*semantic.FunctionExpression),
	}
}
func (spec *FromRangeFieldFilterProcedureSpec) Cost(inStats []Statistics) (Cost, Statistics) {
	return Cost{}, Statistics{}
}
