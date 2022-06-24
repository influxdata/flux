package arrow

import (
	arrowmemory "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/mvn-trinhnguyen2-dn/flux/memory"
)

func NewAllocator(a memory.Allocator) arrowmemory.Allocator {
	return a
}
