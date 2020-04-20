package plan

import "context"

// Rule is transformation rule for a query operation
type Rule interface {
	// Name of this rule (must be unique).
	Name() string

	// Pattern for this rule to match against.
	Pattern() Pattern

	// Rewrite an operation into an equivalent one.
	// The returned node is the new root of the sub tree.
	// The boolean return value should be true if anything changed during the rewrite.
	Rewrite(context.Context, Node) (Node, bool, error)
}
