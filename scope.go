package flux

import (
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

// Scope is an implementation of interpreter.Scope
type Scope struct {
	parent      *Scope
	values      values.Object
	returnValue values.Value
}

func (s *Scope) Lookup(name string) (values.Value, bool) {
	if s == nil {
		return nil, false
	}
	v, ok := s.values.Get(name)
	if !ok {
		return s.parent.Lookup(name)
	}
	return v, ok
}

func (s *Scope) Set(name string, v values.Value) {
	s.values.Set(name, v)
}

func (s *Scope) Nest() interpreter.Scope {
	c := &Scope{values: values.NewObject()}
	c.parent = s
	return c
}

func (s *Scope) NestWithValues(obj values.Object) interpreter.Scope {
	c := &Scope{values: obj}
	c.parent = s
	return c
}

func (s *Scope) Size() int {
	if s == nil {
		return 0
	}
	return s.values.Len() + s.parent.Size()
}

func (s *Scope) Range(f func(k string, v values.Value)) {
	s.values.Range(f)
	if s.parent != nil {
		s.parent.Range(f)
	}
}

func (s *Scope) SetReturn(v values.Value) {
	s.returnValue = v
}

func (s *Scope) Return() values.Value {
	return s.returnValue
}

func (s *Scope) Copy() *Scope {
	if s == nil {
		return nil
	}
	obj := values.NewObjectWithBacking(s.values.Len())
	s.values.Range(func(k string, v values.Value) {
		obj.Set(k, v)
	})
	return &Scope{
		values: obj,
		parent: s.parent.Copy(),
	}
}
