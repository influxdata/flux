package executetest

import (
	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"

	"github.com/influxdata/flux/memory"
)

var UnlimitedAllocator = &memory.ResourceAllocator{
	Allocator: arrowmem.DefaultAllocator,
}
