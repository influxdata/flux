package arrow_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
)

func BenchmarkFloatBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := arrow.NewFloatBuilder(&memory.Allocator{})
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append(0.0)
		}
		builder.Release()
	}
}

func BenchmarkArrowFloat64Builder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := array.NewFloat64Builder(arrowmemory.NewGoAllocator())
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append(0.0)
		}
		builder.Release()
	}
}
