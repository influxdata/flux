package values

import (
	"regexp"

	"github.com/influxdata/flux/semantic"
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
	return semantic.NewDynamicType()
}

func (d dynamic) IsNull() bool {
	return d.inner.IsNull()
}

func (d dynamic) Str() string {
	panic(UnexpectedKind(semantic.Dynamic, semantic.String))
}

func (d dynamic) Bytes() []byte {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Bytes))
}

func (d dynamic) Int() int64 {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Int))
}

func (d dynamic) UInt() uint64 {
	panic(UnexpectedKind(semantic.Dynamic, semantic.UInt))
}

func (d dynamic) Float() float64 {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Float))
}

func (d dynamic) Bool() bool {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Bool))
}

func (d dynamic) Time() Time {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Time))
}

func (d dynamic) Duration() Duration {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Duration))
}

func (d dynamic) Regexp() *regexp.Regexp {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Regexp))
}

func (d dynamic) Array() Array {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Array))
}

func (d dynamic) Object() Object {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Object))
}

func (d dynamic) Function() Function {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Function))
}

func (d dynamic) Dict() Dictionary {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Dictionary))
}

func (d dynamic) Dynamic() Dynamic {
	return d
}

func (d dynamic) Vector() Vector {
	panic(UnexpectedKind(semantic.Dynamic, semantic.Vector))
}

func (d dynamic) Equal(v Value) bool {
	dv, ok := v.(Dynamic)
	if !ok {
		return false
	}
	return d.inner.Equal(dv.Inner())
}

func (d dynamic) Retain() {
	d.inner.Retain()
}

func (d dynamic) Release() {
	d.inner.Release()
}
