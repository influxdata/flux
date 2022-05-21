package values

import (
	"github.com/influxdata/flux/semantic"
	"regexp"
)

type Dynamic interface {
	Value
	Inner() Value
}

type dynamic struct {
	inner Value
}

func NewDynamic(inner Value) Dynamic {
	return dynamic{inner: inner}
}

func (d dynamic) Inner() Value {
	return d.inner
}

func (d dynamic) Type() semantic.MonoType {
	// XXX: useful to delegate to the inner type for cases where we need to
	// lookup funcs for operators.
	// Seems like the times where we need to know whether or not the value is
	// dynamic, we could do something different.
	// Assert the type of the `Value` maybe?
	return d.inner.Type()
}

func (d dynamic) IsNull() bool {
	return d.inner.IsNull()
}

func (d dynamic) Str() string {
	return d.inner.Str()
}

func (d dynamic) Bytes() []byte {
	return d.inner.Bytes()

}

func (d dynamic) Int() int64 {
	return d.inner.Int()

}

func (d dynamic) UInt() uint64 {
	return d.inner.UInt()
}

func (d dynamic) Float() float64 {
	return d.inner.Float()

}

func (d dynamic) Bool() bool {
	return d.inner.Bool()

}

func (d dynamic) Time() Time {
	return d.inner.Time()
}

func (d dynamic) Duration() Duration {
	return d.inner.Duration()
}

func (d dynamic) Regexp() *regexp.Regexp {
	return d.inner.Regexp()
}

func (d dynamic) Array() Array {
	return d.inner.Array()
}

func (d dynamic) Object() Object {
	return d.inner.Object()
}

func (d dynamic) Function() Function {
	return d.inner.Function()
}

func (d dynamic) Dict() Dictionary {
	return d.inner.Dict()
}

func (d dynamic) Dynamic() Dynamic {
	return d
}

func (d dynamic) Vector() Vector {
	return d.inner.Vector()
}

func (d dynamic) Equal(v Value) bool {
	return d.inner.Equal(v)
}

func (d dynamic) Retain() {
	d.inner.Retain()
}

func (d dynamic) Release() {
	d.inner.Release()
}
