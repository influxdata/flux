package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/memory"
)

func NewBool(vs []bool, alloc *memory.Allocator) *array.Boolean {
	b := NewBoolBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewBooleanArray()
	b.Release()
	return a
}

func BoolSlice(arr *array.Boolean, i, j int) *array.Boolean {
	return Slice(arr, int64(i), int64(j)).(*array.Boolean)
}

func NewBoolBuilder(a *memory.Allocator) *array.BooleanBuilder {
	return array.NewBooleanBuilder(a)
}
