package arrow

import (
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func NewDuration(vs []arrow.Duration, alloc *memory.Allocator) *array.Duration {
	b := NewDurationBuilder(alloc)
	b.Reserve(len(vs))
	for _, v := range vs {
		b.UnsafeAppend(v)
	}
	a := b.NewDurationArray()
	b.Release()
	return a
}

func DurationSlice(arr *array.Duration, i, j int) *array.Duration {
	data := array.NewSliceData(arr.Data(), int64(i), int64(j))
	defer data.Release()
	return array.NewDurationData(data)
}

func NewDurationBuilder(a *memory.Allocator) *array.DurationBuilder {
	var alloc arrowmemory.Allocator = arrowmemory.NewGoAllocator()
	if a != nil {
		alloc = &allocator{
			Allocator: alloc,
			alloc:     a,
		}
	}
	var newdtype *arrow.DurationType
	return array.NewDurationBuilder(alloc, newdtype)
}


