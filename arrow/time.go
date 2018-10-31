package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Time struct {
	data *arrow.Int64
}

func (a *Time) Type() semantic.Type {
	return semantic.Time
}

func (a *Time) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Time) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Time) Len() int {
	return a.data.Len()
}

func (a *Time) NullN() int {
	return a.data.NullN()
}

func (a *Time) Retain() {
	a.data.Retain()
}

func (a *Time) Release() {
	a.data.Release()
}

func (a *Time) Value(i int) values.Time {
	return values.Time(a.data.Value(i))
}

func NewTimeBuilder(a *memory.Allocator) *TimeBuilder {
	builder := arrow.NewInt64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &TimeBuilder{builder: builder}
}

type TimeBuilder struct {
	builder *arrow.Int64Builder
}

func (b *TimeBuilder) Type() semantic.Type {
	return semantic.Time
}

func (b *TimeBuilder) Len() int {
	return b.builder.Len()
}

func (b *TimeBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *TimeBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *TimeBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *TimeBuilder) Release() {
	b.builder.Release()
}

func (b *TimeBuilder) Append(v values.Time) {
	b.builder.Append(int64(v))
}

func (b *TimeBuilder) AppendValues(vs []values.Time, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.Reserve(len(vs))
		for i, v := range vs {
			if valid[0][i] {
				b.builder.UnsafeAppendBoolToBitmap(false)
			} else {
				b.builder.UnsafeAppend(int64(v))
			}
		}
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(int64(v))
	}
}

func (b *TimeBuilder) NewTimeArray() *Time {
	return &Time{data: b.builder.NewInt64Array()}
}
