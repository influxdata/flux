package array

import (
	"github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

//go:generate -command tmpl ../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@binary.tmpldata -o binary.gen.go binary.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o builder.gen.go builder.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o repeat.gen.go repeat.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o conditional.gen.go conditional.gen.go.tmpl

type DataType = arrow.DataType

var (
	IntType     = arrow.PrimitiveTypes.Int64
	UintType    = arrow.PrimitiveTypes.Uint64
	FloatType   = arrow.PrimitiveTypes.Float64
	StringType  = arrow.BinaryTypes.String
	BooleanType = arrow.FixedWidthTypes.Boolean
)

// Array represents an immutable sequence of values.
//
// This type is derived from the arrow.Array interface.
type Array interface {
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
	NewArray() Array
}

type String struct {
	value  string
	length int
	data   *array.Binary
}

// NewStringFromBinaryArray creates an instance of String from
// an Arrow Binary array.
//
// Note: Generally client code should be using the types for arrays defined in Flux.
// This method allows string data created outside of Flux (such as from Arrow Flight)
// to be used in Flux.
func NewStringFromBinaryArray(data *array.Binary) *String {
	data.Retain()
	return &String{
		data: data,
	}
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
func (a *String) Slice(i, j int) Array {
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
func (a *String) IsConstant() bool {
	return a.data == nil
}

type sliceable interface {
	Slice(i, j int) Array
}

// Slice will construct a new slice of the array using the given
// start and stop index. The returned array must be released.
//
// This is functionally equivalent to using array.NewSlice,
// but array.NewSlice will construct an array.String when
// the data type is a string rather than an array.Binary.
func Slice(arr Array, i, j int) Array {
	if arr, ok := arr.(sliceable); ok {
		return arr.Slice(i, j)
	}
	if arr, ok := arr.(arrow.Array); ok {
		return array.NewSlice(arr, int64(i), int64(j))
	}
	err := errors.Newf(codes.Internal, "cannot slice array of type %T", arr)
	panic(err)
}
