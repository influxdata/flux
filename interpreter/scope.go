package interpreter

import (
	"github.com/influxdata/flux/values"
)

type Scope interface {
	// Lookup a name in the current scope
	Lookup(name string) (values.Value, bool)

	// Bind a variable in the current scope
	Set(name string, v values.Value)

	// Create a new scope by nesting the current scope
	// If the passed in object is not nil, its values will be added to the new nested scope.
	Nest(values.Object) Scope

	// Number of visible names in scope
	Size() int

	// Range over all variable bindings in scope applying f
	Range(f func(k string, v values.Value))

	// Set the return value of the scope
	SetReturn(values.Value)

	// Retrieve the return values of the scope
	Return() values.Value

	// Create a copy of the scope
	Copy() Scope
}

type scope struct {
	parent      Scope
	values      values.Object
	returnValue values.Value
}

func NewScope() Scope {
	return &scope{
		values: values.NewObject(),
	}
}

func NewNestedScope(s Scope, obj values.Object) Scope {
	if obj == nil {
		obj = values.NewObject()
	}
	return &scope{
		parent: s,
		values: obj,
	}
}

func (s *scope) Lookup(name string) (values.Value, bool) {
	v, ok := s.values.Get(name)
	if !ok && s.parent != nil {
		return s.parent.Lookup(name)
	}
	return v, ok
}

func (s *scope) Set(name string, v values.Value) {
	s.values.Set(name, v)
}

func (s *scope) Nest(obj values.Object) Scope {
	return NewNestedScope(s, obj)
}

func (s *scope) Size() int {
	if s.parent == nil {
		return s.values.Len()
	}
	return s.values.Len() + s.parent.Size()
}

func (s *scope) Range(f func(k string, v values.Value)) {
	s.values.Range(f)
	if s.parent != nil {
		s.parent.Range(f)
	}
}

func (s *scope) SetReturn(v values.Value) {
	s.returnValue = v
}

func (s *scope) Return() values.Value {
	return s.returnValue
}

func (s *scope) Copy() Scope {
	obj := values.NewObjectWithBacking(s.values.Len())
	s.values.Range(func(k string, v values.Value) {
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
