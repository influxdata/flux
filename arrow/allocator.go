package arrow

import (
	"github.com/InfluxCommunity/flux/memory"
	arrowmemory "github.com/apache/arrow/go/v7/arrow/memory"
)

func NewAllocator(a memory.Allocator) arrowmemory.Allocator {
	return a
}
