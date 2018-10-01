package semantic

// VariableDeclarationVisitor is a visitor that maps identifiers
// to their respective variable declarations in a semantic graph.
type VariableDeclarationVisitor struct {
	current Node
	scope   *VariableScope
	outer   *VariableScope
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
		v.scope.Set(n.Identifier.Name, n.Identifier)
	case *IdentifierExpression:
		n.identifier, _ = v.scope.Lookup(n.Name)
	case *FunctionExpression:
		// Nest scope and keep referene to outer scope
		v.outer = v.scope
		v.scope = v.scope.Nest()
	case *ParamKeys:
		for _, k := range n.Identifiers {
			v.scope.Set(k.Name, k)
		}
		return nil
	case *ParamValues:
		// outer is no longer the outer scope
		v.outer = v.scope
		v.scope = v.scope.Parent()
	return v
}

// Done resets the variable context after visiting each node
func (v *VariableDeclarationVisitor) Done() {
	switch v.current.(type) {
	case *BlockStatement:
		v.scope = v.scope.Parent()
	case *FunctionExpression:
		v.scope = v.scope.Parent()
	case *FunctionDefaults:
		v.scope = v.outer
	}
}

// VariableScope of the program
type VariableScope struct {
	parent *VariableScope
	// Identifiers in the current scope
	vars map[string]*Identifier
}

// NewVariableScope returns a new variable scope
func NewVariableScope() *VariableScope {
	return &VariableScope{
		vars: make(map[string]*Identifier, 8)}
}

// Set adds a new binding to the current scope
func (s *VariableScope) Set(name string, ident *Identifier) {
	s.vars[name] = ident
}

// Lookup returns the variable declaration associated with name in the current scope
func (s *VariableScope) Lookup(name string) (*Identifier, bool) {
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
		vars: make(map[string]*Identifier, 8),
	}
}

// Parent returns the parent scope of the current scope
func (s *VariableScope) Parent() *VariableScope {
	return s.parent
}
