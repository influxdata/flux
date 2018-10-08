package planner

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/platform"
)

const FromKind = "from"

// FromProcedureSpec <=> from()
type FromProcedureSpec struct {
	Bucket   string
	BucketID platform.ID
}

func (spec *FromProcedureSpec) Copy() ProcedureSpec {
	return &FromProcedureSpec{
		Bucket:   spec.Bucket,
		BucketID: spec.BucketID,
	}
}
func (spec *FromProcedureSpec) Kind() ProcedureKind {
	return FromKind
}

const RangeKind = "range"

// RangeProcedureSpec <=> range()
type RangeProcedureSpec struct {
	Bounds   flux.Bounds
	TimeCol  string
	StartCol string
	StopCol  string
}

func (spec *RangeProcedureSpec) Copy() ProcedureSpec {
	return &RangeProcedureSpec{
		Bounds:   spec.Bounds,
		TimeCol:  spec.TimeCol,
		StartCol: spec.StartCol,
		StopCol:  spec.StopCol,
	}
}
func (spec *RangeProcedureSpec) Kind() ProcedureKind {
	return RangeKind
}

const FilterKind = "filter"

// FilterProcedureSpec <=> filter()
type FilterProcedureSpec struct {
	Fn *semantic.FunctionExpression
}

func (spec *FilterProcedureSpec) Copy() ProcedureSpec {
	return &FilterProcedureSpec{
		Fn: spec.Fn.Copy().(*semantic.FunctionExpression),
	}
}
func (spec *FilterProcedureSpec) Kind() ProcedureKind {
	return FilterKind
}

const YieldKind = "yield"

// YieldProcedureSpec <=> yield()
type YieldProcedureSpec struct {
	Name string
}

func (spec *YieldProcedureSpec) Copy() ProcedureSpec {
	return &YieldProcedureSpec{
		Name: spec.Name,
	}
}

func (spec *YieldProcedureSpec) Kind() ProcedureKind {
	return YieldKind
}

const JoinKind = "join"

// joinProcedureSpec <=> join()
type JoinProcedureSpec struct {
	On []string
}

func (spec *JoinProcedureSpec) Copy() ProcedureSpec {
	onList := make([]string, len(spec.On))
	copy(onList, spec.On)
	return &JoinProcedureSpec{
		On: onList,
	}
}

func (spec *JoinProcedureSpec) Kind() ProcedureKind {
	return JoinKind
}
