package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Bool struct {
	data *arrow.Boolean
}

func (a *Bool) Type() semantic.Type {
	return semantic.Bool
}

func (a *Bool) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Bool) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Bool) Len() int {
	return a.data.Len()
}

func (a *Bool) NullN() int {
	return a.data.NullN()
}

func (a *Bool) Retain() {
	a.data.Retain()
}

func (a *Bool) Release() {
	a.data.Release()
}

func (a *Bool) Value(i int) bool {
	return a.data.Value(i)
}

func NewBoolBuilder(a *memory.Allocator) *BoolBuilder {
	builder := arrow.NewBooleanBuilder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &BoolBuilder{builder: builder}
}

type BoolBuilder struct {
	builder *arrow.BooleanBuilder
}

func (b *BoolBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *BoolBuilder) Len() int {
	return b.builder.Len()
}

func (b *BoolBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *BoolBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *BoolBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *BoolBuilder) Release() {
	b.builder.Release()
}

func (b *BoolBuilder) Append(v bool) {
	b.builder.Append(v)
}

func (b *BoolBuilder) AppendValues(vs []bool, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.Append(v)
	}
}

func (b *BoolBuilder) NewBoolArray() *Bool {
	return &Bool{data: b.builder.NewBooleanArray()}
}
