package arrow_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
)

func BenchmarkBoolBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := arrow.NewBoolBuilder(&memory.Allocator{})
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append(false)
		}
		builder.Release()
	}
}

func BenchmarkArrowBooleanBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := array.NewBooleanBuilder(arrowmemory.NewGoAllocator())
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append(false)
		}
		builder.Release()
	}
}
