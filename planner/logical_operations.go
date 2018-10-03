package planner

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/platform"
)

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
	return "FromKind"
}

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
	return "RangeKind"
}

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
	return "FilterKind"
}
