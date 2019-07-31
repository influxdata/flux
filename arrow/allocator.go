package arrow

import (
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

type allocator struct {
	arrowmemory.Allocator
	alloc *memory.Allocator
}

func NewAllocator(a *memory.Allocator) arrowmemory.Allocator {
	var alloc arrowmemory.Allocator = arrowmemory.NewGoAllocator()
	if a != nil {
		alloc = &allocator{
			Allocator: alloc,
			alloc:     a,
		}
	}
	return alloc
}

func (a *allocator) Allocate(size int) []byte {
	if err := a.alloc.Allocate(size); err != nil {
		panic(err)
	}
	return a.Allocator.Allocate(size)
}

func (a *allocator) Reallocate(size int, b []byte) []byte {
	sizediff := size - cap(b)
	if sizediff > 0 {
		if err := a.alloc.Allocate(sizediff); err != nil {
			panic(err)
		}
	} else {
		a.alloc.Free(-sizediff)
	}
	return a.Allocator.Reallocate(size, b)
}

func (a *allocator) Free(b []byte) {
	a.alloc.Free(cap(b))
	a.Allocator.Free(b)
}
