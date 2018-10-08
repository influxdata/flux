package planner

// PlanNode defines the common interface for interacting with
// logical and physical plan nodes.
type PlanNode interface {
	// Returns an identifier this this plan node
	ID() NodeID

	// Plan nodes executed immediately before this node
	Predecessors() []PlanNode

	// Plan nodes executed immediately after this node
	Successors() []PlanNode

	ProcedureSpec() ProcedureSpec
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

// QueryPlan holds the roots of the query plan DAG.
// Roots correspond to nodes that produce final results.
type QueryPlan struct {
	roots map[PlanNode]struct{}
}

// NewQueryPlan instantiates a new query plan
func NewQueryPlan(roots []PlanNode) *QueryPlan {
	r := make(map[PlanNode]struct{}, len(roots))
	for _, root := range roots {
		r[root] = struct{}{}
	}
	return &QueryPlan{roots: r}
}

// Roots returns the roots of the query plan
func (plan *QueryPlan) Roots() []PlanNode {
	roots := []PlanNode{}
	for k := range plan.roots {
		roots = append(roots, k)
	}
	return roots
}

// Replace replaces one of the roots of the query plan
func (plan *QueryPlan) Replace(root, with PlanNode) {
	delete(plan.roots, root)
	plan.roots[with] = struct{}{}
}

// ProcedureSpec specifies a query operation
type ProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
}

// ProcedureKind denotes the kind of operation
type ProcedureKind string

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

func (e *Edges) AddSuccessors(nodes ...PlanNode) {
	e.successors = append(e.successors, nodes...)
}

func (e *Edges) AddPredecessors(nodes ...PlanNode) {
	e.predecessors = append(e.predecessors, nodes...)
}

func (e *Edges) RemovePredecessor(node PlanNode) {
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

func (e *Edges) ClearSuccessors() {
	e.successors = e.successors[0:0]
}
