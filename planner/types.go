package planner

import (
	"time"

	"github.com/influxdata/flux"
)

const DefaultYieldName = "_result"

// PlanNode defines the common interface for interacting with
// logical and physical plan nodes.
type PlanNode interface {
	// Returns an identifier for this plan node
	ID() NodeID

	// Plan nodes executed immediately before this node
	Predecessors() []PlanNode

	// Plan nodes executed immediately after this node
	Successors() []PlanNode

	// Specification of the procedure represented by this node
	ProcedureSpec() ProcedureSpec

	// Type of procedure represented by this node
	Kind() ProcedureKind

	// The types of the tables produced by this node
	// Is it possible to know this at plan time?
	// Type() []semantic.Type

	// Helper methods for manipulating a plan
	// These methods are used during planning
	AddSuccessors(...PlanNode)
	AddPredecessors(...PlanNode)
	RemovePredecessor(PlanNode)
	ClearSuccessors()
}

type NodeID string

// PlanSpec holds the roots of the query plan DAG.
// Roots correspond to nodes that produce final results.
type PlanSpec struct {
	roots     map[PlanNode]struct{}
	Resources flux.ResourceManagement
	Now       time.Time
}

// NewPlanSpec instantiates a new plan spec.
func NewPlanSpec(roots []PlanNode) *PlanSpec {
	r := make(map[PlanNode]struct{}, len(roots))
	for _, root := range roots {
		r[root] = struct{}{}
	}
	return &PlanSpec{roots: r}
}

// Roots returns the roots (the successor-less nodes) of the query plan
func (plan *PlanSpec) Roots() []PlanNode {
	roots := []PlanNode{}
	for k := range plan.roots {
		roots = append(roots, k)
	}
	return roots
}

// Replace replaces one of the roots of the query plan
func (plan *PlanSpec) Replace(root, with PlanNode) {
	delete(plan.roots, root)
	plan.roots[with] = struct{}{}
}

// BottomUpWalk will execute f for each plan node in the PlanSpec,
// starting from the sources, and only visiting a node after all its
// predecessors have been visited.
func (p *PlanSpec) BottomUpWalk(f func(PlanNode) error) error {
	visited := make(map[PlanNode]struct{})

	for _, root := range p.Roots() {
		err := walk(root, f, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

func walk(node PlanNode, f func(PlanNode) error, visited map[PlanNode]struct{}) error {
	if _, ok := visited[node]; ok {
		return nil
	}

	visited[node] = struct{}{}

	// Visit each predecessor first
	for _, pred := range node.Predecessors() {
		walk(pred, f, visited)
	}

	return f(node)
}

type YieldProcedureSpec interface {
	YieldName() string
}

// ProcedureSpec specifies a query operation
type ProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
}

// ProcedureKind denotes the kind of operation
type ProcedureKind string

type edges struct {
	predecessors []PlanNode
	successors   []PlanNode
}

func (e *edges) Predecessors() []PlanNode {
	return e.predecessors
}

func (e *edges) Successors() []PlanNode {
	return e.successors
}

func (e *edges) AddSuccessors(nodes ...PlanNode) {
	e.successors = append(e.successors, nodes...)
}

func (e *edges) AddPredecessors(nodes ...PlanNode) {
	e.predecessors = append(e.predecessors, nodes...)
}

func (e *edges) RemovePredecessor(node PlanNode) {
	idx := -1
	for i, pred := range e.predecessors {
		if node == pred {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	} else if idx == len(e.predecessors)-1 {
		e.predecessors = e.predecessors[:idx]
	} else {
		e.predecessors = append(e.predecessors[:idx], e.predecessors[idx+1:]...)
	}
}

func (e *edges) ClearSuccessors() {
	e.successors = e.successors[0:0]
}

type WindowSpec struct {
	Every  flux.Duration
	Period flux.Duration
	Round  flux.Duration
	Start  flux.Time
}
