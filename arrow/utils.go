package arrow

import (
	"fmt"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// NewBuilder constructs a new builder for the given
// column type. The allocator passed in must be non-nil.
func NewBuilder(typ flux.ColType, mem memory.Allocator) array.Builder {
	switch typ {
	case flux.TInt, flux.TTime:
		return array.NewIntBuilder(mem)
	case flux.TUInt:
		return array.NewUintBuilder(mem)
	case flux.TFloat:
		return array.NewFloatBuilder(mem)
	case flux.TString:
		return array.NewStringBuilder(mem)
	case flux.TBool:
		return array.NewBooleanBuilder(mem)
	default:
		panic(fmt.Errorf("unknown builder for type: %s", typ))
	}
}

// AppendValue will append a value to the builder.
//
// Be aware when using this function that it will perform
// more slowly than type switching the builder to its
// appropriate type and appending multiple values in a row.
func AppendValue(b array.Builder, v values.Value) error {
	if v.IsNull() {
		b.AppendNull()
		return nil
	}

	switch v.Type().Nature() {
	case semantic.Int:
		return AppendInt(b, v.Int())
	case semantic.UInt:
		return AppendUint(b, v.UInt())
	case semantic.Float:
		return AppendFloat(b, v.Float())
	case semantic.String:
		return AppendString(b, v.Str())
	case semantic.Bool:
		return AppendBool(b, v.Bool())
	case semantic.Time:
		return AppendTime(b, v.Time())
	default:
		panic(fmt.Errorf("unknown builder for type: %s", v.Type()))
	}
}

// AppendInt will append an int64 to a compatible builder.
func AppendInt(b array.Builder, v int64) error {
	vb, ok := b.(*array.IntBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TInt)
	}
	vb.Append(v)
	return nil
}

// AppendUint will append a uint64 to a compatible builder.
func AppendUint(b array.Builder, v uint64) error {
	vb, ok := b.(*array.UintBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TUInt)
	}
	vb.Append(v)
	return nil
}

// AppendFloat will append a float64 to a compatible builder.
func AppendFloat(b array.Builder, v float64) error {
	vb, ok := b.(*array.FloatBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TFloat)
	}
	vb.Append(v)
	return nil
}

// AppendString will append a string to a compatible builder.
func AppendString(b array.Builder, v string) error {
	vb, ok := b.(*array.StringBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TString)
	}
	vb.Append(v)
	return nil
}

// AppendBool will append a bool to a compatible builder.
func AppendBool(b array.Builder, v bool) error {
	vb, ok := b.(*array.BooleanBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TBool)
	}
	vb.Append(v)
	return nil
}

// AppendTime will append a Time value to a compatible builder.
func AppendTime(b array.Builder, v values.Time) error {
	vb, ok := b.(*array.IntBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TTime)
	}
	vb.Append(int64(v))
	return nil
}

// Slice will construct a new slice of the array using the given
// start and stop index. The returned array must be released.
//
// This is functionally equivalent to using array.NewSlice,
// but array.NewSlice will construct an array.String when
// the data type is a string rather than an array.Binary.
func Slice(arr array.Interface, i, j int64) array.Interface {
	return array.Slice(arr, int(i), int(j))
}

// Nulls creates an array of entirely nulls.
// It uses the ColType to determine which builder to use.
func Nulls(typ flux.ColType, n int, mem memory.Allocator) array.Interface {
	b := NewBuilder(typ, mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

// Empty constructs an empty array for the given type.
func Empty(typ flux.ColType) array.Interface {
	// Empty arrays do not actually use memory and they do not
	// use the allocator so we safely use the default allocator
	// here instead of requiring a memory allocator to be passed in.
	b := NewBuilder(typ, memory.DefaultAllocator)
	return b.NewArray()
}

// EmptyBuffer properly constructs an empty TableBuffer.
func EmptyBuffer(key flux.GroupKey, cols []flux.ColMeta) TableBuffer {
	buffer := TableBuffer{
		GroupKey: key,
		Columns:  cols,
		Values:   make([]array.Interface, len(cols)),
	}
	for i, col := range buffer.Columns {
		buffer.Values[i] = Empty(col.Type)
	}
	return buffer
}
