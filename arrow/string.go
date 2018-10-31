package arrow

import (
	"github.com/apache/arrow/go/arrow"
	arrowarray "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
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

func (a *String) Retain() {
	a.data.Retain()
}

func (a *String) Release() {
	a.data.Release()
}

func (a *String) Value(i int) string {
	return a.data.ValueString(i)
}

func NewStringBuilder(a *memory.Allocator) *StringBuilder {
	builder := arrowarray.NewBinaryBuilder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	}, arrow.BinaryTypes.String)
	return &StringBuilder{builder: builder}
}

type StringBuilder struct {
	builder *arrowarray.BinaryBuilder
}

func (b *StringBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *StringBuilder) Len() int {
	return b.builder.Len()
}

func (b *StringBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *StringBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *StringBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *StringBuilder) Release() {
	b.builder.Release()
}

func (b *StringBuilder) Append(v string) {
	b.builder.AppendString(v)
}

func (b *StringBuilder) AppendValues(vs []string, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendStringValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.AppendString(v)
	}
}

func (b *StringBuilder) NewStringArray() *String {
	return &String{data: b.builder.NewBinaryArray()}
}
