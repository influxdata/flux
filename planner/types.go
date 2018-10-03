package planner

// PlanNode defines the common interface for interacting with
// logical and physical plan nodes.
type PlanNode interface {
	// Plan nodes executed immediately before this node
	Predecessors() []PlanNode

	// Plan nodes executed immediately after this node
	Successors() []PlanNode

	ProcedureSpec() ProcedureSpec

	// The types of the tables produced by this node
	// Is it possible to know this at plan time?
	// Type() []semantic.Type

	// Helper methods for manipulating a plan
	// These methods are used during planning
	AddSuccessor(PlanNode)
	AddPredecessor(PlanNode)
	RemovePredecessor(PlanNode)
}

// ProcedureSpec specifies a query operation
type ProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
}

// ProcedureKind denotes the kind of operation
type ProcedureKind string

type Edges struct {
	Pred []PlanNode
	Succ []PlanNode
}

func (e *Edges) Predecessors() []PlanNode {
	return e.Pred
}

func (e *Edges) Successors() []PlanNode {
	return e.Succ
}

func (e *Edges) AddSuccessor(node PlanNode) {
	e.Succ = append(e.Succ, node)
}

func (e *Edges) AddPredecessor(node PlanNode) {
	e.Pred = append(e.Pred, node)
}

func (e *Edges) RemovePredecessor(node PlanNode) {
	idx := -1
	for i, pred := range e.Pred {
		if node == pred {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	} else if idx == len(e.Pred)-1 {
		e.Pred = e.Pred[:idx]
	} else {
		e.Pred = append(e.Pred[:idx], e.Pred[idx+1:]...)
	}
}
