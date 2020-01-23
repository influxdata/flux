package interpreter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Package is an importable package that can be used from another
// section of code. The package itself cannot have its attributes
// modified after creation, but the options may be changed.
type Package struct {
	// name is the name of the package.
	name string

	// object contains the object properties of this package.
	object values.Object

	// options contains the option overrides for this package.
	// Values cannot exist in here unless they also exist in
	// the underlying object.
	options map[string]values.Value

	// sideEffects contains the side effects caused by this package.
	// This is currently unused.
	sideEffects []SideEffect
}

func NewPackageWithValues(name string, obj values.Object) *Package {
	return &Package{
		name:   name,
		object: obj,
	}
}

func NewPackage(name string) *Package {
	return NewPackageWithValues(name, nil)
}

func (p *Package) Copy() *Package {
	var options map[string]values.Value
	if len(p.options) > 0 {
		options = make(map[string]values.Value, len(p.options))
		for k, v := range p.options {
			options[k] = v
		}
	}
	var sideEffects []SideEffect
	if len(p.sideEffects) > 0 {
		sideEffects = make([]SideEffect, len(p.sideEffects))
		copy(sideEffects, p.sideEffects)
	}
	return &Package{
		name:        p.name,
		object:      p.object,
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
	if ok && values.IsOption(v) {
		// If this value is an option, the option may have
		// been overriden. Check the override map.
		if ov, ok := p.options[name]; ok {
			v = ov
		}
	}
	return v, ok
}
func (p *Package) Set(name string, v values.Value) {
	panic(errors.New(codes.Internal, "package members cannot be modified"))
}
func (p *Package) SetOption(name string, v values.Value) {
	// TODO(jsternberg): Setting an invalid option on a package wasn't previously
	// an error so it continues to not be an error. We should probably find a way
	// to make it so setting an invalid option is an error.
	if p.options == nil {
		p.options = make(map[string]values.Value)
	}
	p.options[name] = v
}
func (p *Package) Len() int {
	return p.object.Len()
}
func (p *Package) Range(f func(name string, v values.Value)) {
	p.object.Range(func(name string, v values.Value) {
		// Check if the value was overridden.
		if values.IsOption(v) {
			if ov, ok := p.options[name]; ok {
				v = ov
			}
		}
		f(name, v)
	})
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
