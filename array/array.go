package array

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/bitutil"
	"github.com/apache/arrow-go/v18/arrow/memory"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

//go:generate -command tmpl ../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@binary.tmpldata -o binary.gen.go binary.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o builder.gen.go builder.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o repeat.gen.go repeat.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o conditional.gen.go conditional.gen.go.tmpl
//go:generate tmpl -data=@unary.tmpldata -o unary.gen.go unary.gen.go.tmpl

type DataType = arrow.DataType

var (
	IntType              = arrow.PrimitiveTypes.Int64
	UintType             = arrow.PrimitiveTypes.Uint64
	FloatType            = arrow.PrimitiveTypes.Float64
	StringType           = arrow.BinaryTypes.String
	BooleanType          = arrow.FixedWidthTypes.Boolean
	StringDictionaryType = &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Int32,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}
	StringREEType = arrow.RunEndEncodedOf(arrow.PrimitiveTypes.Int32, arrow.BinaryTypes.String)
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

// String holds an array of flux string values. The arrow data must be
// either a `utf8`, a `dictionary<value=utf8, indices=int32, ordered=false>`,
// or a `run_end_encoded<run_ends:int32, values:utf8>`.
// Internally the string data is stored in an array.Binary value.
type String struct {
	refCount        int64
	data            *array.Data
	nullBitmapBytes []byte

	indices *array.Int32
	runEnds *array.Int32
	values  *array.Binary
}

// Create a new String array from an arrow.ArrayData that contains
// either a `utf8`, a `dictionary<values=utf8, indices=int32, ordered=false>`,
// or a `run_end_encoded<run_ends:int32, values:utf8>` set of data
// buffers. NewStringData will panic if the array data is of an
// unsupported type.
func NewStringData(data arrow.ArrayData) *String {
	a := String{
		refCount: 1,
	}
	a.setData(data.(*array.Data))
	return &a
}

// isStringDataType checks if the given arrow.DataType is a string type
// supported by flux.
func isStringDataType(dt arrow.DataType) bool {
	switch dt := dt.(type) {
	case *arrow.DictionaryType:
		if dt.IndexType.ID() == arrow.INT32 && dt.ValueType.ID() == arrow.STRING {
			return true
		}
	case *arrow.RunEndEncodedType:
		if dt.RunEnds().ID() == arrow.INT32 && dt.Encoded().ID() == arrow.STRING {
			return true
		}
	default:
		if dt.ID() == arrow.STRING {
			return true
		}
	}
	return false
}

// validateStringDataType checks that the datatype is supported for
// using to create a String array.
func validateStringDataType(dt arrow.DataType) {
	if isStringDataType(dt) {
		return
	}
	panic(errors.Newf(codes.Internal, "incorrect data type for String (%s)", dt))
}

func (a *String) setData(data *array.Data) {
	validateStringDataType(data.DataType())
	data.Retain()

	if a.data != nil {
		a.data.Release()
		a.data = nil
		a.nullBitmapBytes = nil
	}
	buffers := data.Buffers()
	if len(buffers) > 0 && buffers[0] != nil {
		a.nullBitmapBytes = buffers[0].Bytes()
	}

	var indices *array.Int32
	var runEnds *array.Int32
	var values *array.Binary

	if data.DataType().ID() == arrow.DICTIONARY {
		idxData := array.NewData(arrow.PrimitiveTypes.Int32, data.Len(), data.Buffers(), nil, data.NullN(), data.Offset())
		indices = array.NewInt32Data(idxData)
		idxData.Release()
		values = array.NewBinaryData(data.Dictionary())
	} else if data.DataType().ID() == arrow.RUN_END_ENCODED {
		runEnds = array.NewInt32Data(data.Children()[0])
		values = array.NewBinaryData(data.Children()[1])
	} else {
		values = array.NewBinaryData(data)
	}
	if a.indices != nil {
		a.indices.Release()
	}
	if a.runEnds != nil {
		a.runEnds.Release()
	}
	if a.values != nil {
		a.values.Release()
	}
	a.indices = indices
	a.runEnds = runEnds
	a.values = values
	a.data = data
}

func (a *String) valuesIndex(i int) (int, bool) {
	if a.indices != nil {
		if a.indices.IsNull(i) {
			return 0, false
		}
		return int(a.indices.Value(i)), true
	} else if a.runEnds != nil {
		return sort.Search(a.runEnds.Len(), func(j int) bool {
			return a.runEnds.Value(j) > int32(i+a.data.Offset())
		}), true
	}
	return i, true
}

func (a *String) DataType() arrow.DataType {
	return a.data.DataType()
}

func (a *String) NullN() int {
	if a.runEnds != nil {
		nbm := a.NullBitmapBytes()
		if nbm == nil {
			return 0
		}
		sz := a.data.Len()
		return sz - bitutil.CountSetBits(nbm, 0, sz)
	}
	return a.data.NullN()
}

func (a *String) NullBitmapBytes() []byte {
	if a.runEnds == nil {
		return a.nullBitmapBytes
	}
	if a.values.NullN() == 0 {
		return nil
	}
	if a.nullBitmapBytes == nil {
		a.nullBitmapBytes = make([]byte, bitutil.BytesForBits(int64(a.data.Len())))
		last := int64(a.data.Offset())
		end := last + int64(a.data.Len())
		for i, _ := a.valuesIndex(0); i < a.runEnds.Len() && last < end; i++ {
			runEnd := int64(a.runEnds.Value(i))
			if runEnd > end {
				runEnd = end
			}
			count := runEnd - last
			bitutil.SetBitsTo(a.nullBitmapBytes, last, count, a.values.IsValid(i))
			last += count
		}
	}
	return a.nullBitmapBytes
}

func (a *String) IsNull(i int) bool {
	if i, ok := a.valuesIndex(i); ok {
		return a.values.IsNull(i)
	}
	return true
}

func (a *String) IsValid(i int) bool {
	if i, ok := a.valuesIndex(i); ok {
		return a.values.IsValid(i)
	}
	return false
}

func (a *String) ValueStr(i int) string {
	if a.IsNull(i) {
		return array.NullValueStr
	}
	return a.Value(i)
}

func (a *String) GetOneForMarshal(i int) interface{} {
	if a.IsNull(i) {
		return nil
	}
	return a.Value(i)
}

func (a *String) MarshalJSON() ([]byte, error) {
	vals := make([]interface{}, a.Len())
	if a.indices != nil {
		for i := 0; i < a.Len(); i++ {
			if a.indices.IsValid(i) {
				idx := int(a.indices.Value(i))
				vals[i] = a.values.ValueString(idx)
			} else {
				vals[i] = nil
			}
		}
	} else {
		for i := 0; i < a.Len(); i++ {
			if a.values.IsValid(i) {
				vals[i] = a.values.ValueString(i)
			} else {
				vals[i] = nil
			}
		}
	}
	return json.Marshal(vals)
}

func (a *String) Data() arrow.ArrayData {
	return a.data
}

func (a *String) Len() int {
	return a.data.Len()
}

func (a *String) Retain() {
	atomic.AddInt64(&a.refCount, 1)
}

func (a *String) Release() {
	if atomic.AddInt64(&a.refCount, -1) == 0 {
		a.nullBitmapBytes = nil
		if a.indices != nil {
			a.indices.Release()
			a.indices = nil
		}
		if a.runEnds != nil {
			a.runEnds.Release()
			a.runEnds = nil
		}
		if a.values != nil {
			a.values.Release()
			a.values = nil
		}
		if a.data != nil {
			a.data.Release()
		}
	}
}

func (a *String) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < a.Len(); i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if a.IsValid(i) {
			fmt.Fprintf(&sb, "%q", a.Value(i))
		} else {
			sb.WriteString(array.NullValueStr)
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

// Value returns a string view of the bytes in the array. The string
// is only valid for the lifetime of the array. Care should be taken not
// to store this string without also retaining the array.
func (a *String) Value(i int) string {
	i, ok := a.valuesIndex(i)
	if !ok {
		// Flux relies on a NULL entry in the String array returning
		// the empty string.
		return ""
	}
	return a.values.ValueString(i)
}

func (a *String) ValueLen(i int) int {
	i, ok := a.valuesIndex(i)
	if !ok {
		// Null values are zero length.
		return 0
	}
	return a.values.ValueLen(i)
}

func (a *String) IsConstant() bool {
	// If all the values are NULL then this is constant.
	if a.data.NullN() == a.data.Len() {
		return true
	}
	// Otherwise if any values are NULL then it can't be constant.
	if a.data.NullN() > 0 {
		return false
	}
	// If values is only 1 item long then it is constant.
	if a.values.Len() == 1 {
		return true
	}

	// Slow method - check all values.
	for i := 1; i < a.Len(); i++ {
		if a.Value(i) != a.Value(i-1) {
			return false
		}
	}
	return true
}

// Slice will construct a new slice of the array using the given
// start and stop index. The returned array must be released.
//
// This is functionally equivalent to using array.NewSlice,
// but array.NewSlice will construct an array.String when
// the data type is a string rather than an array.Binary.
func Slice(arr Array, i, j int) Array {
	data := array.NewSliceData(arr.Data(), int64(i), int64(j))
	defer data.Release()
	return MakeFromData(data)
}

// MakeFromData creates a flux Array from the given data. This will
// panic if the data type that is not understood as a flux array type.
func MakeFromData(data arrow.ArrayData) Array {
	switch data.DataType().ID() {
	case arrow.BOOL:
		return array.NewBooleanData(data)
	case arrow.FLOAT64:
		return array.NewFloat64Data(data)
	case arrow.INT64:
		return array.NewInt64Data(data)
	case arrow.UINT64:
		return array.NewUint64Data(data)
	case arrow.STRING:
		return NewStringData(data)
	case arrow.DICTIONARY, arrow.RUN_END_ENCODED:
		if isStringDataType(data.DataType()) {
			return NewStringData(data)
		}
	}
	panic(errors.Newf(codes.Internal, "invalid data type for flux array (%s)", data.DataType()))
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
	switch vec := arr.(type) {
	case *String:
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
	case *Int:
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
			} else {
				conv.Append(float64(vec.Value(i)))
			}
		}
	case *Uint:
		for i := 0; i < size; i++ {
			if vec.IsNull(i) {
				conv.AppendNull()
			} else {
				conv.Append(float64(vec.Value(i)))
			}
		}
	case *Boolean:
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
