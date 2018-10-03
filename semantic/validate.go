package semantic

type ValidateVisitor struct{}

func (v ValidateVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *FunctionExpression:
		// Ensure the function expression does not have nil children
		if n.Defaults == nil {
			n.Defaults = new(FunctionDefaults)
		}
		if n.Params == nil {
			n.Params = new(FunctionParams)
		}
	}
	return v
}

func (v ValidateVisitor) Done() {}

// Validate ensures that it is safe to walk the node.
// The node may be modified in order to ensure safe traversal.
// The modification will not change the semantics of the graph.
func Validate(n Node) {
	Walk(ValidateVisitor{}, n)
}
