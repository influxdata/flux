package mutable

import (
	"sync/atomic"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

// Array specifies the common interface for all mutable arrays.
type Array interface {
	// AppendNull will append a null value to the array.
	AppendNull()

	// SetNull will set whether the value at the index is null.
	SetNull(i int, v bool)

	// IsNull will return if the value at the given index is null.
	IsNull(i int) bool

	// Cap returns the capacity of the array.
	Cap() int

	// Len returns the current length of the array.
	Len() int

	// NewArray will construct a new arrow array.
	// Once the array is created, the mutable array will
	// be reset.
	NewArray() array.Interface

	// Retain will retain a reference to the builder.
	Retain()

	// Release will release any reference to data buffers.
	Release()

	// Reserve will reserve additional capacity in the array for
	// the number of elements to be appended.
	//
	// This does not change the length of the array, but only the capacity.
	Reserve(n int)

	// Resize will resize the array to the given size. It will potentially
	// shrink the array if the requested size is less than the current size.
	//
	// This will change the length of the array.
	Resize(n int)

	// Swap will swap the values at i and j.
	Swap(i, j int)
}

// arrayBase implements the common base for all mutable
// array implementations.
type arrayBase struct {
	refCount int64
	mem      memory.Allocator
	length   int
}

// Len returns the length of the array.
func (b *arrayBase) Len() int {
	return b.length
}

// Retain will retain a reference to the builder.
func (b *arrayBase) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}

func (b *arrayBase) SetNull(i int, v bool) {
	panic("implement me")
}

func (b *arrayBase) IsNull(i int) bool {
	return false
}
