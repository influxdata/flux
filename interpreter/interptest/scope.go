package interptest

import (
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

// Scope is an implementation of interpreter.Scope
type Scope struct {
	parent      *Scope
	values      map[string]values.Value
	returnValue values.Value
}

func NewScope() *Scope {
	return &Scope{
		values: make(map[string]values.Value),
	}
}

func NewScopeWithValues(values map[string]values.Value) *Scope {
	return &Scope{values: values}
}

func (s *Scope) Lookup(name string) (values.Value, bool) {
	if s == nil {
		return nil, false
	}
	v, ok := s.values[name]
	if !ok {
		return s.parent.Lookup(name)
	}
	return v, ok
}

func (s *Scope) Set(name string, value values.Value) {
	s.values[name] = value
}

func (s *Scope) SetReturn(value values.Value) {
	s.returnValue = value
}

func (s *Scope) Return() values.Value {
	return s.returnValue
}

func (s *Scope) Nest() interpreter.Scope {
	c := NewScope()
	c.parent = s
	return c
}

func (s *Scope) Range(f func(k string, v values.Value)) {
	for k, v := range s.values {
		f(k, v)
	}
	if s.parent != nil {
		s.parent.Range(f)
	}
}
func (s *Scope) Size() int {
	if s == nil {
		return 0
	}
	return len(s.values) + s.parent.Size()
}

func (s *Scope) Copy() *Scope {
	if s == nil {
		return nil
	}
	values := make(map[string]values.Value)
	for k, v := range s.values {
		values[k] = v
	}
	return &Scope{
		values: values,
		parent: s.parent.Copy(),
	}
}
