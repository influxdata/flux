package executetest

import (
	arrowmem "github.com/apache/arrow-go/v18/arrow/memory"

	"github.com/influxdata/flux/memory"
)

var UnlimitedAllocator = &memory.ResourceAllocator{
	Allocator: arrowmem.DefaultAllocator,
}
