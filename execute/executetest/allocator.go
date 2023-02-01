package executetest

import (
	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"

	"github.com/influxdata/flux/memory"
)

var UnlimitedAllocator = &memory.ResourceAllocator{
	Allocator: Allocator{},
}

// Allocator is an allocator for use in test. When a buffer is freed the
// contents are overwritten with a predictable pattern to help detect
// use-after-free scenarios.
type Allocator struct{}

func (Allocator) Allocate(size int) []byte {
	return arrowmem.DefaultAllocator.Allocate(size)
}

func (a Allocator) Reallocate(size int, b []byte) []byte {
	b1 := a.Allocate(size)
	copy(b1, b)
	a.Free(b)
	return b1
}

func (a Allocator) Free(b []byte) {
	var pattern = [...]byte{0x00, 0x33, 0xcc, 0xff}
	for i := range b {
		b[i] = pattern[i%len(pattern)]
	}
	arrowmem.DefaultAllocator.Free(b)
}
