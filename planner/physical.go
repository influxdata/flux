package planner

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/platform"
)

// PhysicalPlan represents a query tree/subtree whose nodes
// are the physical operations/algorithms to be performed.
type PhysicalPlan interface {
	physical()
	Predecessors() []PhysicalPlan
	Successors() []PhysicalPlan
	Cost() Cost
}

func (node *FromRangeScan) physical()      {}
func (node *FromRangeIndexScan) physical() {}
func (node *Filter) physical()             {}
func (node *Sort) physical()               {}

func (node *FromRangeIndexScan) Predecessors() []PhysicalPlan {
	return node.predecessors
}
func (node *FromRangeScan) Predecessors() []PhysicalPlan {
	return node.predecessors
}
func (node *Filter) Predecessors() []PhysicalPlan {
	return node.predecessors
}
func (node *Sort) Predecessors() []PhysicalPlan {
	return node.predecessors
}

func (node *FromRangeIndexScan) Successors() []PhysicalPlan {
	return node.successors
}
func (node *FromRangeScan) Successors() []PhysicalPlan {
	return node.successors
}
func (node *Filter) Successors() []PhysicalPlan {
	return node.successors
}
func (node *Sort) Successors() []PhysicalPlan {
	return node.successors
}

// FromRangeScan represents the physical operation of
// scanning a series from a start time to an end time.
type FromRangeScan struct {
	Bucket   string
	BucketID platform.ID
	Bounds   flux.Bounds

	predecessors []PhysicalPlan
	successors   []PhysicalPlan
}

func (node *FromRangeScan) cost() Cost { return Cost{} }

// FromRangeIndexScan represents the physical operation of
// scanning a series within a specified time range using an index.
type FromRangeIndexScan struct {
	Bucket   string
	BucketID platform.ID
	Bounds   flux.Bounds
	Filter   *semantic.FunctionExpression

	predecessors []PhysicalPlan
	successors   []PhysicalPlan
}

func (node *FromRangeIndexScan) cost() Cost { return Cost{} }

// Filter represents the physical operation of filtering a
// stream of data based on some condition. No index is involved.
type Filter struct {
	Schema    []flux.ColMeta
	Condition *semantic.FunctionExpression

	predecessors []PhysicalPlan
	successors   []PhysicalPlan
}

func (node *Filter) cost() Cost { return Cost{} }

// Sort represents the physical operations of sorting a stream
// according to the values in a specified column/columns.
type Sort struct {
	Schema flux.ColMeta
	On     []flux.ColMeta

	predecessors []PhysicalPlan
	successors   []PhysicalPlan
}

func (node *Sort) cost() Cost { return Cost{} }
