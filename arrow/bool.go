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
	data := array.NewSliceData(arr.Data(), int64(i), int64(j))
	defer data.Release()
	return array.NewBooleanData(data)
}

func NewBoolBuilder(a *memory.Allocator) *array.BooleanBuilder {
	return array.NewBooleanBuilder(a)
}

// AsBools will return the array as a boolean array.
// This will panic if the array.Interface does not have the
// bool datatype.
func AsBools(arr array.Interface) *array.Boolean {
	if a, ok := arr.(*array.Boolean); ok || As(arr, &a) {
		return a
	}
	// Initiate a panic if we could not typecast this.
	return arr.(*array.Boolean)
}
