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
	sideEffects []values.Value
}

func NewPackageWithValues(name string, obj values.Object) *Package {
	return &Package{
		name:   name,
		object: obj,
	}
}
func NewPackage(name string) *Package {
	return &Package{
		name:   name,
		object: values.NewObject(),
	}
}
func (p *Package) Copy() *Package {
	object := values.NewObjectWithBacking(p.object.Len())
	p.object.Range(func(k string, v values.Value) {
		object.Set(k, v)
	})
	sideEffects := make([]values.Value, len(p.sideEffects))
	copy(sideEffects, p.sideEffects)
	return &Package{
		name:        p.name,
		object:      object,
		sideEffects: sideEffects,
	}
}
func (p *Package) Name() string {
	return p.name
}
func (p *Package) SideEffects() []values.Value {
	return p.sideEffects
}
func (p *Package) Type() semantic.Type {
	return p.object.Type()
}
func (p *Package) PolyType() semantic.PolyType {
	return p.object.PolyType()
}
func (p *Package) Get(name string) (values.Value, bool) {
	return p.object.Get(name)
}
func (p *Package) Set(name string, v values.Value) {
	p.object.Set(name, v)
}
func (p *Package) Len() int {
	return p.object.Len()
}
func (p *Package) Range(f func(name string, v values.Value)) {
	p.object.Range(f)
}
func (p *Package) IsNull() bool {
	return false
}
func (p *Package) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
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
		val, ok := r.Get(k)
		if !ok || !v.Equal(val) {
			equal = false
			return
		}
	})
	return equal
}

func (p *Package) String() string {
	var builder strings.Builder
	builder.WriteString("pkg{")
	i := 0
	p.Range(func(k string, v values.Value) {
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprintf("%v", v.PolyType()))
		i++
	})
	builder.WriteRune('}')
	return builder.String()
}
