package interpreter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Package struct {
	name        string
	object      values.Object
	options     values.Object
	sideEffects []SideEffect
}

func NewPackageWithValues(name string, obj values.Object) *Package {
	if obj == nil {
		obj = values.NewObjectWithValues(nil)
	}
	return &Package{
		name:    name,
		options: values.NewObjectWithValues(nil),
		object:  obj,
	}
}

func NewPackage(name string) *Package {
	return NewPackageWithValues(name, nil)
}

func (p *Package) Copy() *Package {
	object := values.NewObject(p.object.Type())
	p.object.Range(func(k string, v values.Value) {
		object.Set(k, v)
	})
	options := values.NewObject(p.options.Type())
	p.options.Range(func(k string, v values.Value) {
		options.Set(k, v)
	})
	sideEffects := make([]SideEffect, len(p.sideEffects))
	copy(sideEffects, p.sideEffects)
	return &Package{
		name:        p.name,
		object:      object,
		options:     options,
		sideEffects: sideEffects,
	}
}
func (p *Package) Name() string {
	return p.name
}
func (p *Package) SideEffects() []SideEffect {
	return p.sideEffects
}
func (p *Package) Type() semantic.MonoType {
	return p.object.Type()
}
func (p *Package) Get(name string) (values.Value, bool) {
	v, ok := p.object.Get(name)
	if !ok {
		return p.options.Get(name)
	}
	return v, true
}
func (p *Package) Set(name string, v values.Value) {
	p.object.Set(name, v)
}
func (p *Package) SetOption(name string, v values.Value) {
	p.options.Set(name, v)
}
func (p *Package) Len() int {
	return p.object.Len()
}
func (p *Package) Range(f func(name string, v values.Value)) {
	p.object.Range(f)
	p.options.Range(f)
}
func (p *Package) IsNull() bool {
	return false
}
func (p *Package) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
}
func (p *Package) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bytes))
}
func (p *Package) Int() int64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Int))
}
func (p *Package) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.UInt))
}
func (p *Package) Float() float64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Float))
}
func (p *Package) Bool() bool {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bool))
}
func (p *Package) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Object, semantic.Time))
}
func (p *Package) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Object, semantic.Duration))
}
func (p *Package) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Object, semantic.Regexp))
}
func (p *Package) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Object, semantic.Array))
}
func (p *Package) Object() values.Object {
	return p
}
func (p *Package) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}
func (p *Package) Equal(rhs values.Value) bool {
	if p.Type() != rhs.Type() {
		return false
	}
	r := rhs.Object()
	if p.Len() != r.Len() {
		return false
	}
	equal := true
	p.Range(func(k string, v values.Value) {
		if !equal {
			return
		}
		val, ok := r.Get(k)
		equal = ok && v.Equal(val)
	})
	return equal
}

func (p *Package) String() string {
	var builder strings.Builder
	builder.WriteString("pkg{")
	i := 0
	p.Range(func(k string, v values.Value) {
		if _, ok := v.(*Package); ok {
			return
		}
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprintf("%v", v.Type()))
		i++
	})
	builder.WriteRune('}')
	return builder.String()
}
