package arrow

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
)

func NewUint(vs []uint64, alloc *memory.Allocator) *array.Uint {
	b := NewUintBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewUintArray()
	b.Release()
	return a
}

func UintSlice(arr *array.Uint, i, j int) *array.Uint {
	return Slice(arr, int64(i), int64(j)).(*array.Uint)
}

func NewUintBuilder(a *memory.Allocator) *array.UintBuilder {
	return array.NewUintBuilder(a)
}
