package arrow

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
)

func NewFloat(vs []float64, alloc *memory.Allocator) *array.Float {
	b := NewFloatBuilder(alloc)
	b.Resize(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewFloatArray()
	b.Release()
	return a
}

func FloatSlice(arr *array.Float, i, j int) *array.Float {
	return Slice(arr, int64(i), int64(j)).(*array.Float)
}

func NewFloatBuilder(a *memory.Allocator) *array.FloatBuilder {
	return array.NewFloatBuilder(a)
}
