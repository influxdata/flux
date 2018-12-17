package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func NewFloat(vs []float64, alloc *memory.Allocator) *array.Float64 {
	b := NewFloatBuilder(alloc)
	b.Reserve(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewFloat64Array()
	b.Release()
	return a
}

func NewFloatBuilder(a *memory.Allocator) *array.Float64Builder {
	return array.NewFloat64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
}
