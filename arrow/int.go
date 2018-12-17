package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func NewInt(vs []int64, alloc *memory.Allocator) *array.Int64 {
	b := NewIntBuilder(alloc)
	b.Reserve(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewInt64Array()
	b.Release()
	return a
}

func NewIntBuilder(a *memory.Allocator) *array.Int64Builder {
	return array.NewInt64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
}
