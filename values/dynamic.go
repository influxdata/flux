package values

import (
	"fmt"
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

// NewDynamic will recursively wrap a Value in Dynamic.
// Note that any Value can be wrapped, but only a subset have user-facing
// means of extraction.
// If you want to produce a user-facing error for certain types, do so in the
// caller.
func NewDynamic(v Value) Dynamic {

	// Both typed and untyled nulls should wrap plainly
	if v.IsNull() && v.Type().Nature() != semantic.Dynamic {
		return dynamic{inner: v}
	}

	switch n := v.Type().Nature(); n {
	// N.b check to see if the incoming value is Dynamic before all else.
	// We want to avoid re-wrapping, and in the case of nulls a check like
	// `Dynamic.IsNull` will report `true` when the inner value is null.
	case semantic.Dynamic:
		return v.Dynamic()
	case
		// Basic types wrap plainly.
		semantic.String,
		semantic.Bytes,
		semantic.Int,
		semantic.UInt,
		semantic.Float,
		semantic.Bool,
		semantic.Time,
		semantic.Duration,
		// Currently this set of types are not well-supported.
		// For now, wrap them like basic types.
		// Callers may not be able to access the inner types in these cases.
		semantic.Regexp,
		semantic.Dictionary,
		semantic.Vector,
		semantic.Stream,
		semantic.Function:
		return dynamic{inner: v}
	// Composite types need to recurse.
	case semantic.Array:
		arr := v.Array()
		elems := make([]Value, arr.Len())
		arr.Range(func(i int, v Value) {
			val := NewDynamic(v)
			elems[i] = val
		})
		return dynamic{
			inner: NewArrayWithBacking(
				semantic.NewArrayType(semantic.NewDynamicType()),
				elems,
			),
		}
	case semantic.Object:
		obj := v.Object()
		o := make(map[string]Value, obj.Len())
		obj.Range(func(k string, v Value) {
			val := NewDynamic(v)
			o[k] = val
		})
		return dynamic{inner: NewObjectWithValues(o)}
	default:
		panic(fmt.Errorf("unexpected nature %v", n))
	}
}
