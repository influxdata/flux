package planner

import (
	"sort"
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

	// Helper methods for manipulating a plan
	// These methods are used during planning
	AddSuccessors(...PlanNode)
	AddPredecessors(...PlanNode)
	RemovePredecessor(PlanNode)
	RemoveSuccessor(PlanNode)
	ClearSuccessors()
}

type NodeID string

// PlanSpec holds the result nodes of a query plan with associated metadata
type PlanSpec struct {
	Roots     map[PlanNode]struct{}
	Results   map[string]PlanNode
	Resources flux.ResourceManagement
	Now       time.Time
}

// NewPlanSpec initializes a new query plan
func NewPlanSpec() *PlanSpec {
	return &PlanSpec{
		Roots:   make(map[PlanNode]struct{}),
		Results: make(map[string]PlanNode),
	}
}

// Replace replaces one of the root nodes of the query plan
func (plan *PlanSpec) Replace(root, with PlanNode) {
	delete(plan.Roots, root)
	plan.Roots[with] = struct{}{}
}

// TopDownWalk will execute f for each plan node in the PlanSpec.
// It always visits a node before visiting its predecessors.
func (plan *PlanSpec) TopDownWalk(f func(node PlanNode) error) error {
	visited := make(map[PlanNode]struct{})

	roots := make([]PlanNode, 0, len(plan.Roots))
	for root := range plan.Roots {
		roots = append(roots, root)
	}

	// Make sure to sort the roots first otherwise
	// an in-consistent walk order is possible.
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	postFn := func(PlanNode) error {
		return nil
	}

	for _, root := range roots {
		err := walk(root, f, postFn, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

// BottomUpWalk will execute f for each plan node in the PlanSpec,
// starting from the sources, and only visiting a node after all its
// predecessors have been visited.
func (plan *PlanSpec) BottomUpWalk(f func(PlanNode) error) error {
	visited := make(map[PlanNode]struct{})

	roots := make([]PlanNode, 0, len(plan.Roots))
	for root := range plan.Roots {
		roots = append(roots, root)
	}

	// Make sure to sort the roots first otherwise
	// an in-consistent walk order is possible.
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	preFn := func(PlanNode) error {
		return nil
	}

	for _, root := range roots {
		err := walk(root, preFn, f, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

func walk(node PlanNode, preFn, postFn func(PlanNode) error, visited map[PlanNode]struct{}) error {
	if _, ok := visited[node]; ok {
		return nil
	}

	visited[node] = struct{}{}

	// Pre-order traversal
	if err := preFn(node); err != nil {
		return err
	}

	for _, pred := range node.Predecessors() {
		walk(pred, preFn, postFn, visited)
	}

	// Post-order traversal
	return postFn(node)
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

func (e *edges) RemoveSuccessor(node PlanNode) {
	idx := -1
	for i, succ := range e.successors {
		if node == succ {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	} else if idx == len(e.successors)-1 {
		e.successors = e.successors[:idx]
	} else {
		e.successors = append(e.successors[:idx], e.successors[idx+1:]...)
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
