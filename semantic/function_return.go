package semantic

// FunctionReturnVisitor visits an annotated semantic graph
// and records every function's return type variable.
type FunctionReturnVisitor struct {
	tenv       map[Node]TypeVar
	current    Node
	returnType TypeVar
}

// NewFunctionReturnVisitor instantiates a new FunctionReturnVisitor from a type environment
func NewFunctionReturnVisitor(tenv map[Node]TypeVar) *FunctionReturnVisitor {
	return &FunctionReturnVisitor{tenv: tenv}
}

// Visit records the type variable associated with a return statement or an expression statement
func (v *FunctionReturnVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *ReturnStatement:
		v.returnType = v.tenv[n.Argument]
	case *ExpressionStatement:
		v.returnType = v.tenv[n.Expression]
	}
	return v
}

// Done assigns the most recent return type to a Function Expression
func (v *FunctionReturnVisitor) Done() {
	if n, ok := v.current.(*FunctionExpression); ok {
		n.returnTypeVar = v.returnType
	}
}
