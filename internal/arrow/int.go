package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Int64 struct {
	data *arrow.Int64
}

func (a *Int64) Type() semantic.Type {
	return semantic.Int
}

func (a *Int64) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Int64) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Int64) Len() int {
	return a.data.Len()
}

func (a *Int64) NullN() int {
	return a.data.NullN()
}

func (a *Int64) Slice(start, stop int) array.BaseRef {
	return a.IntSlice(start, stop)
}

func (a *Int64) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *Int64) Value(i int) int64 {
	return a.data.Value(i)
}

func (a *Int64) IntSlice(start, stop int) array.IntRef {
	panic("implement me")
}

func (a *Int64) Int64Values() []int64 {
	return a.data.Int64Values()
}

func (a *Int64) Free() {
	a.data.Release()
}

func IntBuilder(a *memory.Allocator) array.IntBuilder {
	builder := arrow.NewInt64Builder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &intBuilder{builder: builder}
}

type intBuilder struct {
	builder *arrow.Int64Builder
}

func (b *intBuilder) Type() semantic.Type {
	return semantic.Int
}

func (b *intBuilder) Len() int {
	return b.builder.Len()
}

func (b *intBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *intBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *intBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *intBuilder) BuildArray() array.Base {
	return b.BuildIntArray()
}

func (b *intBuilder) Free() {
	b.builder.Release()
}

func (b *intBuilder) Append(v int64) {
	b.builder.Append(v)
}

func (b *intBuilder) AppendValues(vs []int64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *intBuilder) BuildIntArray() array.Int {
	return &Int64{data: b.builder.NewInt64Array()}
}
