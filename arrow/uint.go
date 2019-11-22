package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/memory"
)

func NewUint(vs []uint64, alloc *memory.Allocator) *array.Uint64 {
	b := NewUintBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewUint64Array()
	b.Release()
	return a
}

func UintSlice(arr *array.Uint64, i, j int) *array.Uint64 {
	return Slice(arr, i, j).(*array.Uint64)
}

func NewUintBuilder(a *memory.Allocator) *array.Uint64Builder {
	return array.NewUint64Builder(a)
}
