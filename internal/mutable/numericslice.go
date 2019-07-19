package mutable

import (
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	fluxarrow "github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
)

// Int64TrackedSlice wraps a slice of int64 values and tracks its memory usage.
type Int64TrackedSlice struct {
	refCount int64
	alloc    *memory.Allocator
	rawData  []int64
}

// NewInt64TrackedSlice constructs a new Int64TrackedSlice given an allocator, length, and capacity.
func NewInt64TrackedSlice(alloc *memory.Allocator, l, c int) *Int64TrackedSlice {
	s := &Int64TrackedSlice{
		refCount: 1,
		alloc:    alloc,
		rawData:  make([]int64, l, c),
	}
	s.allocate()
	return s
}

func (b *Int64TrackedSlice) allocate() {
	bs := arrow.Int64Traits.BytesRequired(cap(b.rawData))
	if err := b.alloc.Allocate(bs); err != nil {
		panic(err)
	}
}

func (b *Int64TrackedSlice) free() {
	bs := arrow.Int64Traits.BytesRequired(cap(b.rawData))
	b.alloc.Free(bs)
}

func (b *Int64TrackedSlice) reset() {
	b.free()
	b.rawData = nil
}

// Append will append a value to the array. This will increase
// the length by 1 and may trigger a reallocation if the length
// would go over the current capacity.
func (b *Int64TrackedSlice) Append(v int64) {
	b.Reserve(1)
	b.rawData = append(b.rawData, v)
}

// AppendValues will append the given values to the array.
// This will increase the length for the new values and may
// trigger a reallocation if the length would go over the current
// capacity.
func (b *Int64TrackedSlice) AppendValues(v []int64) {
	b.Reserve(len(v))
	b.rawData = append(b.rawData, v...)
}

// Cap returns the capacity of the array.
func (b *Int64TrackedSlice) Cap() int { return cap(b.rawData) }

// Len returns the length of the array.
func (b *Int64TrackedSlice) Len() int { return len(b.rawData) }

// Retain will retain a reference to the builder.
func (b *Int64TrackedSlice) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}

// Release will release any reference to data buffers.
func (b *Int64TrackedSlice) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		b.reset()
	}
}

// Reserve will reserve additional capacity in the array for
// the number of elements to be appended.
//
// This does not change the length of the array, but only the capacity.
func (b *Int64TrackedSlice) Reserve(n int) {
	if len(b.rawData)+n > cap(b.rawData) {
		b.free()
		t := make([]int64, len(b.rawData), len(b.rawData)+n)
		copy(t, b.rawData)
		b.rawData = t
		b.allocate()
	}
}

// Resize will resize the array to the given size. It will potentially
// shrink the array if the requested size is less than the current size.
//
// This will change the length of the array.
func (b *Int64TrackedSlice) Resize(n int) {
	b.free()
	t := make([]int64, n)
	copy(t, b.rawData)
	b.rawData = t
	b.allocate()
}

// Value will return the value at index i.
func (b *Int64TrackedSlice) Value(i int) int64 {
	return b.rawData[i]
}

// Set will set the value at index i.
func (b *Int64TrackedSlice) Set(i int, v int64) {
	b.rawData[i] = v
}

// NewArray returns a new array from the data using NewInt64TrackedSlice.
func (b *Int64TrackedSlice) NewArray() array.Interface {
	return b.NewInt64Array()
}

// NewInt64TrackedSlice will construct a new arrow array from the
// buffered data.
//
// This will reset the current array.
func (b *Int64TrackedSlice) NewInt64Array() *array.Int64 {
	builder := array.NewInt64Builder(fluxarrow.NewAllocator(b.alloc))
	// every value is valid (for now).
	builder.AppendValues(b.rawData, nil)
	b.reset()
	return builder.NewInt64Array()
}
