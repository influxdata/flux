package arrow_test

import (
	"testing"

	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func BenchmarkTimeBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := arrow.NewTimeBuilder(&memory.Allocator{})
		builder.Reserve(1000)
		for j := 0; j < 1000; j++ {
			builder.Append(values.Time(0))
		}
		builder.Release()
	}
}
