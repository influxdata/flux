package array

import (
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

//go:generate -command tmpl ../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@numeric.tmpldata -o numeric.gen.go numeric.gen.go.tmpl

type DataType = arrow.DataType

var (
	IntType     = arrow.PrimitiveTypes.Int64
	UintType    = arrow.PrimitiveTypes.Uint64
	FloatType   = arrow.PrimitiveTypes.Float64
	StringType  = arrow.BinaryTypes.String
	BooleanType = arrow.FixedWidthTypes.Boolean
)

// Interface represents an immutable sequence of values.
//
// This type is derived from the arrow array.Interface interface.
type Interface interface {
	// DataType returns the type metadata for this instance.
	DataType() DataType

	// NullN returns the number of null values in the array.
	NullN() int

	// NullBitmapBytes returns a byte slice of the validity bitmap.
	NullBitmapBytes() []byte

	// IsNull returns true if value at index is null.
	// NOTE: IsNull will panic if NullBitmapBytes is not empty and 0 > i ≥ Len.
	IsNull(i int) bool

	// IsValid returns true if value at index is not null.
	// NOTE: IsValid will panic if NullBitmapBytes is not empty and 0 > i ≥ Len.
	IsValid(i int) bool

	// Len returns the number of elements in the array.
	Len() int

	// Retain increases the reference count by 1.
	// Retain may be called simultaneously from multiple goroutines.
	Retain()

	// Release decreases the reference count by 1.
	// Release may be called simultaneously from multiple goroutines.
	// When the reference count goes to zero, the memory is freed.
	Release()
}

// Builder provides an interface to build arrow arrays.
//
// This type is derived from the arrow array.Builder interface.
type Builder interface {
	// Retain increases the reference count by 1.
	// Retain may be called simultaneously from multiple goroutines.
	Retain()

	// Release decreases the reference count by 1.
	Release()

	// Len returns the number of elements in the array builder.
	Len() int

	// Cap returns the total number of elements that can be stored
	// without allocating additional memory.
	Cap() int

	// NullN returns the number of null values in the array builder.
	NullN() int

	// AppendNull adds a new null value to the array being built.
	AppendNull()

	// Reserve ensures there is enough space for appending n elements
	// by checking the capacity and calling Resize if necessary.
	Reserve(n int)

	// Resize adjusts the space allocated by b to n elements. If n is greater than b.Cap(),
	// additional memory will be allocated. If n is smaller, the allocated memory may reduced.
	Resize(n int)

	// NewArray creates a new array from the memory buffers used
	// by the builder and resets the Builder so it can be used to build
	// a new array.
	NewArray() Interface
}

type Boolean = array.Boolean

type BooleanBuilder struct {
	b *array.BooleanBuilder
}

func NewBooleanBuilder(mem memory.Allocator) *BooleanBuilder {
	return &BooleanBuilder{
		b: array.NewBooleanBuilder(mem),
	}
}
func (b *BooleanBuilder) Retain() {
	b.b.Retain()
}
func (b *BooleanBuilder) Release() {
	b.b.Release()
}
func (b *BooleanBuilder) Len() int {
	return b.b.Len()
}
func (b *BooleanBuilder) Cap() int {
	return b.b.Cap()
}
func (b *BooleanBuilder) NullN() int {
	return b.b.NullN()
}
func (b *BooleanBuilder) Append(v bool) {
	b.b.Append(v)
}
func (b *BooleanBuilder) AppendValues(v []bool, valid []bool) {
	b.b.AppendValues(v, valid)
}
func (b *BooleanBuilder) UnsafeAppend(v bool) {
	b.b.UnsafeAppend(v)
}
func (b *BooleanBuilder) AppendNull() {
	b.b.AppendNull()
}
func (b *BooleanBuilder) UnsafeAppendBoolToBitmap(isValid bool) {
	b.b.UnsafeAppendBoolToBitmap(isValid)
}
func (b *BooleanBuilder) Reserve(n int) {
	b.b.Reserve(n)
}
func (b *BooleanBuilder) Resize(n int) {
	b.b.Resize(n)
}
func (b *BooleanBuilder) NewArray() Interface {
	return b.NewBooleanArray()
}
func (b *BooleanBuilder) NewBooleanArray() *Boolean {
	return b.b.NewBooleanArray()
}

type String struct {
	value  string
	length int
	data   *array.Binary
}

func (a *String) DataType() DataType {
	return StringType
}
func (a *String) NullN() int {
	if a.data != nil {
		return a.data.NullN()
	}
	return 0
}
func (a *String) NullBitmapBytes() []byte {
	if a.data != nil {
		return a.data.NullBitmapBytes()
	}
	return nil
}
func (a *String) IsNull(i int) bool {
	if a.data != nil {
		return a.data.IsNull(i)
	}
	return false
}
func (a *String) IsValid(i int) bool {
	if a.data != nil {
		return a.data.IsValid(i)
	}
	return true
}
func (a *String) Len() int {
	if a.data != nil {
		return a.data.Len()
	}
	return a.length
}
func (a *String) Retain() {
	if a.data != nil {
		a.data.Retain()
	}
}
func (a *String) Release() {
	if a.data != nil {
		a.data.Release()
	}
}
func (a *String) Slice(i, j int) Interface {
	if a.data != nil {
		data := array.NewSliceData(a.data.Data(), int64(i), int64(j))
		defer data.Release()
		return &String{
			data: array.NewBinaryData(data),
		}
	}
	return &String{
		value:  a.value,
		length: j - i,
	}
}
func (a *String) Value(i int) string {
	if a.data != nil {
		return a.data.ValueString(i)
	}
	return a.value
}
func (a *String) ValueLen(i int) int {
	if a.data != nil {
		return a.data.ValueLen(i)
	}
	return len(a.value)
}

type StringBuilder struct {
	builder      *array.BinaryBuilder
	mem          memory.Allocator
	value        string
	length       int
	capacity     int
	dataCapacity int
	refCount     int
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:      mem,
		refCount: 1,
	}
}
func (b *StringBuilder) init() {
	if b.builder == nil {
		if b.refCount <= 0 {
			return
		}

		builder := array.NewBinaryBuilder(b.mem, StringType)
		if capacity := b.Cap(); capacity > 0 {
			builder.Resize(capacity)
			dataCapacity := len(b.value) * capacity
			if dataCapacity < b.dataCapacity {
				dataCapacity = b.dataCapacity
			}
			builder.ReserveData(dataCapacity)
		}
		if b.length > 0 {
			for i := 0; i < b.length; i++ {
				builder.AppendString(b.value)
			}
		}

		for i := 1; i < b.refCount; i++ {
			builder.Retain()
		}
		b.builder = builder
	}
}
func (b *StringBuilder) reset() {
	b.builder = nil
	b.length = 0
	b.capacity = 0
	b.dataCapacity = 0
	b.value = ""
}
func (b *StringBuilder) Retain() {
	if b.builder != nil {
		b.builder.Retain()
		return
	}
	b.refCount++
}
func (b *StringBuilder) Release() {
	if b.builder != nil {
		b.builder.Release()
		return
	}
	b.refCount--
}
func (b *StringBuilder) Len() int {
	if b.builder != nil {
		return b.builder.Len()
	}
	return b.length
}
func (b *StringBuilder) Cap() int {
	if b.builder != nil {
		return b.builder.Cap()
	}

	capacity := b.capacity
	if capacity < b.length {
		capacity = b.length
	}
	return capacity
}
func (b *StringBuilder) NullN() int {
	if b.builder != nil {
		return b.builder.NullN()
	}
	return 0
}
func (b *StringBuilder) Append(v string) {
	if b.builder == nil && (b.length == 0 || v == b.value) {
		b.value = v
		b.length++
		return
	}
	b.init()
	b.builder.AppendString(v)
}
func (b *StringBuilder) AppendValues(v []string, valid []bool) {
	for i, val := range v {
		if len(valid) != 0 && valid[i] {
			b.AppendNull()
			continue
		}
		b.Append(val)
	}
}
func (b *StringBuilder) AppendNull() {
	b.init()
	b.builder.AppendNull()
}
func (b *StringBuilder) UnsafeAppendBoolToBitmap(isValid bool) {
	b.init()
	b.builder.UnsafeAppendBoolToBitmap(isValid)
}
func (b *StringBuilder) Reserve(n int) {
	if b.builder != nil {
		b.builder.Reserve(n)
		return
	}
	b.capacity = n
}
func (b *StringBuilder) ReserveData(n int) {
	if b.builder != nil {
		b.builder.ReserveData(n)
		return
	}
	b.dataCapacity = n
}
func (b *StringBuilder) Resize(n int) {
	if b.builder != nil {
		b.builder.Resize(n)
		return
	}
	// In arrow, resize and reserve both affect
	// the capacity. Neither of them change the
	// length of the built array.
	b.capacity = n
}
func (b *StringBuilder) NewArray() Interface {
	return b.NewStringArray()
}
func (b *StringBuilder) NewStringArray() *String {
	arr := &String{}
	if b.builder == nil {
		arr.value, arr.length = b.value, b.length
	} else {
		arr.data = b.builder.NewBinaryArray()
	}
	b.reset()
	return arr
}

type sliceable interface {
	Slice(i, j int) Interface
}

// Slice will construct a new slice of the array using the given
// start and stop index. The returned array must be released.
//
// This is functionally equivalent to using array.NewSlice,
// but array.NewSlice will construct an array.String when
// the data type is a string rather than an array.Binary.
func Slice(arr Interface, i, j int) Interface {
	if arr, ok := arr.(sliceable); ok {
		return arr.Slice(i, j)
	}
	if arr, ok := arr.(array.Interface); ok {
		return array.NewSlice(arr, int64(i), int64(j))
	}
	err := errors.Newf(codes.Internal, "cannot slice array of type %T", arr)
	panic(err)
}
