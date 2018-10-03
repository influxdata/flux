package planner

// Rule is transformation rule for a query operation
type Rule interface {
	// Pattern for this rule to match against
	Pattern() Pattern

	// Rewrite an operation into an equivalent one
	Rewrite(PlanNode) (PlanNode, bool)
}

// Pattern represents an operator tree pattern
// It can match itself against a query plan
type Pattern interface {
	Root() ProcedureKind
	Match(PlanNode) bool
}

// LeafPattern implements Pattern.
// It matches any node that has no predecessors.
type LeafPattern struct {
	RootType ProcedureKind
}

func (lp *LeafPattern) Root() ProcedureKind {
	return lp.RootType
}

func (lp *LeafPattern) Match(node PlanNode) bool {
	return node.ProcedureSpec().Kind() == lp.Root() && (node.Predecessors() == nil || len(node.Predecessors()) == 0)
}

// AnyPattern implements Pattern.
// It matches a node that can have any number of predecessors.
type AnyPattern struct {
	RootType ProcedureKind
}

func (ap *AnyPattern) Root() ProcedureKind {
	return ap.RootType
}

func (ap *AnyPattern) Match(node PlanNode) bool {
	return node.ProcedureSpec().Kind() == ap.Root()
}

// TreePattern implements Pattern.
// It matches a single node type and a predecessor pattern.
type TreePattern struct {
	RootType     ProcedureKind
	Predecessors []Pattern
}

func (tp *TreePattern) Root() ProcedureKind {
	return tp.RootType
}

func (tp *TreePattern) Match(node PlanNode) bool {
	if node.ProcedureSpec().Kind() != tp.Root() {
		return false
	}

	for i, n := range node.Predecessors() {
		if !tp.Predecessors[i].Match(n) {
			return false
		}
	}

	return true
}
