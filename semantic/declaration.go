package semantic

// VariableDeclarationVisitor is a visitor that maps identifiers
// to their respective variable declarations in a semantic graph.
type VariableDeclarationVisitor struct {
	current Node
	scope   *VariableScope
	next    *VariableScope
}

// NewVariableDeclarationVisitor instantiates a new VariableDeclarationVisitor
func NewVariableDeclarationVisitor() *VariableDeclarationVisitor {
	return &VariableDeclarationVisitor{}
}

// Visit maps identifiers back to their respective variable declarations
func (v *VariableDeclarationVisitor) Visit(node Node) Visitor {
	v.current = node
	switch n := node.(type) {
	case *BlockStatement:
		v.scope = v.scope.Nest()
	case *NativeVariableDeclaration:
		v.scope.Set(n.Identifier.Name, n)
	case *IdentifierExpression:
		n.declaration, _ = v.scope.Lookup(n.Name)
	case *FunctionExpression:
		// Create a new variable context
		v.scope = v.scope.Nest()
	case *FunctionParam:
		v.next = v.scope
		// Default parameter values must be evaluated in
		// the context outside of the function definition.
		v.scope = v.scope.Parent()
		n.declaration = &NativeVariableDeclaration{
			Identifier: n.Key,
			Init:       n.Default,
		}
		// Parameter names must be visible in the context of the function body
		v.next.Set(n.Key.Name, n.declaration)
	}
	return v
}

// Done resets the variable context after visiting each node
func (v *VariableDeclarationVisitor) Done() {
	switch v.current.(type) {
	case *BlockStatement:
		// Reset variable context after visiting BlockStatement
		v.scope = v.scope.Parent()
	case *FunctionParam:
		// Set variable context to that of the function body after
		// visiting parameter. This is necessary in order for the
		// parameter names to be visible to the function body.
		v.scope = v.next
	case *FunctionExpression:
		// Reset variable context after visiting FunctionExpression
		v.scope = v.scope.Parent()
	}
}

// VariableScope of the program
type VariableScope struct {
	parent *VariableScope
	vardec map[string]VariableDeclaration
}

// NewVariableScope returns a new variable scope
func NewVariableScope() *VariableScope {
	return &VariableScope{
		vardec: make(map[string]VariableDeclaration, 8)}
}

// Set adds a new binding to the current scope
func (s *VariableScope) Set(name string, dec VariableDeclaration) {
	s.vardec[name] = dec
}

// Lookup returns the variable declaration associated with name in the current scope
func (s *VariableScope) Lookup(name string) (VariableDeclaration, bool) {
	if s == nil {
		return nil, false
	}
	dec, ok := s.vardec[name]
	if !ok {
		return s.parent.Lookup(name)
	}
	return dec, ok
}

// Nest returns a new variable scope whose parent is the current scope
func (s *VariableScope) Nest() *VariableScope {
	return &VariableScope{
		parent: s,
		vardec: make(map[string]VariableDeclaration, 8),
	}
}

// Parent returns the parent scope of the current scope
func (s *VariableScope) Parent() *VariableScope {
	return s.parent
}
