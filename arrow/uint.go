package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func NewUint(vs []uint64, alloc *memory.Allocator) *array.Uint64 {
	b := NewUintBuilder(alloc)
	b.Reserve(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewUint64Array()
	b.Release()
	return a
}

func NewUintBuilder(a *memory.Allocator) *array.Uint64Builder {
	return array.NewUint64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
}
