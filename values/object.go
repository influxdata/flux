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
	labels semantic.LabelSet
	values []Value
	// TODO (algow): reimplement mutable object value
	//ptyp   map[string]semantic.PolyType
	//ptypv  semantic.PolyType
	//mtyp   map[string]semantic.Type
	//mtypv  semantic.Type
}

func NewObject() *object {
	return &object{
		//ptyp: make(map[string]semantic.PolyType),
		//mtyp: make(map[string]semantic.Type),
	}
}
func NewObjectWithValues(vals map[string]Value) *object {
	//l := len(vals)

	//labels := make(semantic.LabelSet, 0, l)
	//values := make([]Value, 0, l)

	//ptyp := make(map[string]semantic.PolyType, l)
	//mtyp := make(map[string]semantic.Type, l)

	//for k, v := range vals {
	//	labels = append(labels, k)
	//	values = append(values, v)

	//	ptyp[k] = v.PolyType()
	//	mtyp[k] = v.Type()
	//}

	//return &object{
	//	labels: labels,
	//	values: values,
	//	ptyp:   ptyp,
	//	mtyp:   mtyp,
	//}
	return nil
}
func NewObjectWithBacking(size int) *object {
	return &object{
		labels: make(semantic.LabelSet, 0, size),
		values: make([]Value, 0, size),
		//ptyp:   make(map[string]semantic.PolyType, size),
		//mtyp:   make(map[string]semantic.Type, size),
	}
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

func (o *object) Type() semantic.MonoType {
	// TODO (algow): finish implementation of object
	return semantic.MonoType{}
}

func (o *object) Set(k string, v Value) {
	//// update type
	//pt := v.PolyType()
	//if o.ptypv != nil {
	//	if optyp, ok := o.ptyp[k]; !ok || !pt.Equal(optyp) {
	//		o.ptyp[k] = pt
	//		o.ptypv = nil
	//	}
	//} else {
	//	o.ptyp[k] = pt
	//}
	//mt := v.Type()
	//if mt == nil {
	//	mt = semantic.Invalid
	//}
	//if o.mtypv != nil {
	//	if omtyp, ok := o.mtyp[k]; !ok || mt != omtyp {
	//		o.mtyp[k] = mt
	//		o.mtypv = nil
	//	}
	//} else {
	//	o.mtyp[k] = mt
	//}

	//// update value
	//for i, l := range o.labels {
	//	if l == k {
	//		o.values[i] = v
	//		return
	//	}
	//}
	//o.labels = append(o.labels, k)
	//o.values = append(o.values, v)
}

func (o *object) Get(name string) (Value, bool) {
	for i, l := range o.labels {
		if name == l {
			return o.values[i], true
		}
	}
	return Null, false
}
func (o *object) Len() int {
	return len(o.values)
}

func (o *object) Range(f func(name string, v Value)) {
	for i, l := range o.labels {
		f(l, o.values[i])
	}
}

func (o *object) Str() string {
	panic(UnexpectedKind(semantic.Object, semantic.String))
}
func (o *object) Bytes() []byte {
	panic(UnexpectedKind(semantic.Object, semantic.Bytes))
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
	for i, l := range o.labels {
		val, ok := r.Get(l)
		if ok && !o.values[i].Equal(val) {
			return false
		}
	}
	return true
}
