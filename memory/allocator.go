package memory

import (
	"fmt"
	"sync/atomic"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// DefaultAllocator is the default memory allocator for Flux.
//
// This implements the memory.Allocator interface from arrow.
var DefaultAllocator = memory.DefaultAllocator

var _ memory.Allocator = (*Allocator)(nil)

// Allocator tracks the amount of memory being consumed by a query.
type Allocator struct {
	// Limit is the limit on the amount of memory that this allocator
	// can assign. If this is null, there is no limit.
	Limit *int64

	// Allocator is the underlying memory allocator used to
	// allocate and free memory.
	// If this is unset, the DefaultAllocator is used.
	Allocator memory.Allocator

	bytesAllocated int64
	maxAllocated   int64
}

// Allocate will ensure that the requested memory is available and
// record that it is in use.
func (a *Allocator) Allocate(size int) []byte {
	if a == nil {
		return DefaultAllocator.Allocate(size)
	}

	if size < 0 {
		panic(errors.New(codes.Internal, "cannot allocate negative memory"))
	} else if size == 0 {
		return nil
	}

	// Account for the size requested.
	if err := a.count(size); err != nil {
		panic(err)
	}

	// Allocate the amount of memory.
	// TODO(jsternberg): It's technically possible for this to allocate
	// more memory than we requested. How do we deal with that since we
	// likely want to use that feature?
	alloc := a.allocator()
	return alloc.Allocate(size)
}

func (a *Allocator) Reallocate(size int, b []byte) []byte {
	if a == nil {
		return DefaultAllocator.Reallocate(size, b)
	}

	sizediff := size - cap(b)
	if err := a.Account(sizediff); err != nil {
		panic(err)
	}

	alloc := a.allocator()
	return alloc.Reallocate(size, b)
}

// Account will manually account for the amount of memory being used.
// This is typically used for memory that is allocated outside of the
// Allocator that must be recorded in some way.
func (a *Allocator) Account(size int) error {
	if size == 0 {
		return nil
	}
	return a.count(size)
}

// Allocated returns the amount of currently allocated memory.
func (a *Allocator) Allocated() int64 {
	return atomic.LoadInt64(&a.bytesAllocated)
}

// MaxAllocated reports the maximum amount of allocated memory at any point in the query.
func (a *Allocator) MaxAllocated() int64 {
	return atomic.LoadInt64(&a.maxAllocated)
}

// Free will reduce the amount of memory used by this Allocator.
// In general, memory should be freed using the Reference returned
// by Allocate. Not all code is capable of using this though so this
// method provides a low-level way of releasing the memory without
// using a Reference.
// Free will release the memory associated with the byte slice.
func (a *Allocator) Free(b []byte) {
	if a == nil {
		DefaultAllocator.Free(b)
		return
	}

	size := len(b)

	// Release the memory to the allocator first.
	alloc := a.allocator()
	alloc.Free(b)

	// Release the memory in our accounting.
	atomic.AddInt64(&a.bytesAllocated, int64(-size))
}

func (a *Allocator) count(size int) error {
	var c int64
	if a.Limit != nil {
		// We need to load the current bytes allocated, add to it, and
		// compare if it is greater than the limit. If it is not, we need
		// to modify the bytes allocated.
		for {
			allocated := atomic.LoadInt64(&a.bytesAllocated)
			if want := allocated + int64(size); want > *a.Limit {
				return errors.Wrap(LimitExceededError{
					Limit:     *a.Limit,
					Allocated: allocated,
					Wanted:    want - allocated,
				}, codes.ResourceExhausted)
			} else if atomic.CompareAndSwapInt64(&a.bytesAllocated, allocated, want) {
				c = want
				break
			}
			// We did not succeed at swapping the bytes allocated so try again.
		}
	} else {
		// Otherwise, add the size directly to the bytes allocated and
		// compare and swap to modify the max allocated.
		c = atomic.AddInt64(&a.bytesAllocated, int64(size))
	}

	// Modify the max allocated if the amount we just allocated is greater.
	for max := atomic.LoadInt64(&a.maxAllocated); c > max; max = atomic.LoadInt64(&a.maxAllocated) {
		if atomic.CompareAndSwapInt64(&a.maxAllocated, max, c) {
			break
		}
	}
	return nil
}

// allocator returns the underlying memory.Allocator that should be used.
func (a *Allocator) allocator() memory.Allocator {
	if a.Allocator == nil {
		return DefaultAllocator
	}
	return a.Allocator
}

// LimitExceededError is an error when the allocation limit is exceeded.
type LimitExceededError struct {
	Limit     int64
	Allocated int64
	Wanted    int64
}

func (a LimitExceededError) Error() string {
	return fmt.Sprintf("memory allocation limit reached: limit %d bytes, allocated: %d, wanted: %d", a.Limit, a.Allocated, a.Wanted)
}
