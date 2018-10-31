package arrow

import (
	"github.com/apache/arrow/go/arrow"
	arrowarray "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type String struct {
	data *arrowarray.Binary
}

func (a *String) Type() semantic.Type {
	return semantic.String
}

func (a *String) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *String) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *String) Len() int {
	return a.data.Len()
}

func (a *String) NullN() int {
	return a.data.NullN()
}

func (a *String) Slice(start, stop int) array.BaseRef {
	return a.StringSlice(start, stop)
}

func (a *String) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *String) Value(i int) string {
	return a.data.ValueString(i)
}

func (a *String) StringSlice(start, stop int) array.StringRef {
	panic("implement me")
}

func (a *String) Free() {
	a.data.Release()
}

func StringBuilder(a *memory.Allocator) array.StringBuilder {
	builder := arrowarray.NewBinaryBuilder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	}, arrow.BinaryTypes.String)
	return &stringBuilder{builder: builder}
}

type stringBuilder struct {
	builder *arrowarray.BinaryBuilder
}

func (b *stringBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *stringBuilder) Len() int {
	return b.builder.Len()
}

func (b *stringBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *stringBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *stringBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *stringBuilder) BuildArray() array.Base {
	return b.BuildStringArray()
}

func (b *stringBuilder) Free() {
	b.builder.Release()
}

func (b *stringBuilder) Append(v string) {
	b.builder.AppendString(v)
}

func (b *stringBuilder) AppendValues(vs []string, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendStringValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.AppendString(v)
	}
}

func (b *stringBuilder) BuildStringArray() array.String {
	return &String{data: b.builder.NewBinaryArray()}
}
