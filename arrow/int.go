package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/memory"
)

func NewInt(vs []int64, alloc *memory.Allocator) *array.Int64 {
	b := NewIntBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewInt64Array()
	b.Release()
	return a
}

func IntSlice(arr *array.Int64, i, j int) *array.Int64 {
	return Slice(arr, i, j).(*array.Int64)
}

func NewIntBuilder(a *memory.Allocator) *array.Int64Builder {
	return array.NewInt64Builder(a)
}
