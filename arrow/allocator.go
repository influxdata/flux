package arrow

import (
	arrowmemory "github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func NewAllocator(a memory.Allocator) arrowmemory.Allocator {
	return a
}
