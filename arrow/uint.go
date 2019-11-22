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
	data := array.NewSliceData(arr.Data(), int64(i), int64(j))
	defer data.Release()
	return array.NewUint64Data(data)
}

func NewUintBuilder(a *memory.Allocator) *array.Uint64Builder {
	return array.NewUint64Builder(a)
}

// AsUints will return the array as an unsigned integer array.
// This will panic if the array.Interface does not have the
// uint64 datatype.
func AsUints(arr array.Interface) *array.Uint64 {
	if a, ok := arr.(*array.Uint64); ok || As(arr, &a) {
		return a
	}
	// Initiate a panic if we could not typecast this.
	return arr.(*array.Uint64)
}
