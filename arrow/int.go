package arrow

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
)

func NewInt(vs []int64, alloc *memory.Allocator) *array.Int {
	b := NewIntBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewIntArray()
	b.Release()
	return a
}

func IntSlice(arr *array.Int, i, j int) *array.Int {
	return Slice(arr, int64(i), int64(j)).(*array.Int)
}

func NewIntBuilder(a *memory.Allocator) *array.IntBuilder {
	return array.NewIntBuilder(a)
}
