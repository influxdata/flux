package values

import (
	"fmt"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

type Scope interface {
	// Lookup a name in the scope.
	Lookup(name string) (Value, bool)

	// LocalLookup a name in current scope only.
	LocalLookup(name string) (Value, bool)

	// Set binds a variable in the current scope.
	Set(name string, v Value)
	// SetOption binds a variable in the package option scope.
	// Setting an option must occur on the specific package value.
	// If the package cannot be found no option is set, in which case the boolean return is false.
	// An error is reported if the specified package is not a package value.
	SetOption(pkg, name string, v Value) (bool, error)

	// Nest creates a new scope by nesting the current scope.
	// If the passed in object is not nil, its values will be added to the new nested scope.
	Nest(Object) Scope

	// Pop returns the parent of the current scope.
	Pop() Scope

	// Size is the number of visible names in scope.
	Size() int

	// Range iterates over all variable bindings in scope applying f.
	Range(f func(k string, v Value))

	// LocalRange iterates over all variable bindings only in the current scope.
	LocalRange(f func(k string, v Value))

	// SetReturn binds the return value of the scope.
	SetReturn(Value)

	// Return reports the bound return value of the scope.
	Return() Value

	// Copy creates a deep copy of the scope, values are not copied.
	// Copy preserves the nesting structure.
	Copy() Scope
}

type scope struct {
	parent      Scope
	values      Object
	returnValue Value
}

// NewScope creates a new empty scope with no parent.
func NewScope() Scope {
	return &scope{
		values: NewObject(),
	}
}

//NewNestedScope creates a new scope with bindings from obj and a parent.
func NewNestedScope(parent Scope, obj Object) Scope {
	if obj == nil {
		obj = NewObject()
	}
	return &scope{
		parent: parent,
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

func (s *scope) SetOption(pkg, name string, v Value) (bool, error) {
	pv, ok := s.LocalLookup(pkg)
	if !ok {
		parent := s.Pop()
		if parent != nil {
			return parent.SetOption(pkg, name, v)
		}
		return false, nil
	}
	p, ok := pv.(Package)
	if !ok {
		return false, errors.Newf(codes.Invalid, "cannot set option %q is not a package", pkg)
	}
	p.SetOption(name, v)
	return true, nil
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
		state.Write([]byte("{"))
		j := 0
		scope.LocalRange(func(k string, v Value) {
			if j != 0 {
				state.Write([]byte(", "))
			}
			fmt.Fprintf(state, "%s = %v", k, v)
			j++
		})
		state.Write([]byte("} -> "))
	}
	state.Write([]byte("nil ]"))
}
