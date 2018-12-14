package semantic

import "fmt"

// OrderVarDependencies re-orders the variable assignments in a Flux
// program according to declaration order after dependency analysis.
// For example, given the following Flux package:
//
//    package foo
//
//    a = b + c
//    b = f()
//    c = f()
//    d = 0
//    f = () => d + 1
//
// The equivalent declaration ordering would be as follows:
//
//    package foo
//
//    d = 0
//    f = () => d + 1
//    b = f()
//    c = f()
//    a = b + c
//
func OrderVarDependencies(p *Program, externals []string, imp Importer) (*Program, error) {
	scope := newScope(nil)
	vars := make([]*NativeVariableAssignment, 0, len(p.Body))
	// set external vars as initialized
	for _, name := range externals {
		scope.set(name)
	}
	// set imported namespaces as initialized
	for _, n := range p.Imports {
		if n.As != nil {
			scope.set(n.As.Name)
		}
		pkg, ok := imp.Import(n.Path.Value)
		if !ok {
			return nil, fmt.Errorf("invalid import path")
		}
		scope.set(pkg.Name)
	}
	for _, stmt := range p.Body {
		if n, ok := stmt.(*NativeVariableAssignment); ok {
			vars = append(vars, n)
		}
	}
	// Order var assignments according to declaration
	// after all dependencies have been resolved.
	opt := order(vars, scope)
	if len(vars) != len(opt) {
		return nil, fmt.Errorf("unresolvable dependency")
	}
	var idx int
	for i, stmt := range p.Body {
		if _, ok := stmt.(*NativeVariableAssignment); ok {
			p.Body[i] = opt[idx]
			idx++
		}
	}
	return p, nil
}

func order(vars []*NativeVariableAssignment, ready *scope) []*NativeVariableAssignment {
	opt := make([]*NativeVariableAssignment, 0, len(vars))
	w := newDependencyVisitor(ready)
	var more = true
	for more {
		more = false
		for _, v := range vars {
			name := v.Identifier.Name
			if !w.canInit(v.Init) {
				continue
			} else if !ready.lookup(name) {
				opt = append(opt, v)
				ready.set(name)
				more = true
			}
		}
	}
	return opt
}

type dependencyVisitor struct {
	scope    *scope
	resolved bool
}

func newDependencyVisitor(s *scope) *dependencyVisitor {
	return &dependencyVisitor{
		scope:    s,
		resolved: true,
	}
}

func (v *dependencyVisitor) canInit(node Node) bool {
	Walk(NewScopedVisitor(v), node)
	resolved := v.resolved
	v.resolved = true
	return resolved
}

func (v *dependencyVisitor) Nest() NestingVisitor {
	v.scope = v.scope.nest()
	return v
}

func (v *dependencyVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *FunctionParameters:
		for _, p := range n.List {
			v.scope.set(p.Key.Name)
		}
	case *IdentifierExpression:
		if !v.scope.lookup(n.Name) {
			v.resolved = false
		}
	}
	return v
}

func (v *dependencyVisitor) Done(node Node) {
	switch n := node.(type) {
	case *NativeVariableAssignment:
		if v.resolved {
			v.scope.set(n.Identifier.Name)
		}
	}
}

type scope struct {
	parent *scope
	values map[string]bool
}

func newScope(parent *scope) *scope {
	return &scope{
		parent: parent,
		values: map[string]bool{},
	}
}

func (s *scope) lookup(name string) bool {
	if s == nil {
		return false
	}
	if s.values[name] {
		return true
	}
	return s.parent.lookup(name)
}

func (s *scope) set(name string) {
	s.values[name] = true
}

func (s *scope) nest() *scope {
	return newScope(s)
}
