package planner

// Rule is transformation rule for a query operation
type Rule interface {
	// Pattern for this rule to match against
	Pattern() Pattern

	// Rewrite an operation into an equivalent one
	Rewrite(PlanNode) (PlanNode, bool)
}

