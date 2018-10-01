package semantic

// FunctionReturnVisitor visits an annotated semantic graph
// and records every function's return type variable.
type FunctionReturnVisitor struct {
	tenv        map[Node]TypeVar
	currentFunc *FunctionExpression
}

// NewFunctionReturnVisitor instantiates a new FunctionReturnVisitor from a type environment
func NewFunctionReturnVisitor(tenv map[Node]TypeVar) *FunctionReturnVisitor {
	return &FunctionReturnVisitor{tenv: tenv}
}

// Visit records the type variable associated with a return statement or an expression statement
func (v *FunctionReturnVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *FunctionExpression:
		v.currentFunc = n
	case *ReturnStatement:
		v.currentFunc.returnTypeVar = v.tenv[n.Argument]
	case *ExpressionStatement:
		v.currentFunc.returnTypeVar = v.tenv[n.Expression]
	}
	return v
}

// Done implements Visitor interface
func (v *FunctionReturnVisitor) Done() {}
