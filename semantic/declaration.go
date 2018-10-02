package semantic

// VariableDeclarationVisitor is a visitor that maps identifiers
// to their respective variable declarations in a semantic graph.
type VariableDeclarationVisitor struct {
	current Node
	scope   *VariableScope
}

// NewVariableDeclarationVisitor instantiates a new VariableDeclarationVisitor
func NewVariableDeclarationVisitor() *VariableDeclarationVisitor {
	return &VariableDeclarationVisitor{
		scope: NewVariableScope(),
	}
}

func (v *VariableDeclarationVisitor) nest() *VariableDeclarationVisitor {
	return &VariableDeclarationVisitor{
		scope: v.scope.Nest(),
	}
}

// Visit maps identifiers back to their respective variable declarations
func (v *VariableDeclarationVisitor) Visit(node Node) Visitor {
	v.current = node
	switch n := node.(type) {
	case *BlockStatement:
		return v.nest()
	case *FunctionBody:
		return v.nest()
	case *NativeVariableDeclaration:
		v.scope.Set(n.Identifier.Name, n)
	case *FunctionParam:
		v.scope.Set(n.Key.Name, n.Key)
	case *IdentifierExpression:
		n.declaration, _ = v.scope.Lookup(n.Name)
	}
	return v
}
func (v *VariableDeclarationVisitor) Done() {}

// VariableScope of the program
type VariableScope struct {
	parent *VariableScope
	// Identifiers in the current scope
	vars map[string]Node
}

// NewVariableScope returns a new variable scope
func NewVariableScope() *VariableScope {
	return &VariableScope{
		vars: make(map[string]Node, 8),
	}
}

// Set adds a new binding to the current scope
func (s *VariableScope) Set(name string, node Node) {
	s.vars[name] = node
}

// Lookup returns the variable declaration associated with name in the current scope
func (s *VariableScope) Lookup(name string) (Node, bool) {
	if s == nil {
		return nil, false
	}
	dec, ok := s.vars[name]
	if !ok {
		return s.parent.Lookup(name)
	}
	return dec, ok
}

// Nest returns a new variable scope whose parent is the current scope
func (s *VariableScope) Nest() *VariableScope {
	return &VariableScope{
		parent: s,
		vars:   make(map[string]Node, 8),
	}
}

// Parent returns the parent scope of the current scope
func (s *VariableScope) Parent() *VariableScope {
	return s.parent
}
