package planner

import (
	"github.com/influxdata/flux/plan"
)

// PlanNode defines the common interface for interacting with
// logical and physical plan nodes.
type PlanNode interface {
	// Plan nodes executed immediately before this node
	Predecessors() []PlanNode

	// Plan nodes executed immediately after this node
	Successors() []PlanNode

	ProcedureSpec() plan.ProcedureSpec

	// The types of the tables produced by this node
	// Is it possible to know this at plan time?
	// Type() []semantic.Type
}

type Edges struct {
	predecessors []PlanNode
	successors   []PlanNode
}

func (e *Edges) Predecessors() []PlanNode {
	return e.predecessors
}

func (e *Edges) Successors() []PlanNode {
	return e.successors
}
