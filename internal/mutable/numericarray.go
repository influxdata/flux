package mutable

import (
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

// Int64Array is an array of int64 values.
type Int64Array struct {
	arrayBase
	data    *memory.Buffer
	rawData []int64
}

// NewInt64Array constructs a new Int64Array.
func NewInt64Array(mem memory.Allocator) *Int64Array {
	return &Int64Array{
		arrayBase: arrayBase{
			refCount: 1,
			mem:      mem,
		},
	}
}

// Append will append a value to the array. This will increase
// the length by 1 and may trigger a reallocation if the length
// would go over the current capacity.
func (b *Int64Array) Append(v int64) {
	b.Reserve(1)
	b.rawData = append(b.rawData, v)
	b.length = len(b.rawData)
}

func (b *Int64Array) AppendNull() {
	panic("implement me")
}

// AppendValues will append the given values to the array.
// This will increase the length for the new values and may
// trigger a reallocation if the length would go over the current
// capacity.
func (b *Int64Array) AppendValues(v []int64) {
	b.Reserve(len(v))
	b.rawData = append(b.rawData, v...)
	b.length = len(b.rawData)
}

// Cap returns the capacity of the array.
func (b *Int64Array) Cap() int { return cap(b.rawData) }

// NewArray returns a new array from the data using NewInt64Array.
func (b *Int64Array) NewArray() array.Interface {
	return b.NewInt64Array()
}

// NewInt64Array will construct a new arrow array from the
// buffered data.
//
// This will reset the current array.
func (b *Int64Array) NewInt64Array() *array.Int64 {
	data := array.NewData(
		arrow.PrimitiveTypes.Int64,
		len(b.rawData),
		[]*memory.Buffer{nil, b.data},
		nil, 0, 0,
	)
	b.reset()

	a := array.NewInt64Data(data)
	data.Release()
	return a
}

func (b *Int64Array) init() {
	b.data = memory.NewResizableBuffer(b.mem)
}

func (b *Int64Array) reset() {
	b.data.Release()
	b.data = nil
	b.rawData = nil
	b.length = 0
}

// Release will release any reference to data buffers.
func (b *Int64Array) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.data != nil {
			b.reset()
		}
	}
}

// Reserve will reserve additional capacity in the array for
// the number of elements to be appended.
//
// This does not change the length of the array, but only the capacity.
func (b *Int64Array) Reserve(n int) {
	if len(b.rawData)+n > cap(b.rawData) {
		if b.data == nil {
			b.init()
		}
		length := len(b.rawData) + n
		capacity := arrow.Int64Traits.BytesRequired(length)
		b.data.Reserve(capacity)
		b.rawData = arrow.Int64Traits.CastFromBytes(b.data.Buf())[:b.length]
	}
}

// Resize will resize the array to the given size. It will potentially
// shrink the array if the requested size is less than the current size.
//
// This will change the length of the array.
func (b *Int64Array) Resize(n int) {
	if b.data == nil {
		b.init()
	}
	newSize := arrow.Int64Traits.BytesRequired(n)
	b.data.Resize(newSize)
	b.rawData = arrow.Int64Traits.CastFromBytes(b.data.Buf())[:n]
	b.length = n
}

// Value will return the value at index i.
func (b *Int64Array) Value(i int) int64 {
	return b.rawData[i]
}

// Set will set the value at index i.
func (b *Int64Array) Set(i int, v int64) {
	b.rawData[i] = v
}

// Swap will swap the values at i and j.
func (b *Int64Array) Swap(i, j int) {
	b.rawData[i], b.rawData[j] = b.rawData[j], b.rawData[i]
}

// Uint64Array is an array of uint64 values.
type Uint64Array struct {
	arrayBase
	data    *memory.Buffer
	rawData []uint64
}

// NewUint64Array constructs a new Uint64Array.
func NewUint64Array(mem memory.Allocator) *Uint64Array {
	return &Uint64Array{
		arrayBase: arrayBase{
			refCount: 1,
			mem:      mem,
		},
	}
}

// Append will append a value to the array. This will increase
// the length by 1 and may trigger a reallocation if the length
// would go over the current capacity.
func (b *Uint64Array) Append(v uint64) {
	b.Reserve(1)
	b.rawData = append(b.rawData, v)
	b.length = len(b.rawData)
}

func (b *Uint64Array) AppendNull() {
	panic("implement me")
}

// AppendValues will append the given values to the array.
// This will increase the length for the new values and may
// trigger a reallocation if the length would go over the current
// capacity.
func (b *Uint64Array) AppendValues(v []uint64) {
	b.Reserve(len(v))
	b.rawData = append(b.rawData, v...)
	b.length = len(b.rawData)
}

// Cap returns the capacity of the array.
func (b *Uint64Array) Cap() int { return cap(b.rawData) }

// NewArray returns a new array from the data using NewUint64Array.
func (b *Uint64Array) NewArray() array.Interface {
	return b.NewUint64Array()
}

// NewUint64Array will construct a new arrow array from the
// buffered data.
//
// This will reset the current array.
func (b *Uint64Array) NewUint64Array() *array.Uint64 {
	data := array.NewData(
		arrow.PrimitiveTypes.Uint64,
		len(b.rawData),
		[]*memory.Buffer{nil, b.data},
		nil, 0, 0,
	)
	b.reset()

	a := array.NewUint64Data(data)
	data.Release()
	return a
}

func (b *Uint64Array) init() {
	b.data = memory.NewResizableBuffer(b.mem)
}

func (b *Uint64Array) reset() {
	b.data.Release()
	b.data = nil
	b.rawData = nil
	b.length = 0
}

// Release will release any reference to data buffers.
func (b *Uint64Array) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.data != nil {
			b.reset()
		}
	}
}

// Reserve will reserve additional capacity in the array for
// the number of elements to be appended.
//
// This does not change the length of the array, but only the capacity.
func (b *Uint64Array) Reserve(n int) {
	if len(b.rawData)+n > cap(b.rawData) {
		if b.data == nil {
			b.init()
		}
		length := len(b.rawData) + n
		capacity := arrow.Uint64Traits.BytesRequired(length)
		b.data.Reserve(capacity)
		b.rawData = arrow.Uint64Traits.CastFromBytes(b.data.Buf())[:b.length]
	}
}

// Resize will resize the array to the given size. It will potentially
// shrink the array if the requested size is less than the current size.
//
// This will change the length of the array.
func (b *Uint64Array) Resize(n int) {
	if b.data == nil {
		b.init()
	}
	newSize := arrow.Uint64Traits.BytesRequired(n)
	b.data.Resize(newSize)
	b.rawData = arrow.Uint64Traits.CastFromBytes(b.data.Buf())[:n]
	b.length = n
}

// Value will return the value at index i.
func (b *Uint64Array) Value(i int) uint64 {
	return b.rawData[i]
}

// Set will set the value at index i.
func (b *Uint64Array) Set(i int, v uint64) {
	b.rawData[i] = v
}

// Swap will swap the values at i and j.
func (b *Uint64Array) Swap(i, j int) {
	b.rawData[i], b.rawData[j] = b.rawData[j], b.rawData[i]
}

// Float64Array is an array of float64 values.
type Float64Array struct {
	arrayBase
	data    *memory.Buffer
	rawData []float64
}

// NewFloat64Array constructs a new Float64Array.
func NewFloat64Array(mem memory.Allocator) *Float64Array {
	return &Float64Array{
		arrayBase: arrayBase{
			refCount: 1,
			mem:      mem,
		},
	}
}

// Append will append a value to the array. This will increase
// the length by 1 and may trigger a reallocation if the length
// would go over the current capacity.
func (b *Float64Array) Append(v float64) {
	b.Reserve(1)
	b.rawData = append(b.rawData, v)
	b.length = len(b.rawData)
}

func (b *Float64Array) AppendNull() {
	panic("implement me")
}

// AppendValues will append the given values to the array.
// This will increase the length for the new values and may
// trigger a reallocation if the length would go over the current
// capacity.
func (b *Float64Array) AppendValues(v []float64) {
	b.Reserve(len(v))
	b.rawData = append(b.rawData, v...)
	b.length = len(b.rawData)
}

// Cap returns the capacity of the array.
func (b *Float64Array) Cap() int { return cap(b.rawData) }

// NewArray returns a new array from the data using NewFloat64Array.
func (b *Float64Array) NewArray() array.Interface {
	return b.NewFloat64Array()
}

// NewFloat64Array will construct a new arrow array from the
// buffered data.
//
// This will reset the current array.
func (b *Float64Array) NewFloat64Array() *array.Float64 {
	data := array.NewData(
		arrow.PrimitiveTypes.Float64,
		len(b.rawData),
		[]*memory.Buffer{nil, b.data},
		nil, 0, 0,
	)
	b.reset()

	a := array.NewFloat64Data(data)
	data.Release()
	return a
}

func (b *Float64Array) init() {
	b.data = memory.NewResizableBuffer(b.mem)
}

func (b *Float64Array) reset() {
	if b.data != nil {
		b.data.Release()
		b.data = nil
	}
	b.rawData = nil
	b.length = 0
}

// Release will release any reference to data buffers.
func (b *Float64Array) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.data != nil {
			b.reset()
		}
	}
}

// Reserve will reserve additional capacity in the array for
// the number of elements to be appended.
//
// This does not change the length of the array, but only the capacity.
func (b *Float64Array) Reserve(n int) {
	if len(b.rawData)+n > cap(b.rawData) {
		if b.data == nil {
			b.init()
		}
		length := len(b.rawData) + n
		capacity := arrow.Float64Traits.BytesRequired(length)
		b.data.Reserve(capacity)
		b.rawData = arrow.Float64Traits.CastFromBytes(b.data.Buf())[:b.length]
	}
}

// Resize will resize the array to the given size. It will potentially
// shrink the array if the requested size is less than the current size.
//
// This will change the length of the array.
func (b *Float64Array) Resize(n int) {
	if b.data == nil {
		b.init()
	}
	newSize := arrow.Float64Traits.BytesRequired(n)
	b.data.Resize(newSize)
	b.rawData = arrow.Float64Traits.CastFromBytes(b.data.Buf())[:n]
	b.length = n
}

// Value will return the value at index i.
func (b *Float64Array) Value(i int) float64 {
	return b.rawData[i]
}

// Set will set the value at index i.
func (b *Float64Array) Set(i int, v float64) {
	b.rawData[i] = v
}

// Swap will swap the values at i and j.
func (b *Float64Array) Swap(i, j int) {
	b.rawData[i], b.rawData[j] = b.rawData[j], b.rawData[i]
}
