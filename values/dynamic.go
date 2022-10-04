package values

import (
	"regexp"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
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

// WrapDynamic will recursively wrap a Value in Dynamic
func WrapDynamic(v Value) (Dynamic, error) {
	if v.IsNull() {
		return NewDynamic(v), nil
	}
	switch n := v.Type().Nature(); n {
	case semantic.Dynamic:
		return v.Dynamic(), nil // Return as-is

	// Basic types wrap plainly.
	case semantic.String,
		semantic.Bytes,
		semantic.Int,
		semantic.UInt,
		semantic.Float,
		semantic.Bool,
		semantic.Time,
		semantic.Duration:
		return NewDynamic(v), nil

	// The composite types need to recurse.
	case semantic.Array:
		arr := v.Array()
		elems := make([]Value, arr.Len())
		var rangeErr error
		arr.Range(func(i int, v Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
			}
			val, err := WrapDynamic(v)
			if err != nil {
				rangeErr = err
				return
			}
			elems[i] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return NewDynamic(
			NewArrayWithBacking(
				semantic.NewArrayType(semantic.NewDynamicType()),
				elems,
			)), nil
	case semantic.Object:
		obj := v.Object()
		o := make(map[string]Value, obj.Len())
		var rangeErr error
		obj.Range(func(k string, v Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
			}
			val, err := WrapDynamic(v)
			if err != nil {
				rangeErr = err
				return
			}
			o[k] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return NewDynamic(NewObjectWithValues(o)), nil
	// It's possible we could support many of the remaining types but today
	// there aren't good ways to extract the inner value.
	// We'd need to add support for casting dynamic to each.
	default:
		return nil, errors.Newf(
			codes.Invalid,
			"unsupported type for dynamic: %s %s", v.Type().Nature(), v.Type(),
		)
	}
}
