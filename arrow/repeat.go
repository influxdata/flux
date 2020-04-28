package arrow

import (
	"fmt"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Repeat will construct an arrow array that repeats
// the value n times.
func Repeat(v values.Value, n int, mem memory.Allocator) array.Interface {
	switch v.Type().Nature() {
	case semantic.Int:
		b := array.NewInt64Builder(mem)
		b.Resize(n)
		v := v.Int()
		for i := 0; i < n; i++ {
			b.Append(v)
		}
		return b.NewArray()
	case semantic.UInt:
		b := array.NewUint64Builder(mem)
		b.Resize(n)
		v := v.UInt()
		for i := 0; i < n; i++ {
			b.Append(v)
		}
		return b.NewArray()
	case semantic.Float:
		b := array.NewFloat64Builder(mem)
		b.Resize(n)
		v := v.Float()
		for i := 0; i < n; i++ {
			b.Append(v)
		}
		return b.NewArray()
	case semantic.String:
		b := array.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		b.Resize(n)
		b.ReserveData(n * len(v.Str()))
		v := v.Str()
		for i := 0; i < n; i++ {
			b.AppendString(v)
		}
		return b.NewArray()
	case semantic.Bool:
		b := array.NewBooleanBuilder(mem)
		b.Resize(n)
		v := v.Bool()
		for i := 0; i < n; i++ {
			b.Append(v)
		}
		return b.NewArray()
	case semantic.Time:
		b := array.NewInt64Builder(mem)
		b.Resize(n)
		v := int64(v.Time())
		for i := 0; i < n; i++ {
			b.Append(v)
		}
		return b.NewArray()
	default:
		panic(fmt.Errorf("unknown builder for type: %s", v.Type()))
	}
}
