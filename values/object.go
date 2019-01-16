package values

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/flux/semantic"
)

type Object interface {
	Value
	Get(name string) (Value, bool)
	Set(name string, v Value)
	Len() int
	Range(func(name string, v Value))
}

type object struct {
	values map[string]Value
	poly   semantic.PolyType
	typ    semantic.Type
	mod    bool
}

func NewObject() *object {
	return &object{values: map[string]Value{}}
}
func NewObjectWithValues(values map[string]Value) *object {
	obj := &object{values: values, mod: true}
	obj.updateTypes()
	return obj
}
func NewObjectWithBacking(size int) *object {
	return &object{values: make(map[string]Value, size)}
}

func (o *object) IsNull() bool {
	return false
}
func (o *object) String() string {
	b := new(strings.Builder)
	b.WriteString("{")
	i := 0
	o.Range(func(k string, v Value) {
		if i != 0 {
			b.WriteString(", ")
		}
		i++
		b.WriteString(k)
		b.WriteString(": ")
		fmt.Fprint(b, v)
	})
	b.WriteString("}")
	return b.String()
}

func (o *object) updateTypes() {
	if !o.mod {
		return
	}
	l := len(o.values)
	ts := make(map[string]semantic.Type, l)
	ps := make(map[string]semantic.PolyType, l)
	ls := make(semantic.LabelSet, 0, l)
	for k, v := range o.values {
		ts[k] = v.Type()
		ps[k] = v.PolyType()
		ls = append(ls, k)
	}
	o.poly = semantic.NewObjectPolyType(ps, nil, ls)
	o.typ = semantic.NewObjectType(ts)
	o.mod = false
}

func (o *object) Type() semantic.Type {
	o.updateTypes()
	return o.typ
}

func (o *object) PolyType() semantic.PolyType {
	o.updateTypes()
	return o.poly
}

func (o *object) Set(name string, v Value) {
	o.values[name] = v
	o.mod = true
}
func (o *object) Get(name string) (Value, bool) {
	v, ok := o.values[name]
	return v, ok
}
func (o *object) Len() int {
	return len(o.values)
}

func (o *object) Range(f func(name string, v Value)) {
	for k, v := range o.values {
		f(k, v)
	}
}

func (o *object) Str() string {
	panic(UnexpectedKind(semantic.Object, semantic.String))
}
func (o *object) Int() int64 {
	panic(UnexpectedKind(semantic.Object, semantic.Int))
}
func (o *object) UInt() uint64 {
	panic(UnexpectedKind(semantic.Object, semantic.UInt))
}
func (o *object) Float() float64 {
	panic(UnexpectedKind(semantic.Object, semantic.Float))
}
func (o *object) Bool() bool {
	panic(UnexpectedKind(semantic.Object, semantic.Bool))
}
func (o *object) Time() Time {
	panic(UnexpectedKind(semantic.Object, semantic.Time))
}
func (o *object) Duration() Duration {
	panic(UnexpectedKind(semantic.Object, semantic.Duration))
}
func (o *object) Regexp() *regexp.Regexp {
	panic(UnexpectedKind(semantic.Object, semantic.Regexp))
}
func (o *object) Array() Array {
	panic(UnexpectedKind(semantic.Object, semantic.Array))
}
func (o *object) Object() Object {
	return o
}
func (o *object) Function() Function {
	panic(UnexpectedKind(semantic.Object, semantic.Function))
}
func (o *object) Equal(rhs Value) bool {
	if o.Type() != rhs.Type() {
		return false
	}
	r := rhs.Object()
	if o.Len() != r.Len() {
		return false
	}
	for k, v := range o.values {
		val, ok := r.Get(k)
		if !ok || !v.Equal(val) {
			return false
		}
	}
	return true
}
