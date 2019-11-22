package arrow

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/memory"
)

func NewFloat(vs []float64, alloc *memory.Allocator) *array.Float64 {
	b := NewFloatBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewFloat64Array()
	b.Release()
	return a
}

func FloatSlice(arr *array.Float64, i, j int) *array.Float64 {
	return Slice(arr, i, j).(*array.Float64)
}

func NewFloatBuilder(a *memory.Allocator) *array.Float64Builder {
	return array.NewFloat64Builder(a)
}
