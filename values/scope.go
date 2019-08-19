package values

import (
	"fmt"

	"github.com/influxdata/flux/semantic"
)

type Scope interface {
	// Lookup a name in the scope
	Lookup(name string) (Value, bool)

	// LocalLookup a name in current scope only
	LocalLookup(name string) (Value, bool)

	// Bind a variable in the current scope
	Set(name string, v Value)

	// Create a new scope by nesting the current scope
	// If the passed in object is not nil, its values will be added to the new nested scope.
	Nest(Object) Scope

	// Return the parent of the current scope
	Pop() Scope

	// Number of visible names in scope
	Size() int

	// Range over all variable bindings in scope applying f
	Range(f func(k string, v Value))

	// Range over all variable bindings only in the current scope
	LocalRange(f func(k string, v Value))

	// Set the return value of the scope
	SetReturn(Value)

	// Retrieve the return values of the scope
	Return() Value

	// Create a copy of the scope
	Copy() Scope
}

type scope struct {
	parent      Scope
	values      Object
	returnValue Value
}

func NewScope() Scope {
	return &scope{
		values: NewObject(),
	}
}

func NewNestedScope(s Scope, obj Object) Scope {
	if obj == nil {
		obj = NewObject()
	}
	return &scope{
		parent: s,
		values: obj,
	}
}

func (s *scope) Lookup(name string) (Value, bool) {
	v, ok := s.values.Get(name)
	if !ok && s.parent != nil {
		return s.parent.Lookup(name)
	}
	return v, ok
}
func (s *scope) LocalLookup(name string) (Value, bool) {
	return s.values.Get(name)
}

func (s *scope) Set(name string, v Value) {
	s.values.Set(name, v)
}

func (s *scope) Nest(obj Object) Scope {
	return NewNestedScope(s, obj)
}

func (s *scope) Pop() Scope {
	return s.parent
}

func (s *scope) Size() int {
	if s.parent == nil {
		return s.values.Len()
	}
	return s.values.Len() + s.parent.Size()
}

func (s *scope) Range(f func(k string, v Value)) {
	s.values.Range(f)
	if s.parent != nil {
		s.parent.Range(f)
	}
}

func (s *scope) LocalRange(f func(k string, v Value)) {
	s.values.Range(f)
}

func (s *scope) SetReturn(v Value) {
	s.returnValue = v
}

func (s *scope) Return() Value {
	return s.returnValue
}

func (s *scope) Copy() Scope {
	obj := NewObjectWithBacking(s.values.Len())
	s.values.Range(func(k string, v Value) {
		obj.Set(k, v)
	})
	var parent Scope
	if s.parent != nil {
		parent = s.parent.Copy()
	}
	return &scope{
		values: obj,
		parent: parent,
	}
}

// FormattedScope produces a fmt.Formatter for pretty printing a scope.
func FormattedScope(scope Scope) fmt.Formatter {
	return scopeFormatter{scope}
}

type scopeFormatter struct {
	scope Scope
}

func (s scopeFormatter) Format(state fmt.State, _ rune) {
	state.Write([]byte("["))
	for scope := s.scope; scope != nil; scope = scope.Pop() {
		state.Write([]byte("{\n"))
		scope.LocalRange(func(k string, v Value) {
			fmt.Fprintf(state, "%s: %v,\n", k, v)
		})
		state.Write([]byte("} ->"))
	}
	state.Write([]byte("nil ]"))
}

// BuildExternAssignments constructs nested semantic.ExternAssignment nodes mirroring the nested structure of the scope.
func BuildExternAssignments(node semantic.Node, scope Scope) semantic.Node {
	var n = node
	for s := scope; s != nil; s = s.Pop() {
		extern := &semantic.Extern{
			Block: &semantic.ExternBlock{
				Node: n,
			},
		}
		s.LocalRange(func(k string, v Value) {
			extern.Assignments = append(extern.Assignments, &semantic.ExternalVariableAssignment{
				Identifier: &semantic.Identifier{Name: k},
				ExternType: v.PolyType(),
			})
		})
		n = extern
	}
	return n
}
