package array

import (
	"strconv"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
)

//go:generate -command tmpl ../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@binary.tmpldata -o binary.gen.go binary.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o builder.gen.go builder.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o repeat.gen.go repeat.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o conditional.gen.go conditional.gen.go.tmpl
//go:generate tmpl -data=@unary.tmpldata -o unary.gen.go unary.gen.go.tmpl

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

	Data() arrow.ArrayData

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

type binaryArray interface {
	NullN() int
	NullBitmapBytes() []byte
	IsNull(i int) bool
	IsValid(i int) bool
	Data() arrow.ArrayData
	Len() int
	ValueBytes() []byte
	ValueLen(i int) int
	ValueOffset(i int) int
	ValueString(i int) string
	Retain()
	Release()
}

type String struct {
	binaryArray
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
		binaryArray: data,
	}
}

func (a *String) DataType() DataType {
	return StringType
}

func (a *String) Slice(i, j int) Array {
	slice, ok := a.binaryArray.(interface{ Slice(i, j int) binaryArray })
	if ok {
		return &String{binaryArray: slice.Slice(i, j)}
	}
	data := array.NewSliceData(a.binaryArray.Data(), int64(i), int64(j))
	defer data.Release()
	return &String{
		binaryArray: array.NewBinaryData(data),
	}
}

// Value returns a string view of the bytes in the array. The string
// is only valid for the lifetime of the array. Care should be taken not
// to store this string without also retaining the array.
func (a *String) Value(i int) string {
	return a.ValueString(i)
}

func (a *String) IsConstant() bool {
	ic, ok := a.binaryArray.(interface{ IsConstant() bool })
	return ok && ic.IsConstant()
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

func ToFloatConv(mem memory.Allocator, arr Array) (*Float, error) {

	// Skip building a new array if the incoming array is already floats
	if fa, ok := arr.(*Float); ok {
		// For any other input type case, we create a brand new array.
		// This implies the caller is responsible for releasing the input array.
		// Tick up the refcount before handing the array right back to the caller
		// to avoid a use-after-free situation.
		fa.Retain()
		return fa, nil
	}

	conv := NewFloatBuilder(mem)
	defer conv.Release()

	size := arr.Len()
	conv.Resize(size)

	// n.b. we handle the arrow.FLOAT64 case at the top of this func so we don't
	// have to handle it here in this switch.
	switch arr.DataType().ID() {
	case arrow.STRING:
		vec := arr.(*String)
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
				continue
			}

			val, err := strconv.ParseFloat(vec.Value(i), 64)
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "cannot convert string %q to Float due to invalid syntax", vec.Value(i))
			}
			conv.Append(val)
		}
	case arrow.INT64:
		vec := arr.(*Int)
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
			} else {
				conv.Append(float64(vec.Value(i)))
			}
		}
	case arrow.UINT64:
		vec := arr.(*Uint)
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
			} else {
				conv.Append(float64(vec.Value(i)))
			}
		}
	case arrow.BOOL:
		vec := arr.(*Boolean)
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
			} else if vec.Value(i) {
				conv.Append(float64(1))
			} else {
				conv.Append(float64(0))
			}
		}
	default:
		return nil, errors.Newf(codes.Invalid, "cannot convert %v to Float", arr.DataType().Name())
	}

	return conv.NewFloatArray(), nil
}
