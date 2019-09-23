package arrow

import (
	"fmt"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
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
		return array.NewInt64Builder(mem)
	case flux.TUInt:
		return array.NewUint64Builder(mem)
	case flux.TFloat:
		return array.NewFloat64Builder(mem)
	case flux.TString:
		return array.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
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

	switch v.Type() {
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
	vb, ok := b.(*array.Int64Builder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TInt)
	}
	vb.Append(v)
	return nil
}

// AppendUint will append a uint64 to a compatible builder.
func AppendUint(b array.Builder, v uint64) error {
	vb, ok := b.(*array.Uint64Builder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TUInt)
	}
	vb.Append(v)
	return nil
}

// AppendFloat will append a float64 to a compatible builder.
func AppendFloat(b array.Builder, v float64) error {
	vb, ok := b.(*array.Float64Builder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TFloat)
	}
	vb.Append(v)
	return nil
}

// AppendString will append a string to a compatible builder.
func AppendString(b array.Builder, v string) error {
	vb, ok := b.(*array.BinaryBuilder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TString)
	}
	vb.AppendString(v)
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
	vb, ok := b.(*array.Int64Builder)
	if !ok {
		return errors.Newf(codes.Internal, "incompatible builder for type %s", flux.TTime)
	}
	vb.Append(int64(v))
	return nil
}
