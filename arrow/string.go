package arrow

import (
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/memory"
)

func NewString(vs []string, alloc *memory.Allocator) *array.Binary {
	b := NewStringBuilder(alloc)
	b.Resize(len(vs))
	sz := 0
	for _, v := range vs {
		sz += len(v)
	}
	b.ReserveData(sz)
	for _, v := range vs {
		b.AppendString(v)
	}
	a := b.NewBinaryArray()
	b.Release()
	return a
}

func StringSlice(arr *array.Binary, i, j int) *array.Binary {
	return Slice(arr, int64(i), int64(j)).(*array.Binary)
}

func NewStringBuilder(a *memory.Allocator) *array.BinaryBuilder {
	return array.NewBinaryBuilder(a, arrow.BinaryTypes.String)
}
