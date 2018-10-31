package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Boolean struct {
	data *arrow.Boolean
}

func (a *Boolean) Type() semantic.Type {
	return semantic.Bool
}

func (a *Boolean) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Boolean) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Boolean) Len() int {
	return a.data.Len()
}

func (a *Boolean) NullN() int {
	return a.data.NullN()
}

func (a *Boolean) Slice(start, stop int) array.BaseRef {
	return a.BooleanSlice(start, stop)
}

func (a *Boolean) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *Boolean) Value(i int) bool {
	return a.data.Value(i)
}

func (a *Boolean) BooleanSlice(start, stop int) array.BooleanRef {
	panic("implement me")
}

func (a *Boolean) Free() {
	a.data.Release()
}

func BooleanBuilder(a *memory.Allocator) array.BooleanBuilder {
	builder := arrow.NewBooleanBuilder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &booleanBuilder{builder: builder}
}

type booleanBuilder struct {
	builder *arrow.BooleanBuilder
}

func (b *booleanBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *booleanBuilder) Len() int {
	return b.builder.Len()
}

func (b *booleanBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *booleanBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *booleanBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *booleanBuilder) BuildArray() array.Base {
	return b.BuildBooleanArray()
}

func (b *booleanBuilder) Free() {
	b.builder.Release()
}

func (b *booleanBuilder) Append(v bool) {
	b.builder.Append(v)
}

func (b *booleanBuilder) AppendValues(vs []bool, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.Append(v)
	}
}

func (b *booleanBuilder) BuildBooleanArray() array.Boolean {
	return &Boolean{data: b.builder.NewBooleanArray()}
}
