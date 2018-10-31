package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
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

func (a *Time) Slice(start, stop int) array.BaseRef {
	return a.TimeSlice(start, stop)
}

func (a *Time) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *Time) Value(i int) values.Time {
	return values.Time(a.data.Value(i))
}

func (a *Time) TimeSlice(start, stop int) array.TimeRef {
	panic("implement me")
}

func (a *Time) Free() {
	a.data.Release()
}

func TimeBuilder(a *memory.Allocator) array.TimeBuilder {
	builder := arrow.NewInt64Builder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &timeBuilder{builder: builder}
}

type timeBuilder struct {
	builder *arrow.Int64Builder
}

func (b *timeBuilder) Type() semantic.Type {
	return semantic.Time
}

func (b *timeBuilder) Len() int {
	return b.builder.Len()
}

func (b *timeBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *timeBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *timeBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *timeBuilder) BuildArray() array.Base {
	return b.BuildTimeArray()
}

func (b *timeBuilder) Free() {
	b.builder.Release()
}

func (b *timeBuilder) Append(v values.Time) {
	b.builder.Append(int64(v))
}

func (b *timeBuilder) AppendValues(vs []values.Time, valid ...[]bool) {
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

func (b *timeBuilder) BuildTimeArray() array.Time {
	return &Time{data: b.builder.NewInt64Array()}
}
