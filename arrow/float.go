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
	return Slice(arr, int64(i), int64(j)).(*array.Float64)
}

func NewFloatBuilder(a *memory.Allocator) *array.Float64Builder {
	return array.NewFloat64Builder(a)
}

// AsFloats will return the array as a float array.
// This will panic if the array.Interface does not have the
// float64 datatype.
func AsFloats(arr array.Interface) *array.Float64 {
	if a, ok := arr.(*array.Float64); ok || As(arr, &a) {
		return a
	}
	// Initiate a panic if we could not typecast this.
	return arr.(*array.Float64)
}
