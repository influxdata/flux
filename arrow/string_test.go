package arrow_test

import (
	"testing"

	arrowlib "github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
)

func BenchmarkStringBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := arrow.NewStringBuilder(&memory.Allocator{})
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append("")
		}
		builder.Release()
	}
}

func BenchmarkArrowBinaryBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := array.NewBinaryBuilder(arrowmemory.NewGoAllocator(), arrowlib.BinaryTypes.String)
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.AppendString("")
		}
		builder.Release()
	}
}
