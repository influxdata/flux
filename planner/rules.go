package planner

// Rule is transformation rule for a query operation
type Rule interface {
	// Pattern for this rule to match against
	Pattern() Pattern

	// Rewrite an operation into an equivalent one
	Rewrite(PlanNode) PlanNode
}

// Pattern represents an operator tree pattern
// It can match itself against a query plan
type Pattern interface {
	Match(PlanNode) bool
}
