package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Int struct {
	data *arrow.Int64
}

func (a *Int) Type() semantic.Type {
	return semantic.Int
}

func (a *Int) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Int) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Int) Len() int {
	return a.data.Len()
}

func (a *Int) NullN() int {
	return a.data.NullN()
}

func (a *Int) Retain() {
	a.data.Retain()
}

func (a *Int) Release() {
	a.data.Release()
}

func (a *Int) Value(i int) int64 {
	return a.data.Value(i)
}

func (a *Int) Int64Values() []int64 {
	return a.data.Int64Values()
}

func NewIntBuilder(a *memory.Allocator) *IntBuilder {
	builder := arrow.NewInt64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &IntBuilder{builder: builder}
}

type IntBuilder struct {
	builder *arrow.Int64Builder
}

func (b *IntBuilder) Type() semantic.Type {
	return semantic.Int
}

func (b *IntBuilder) Len() int {
	return b.builder.Len()
}

func (b *IntBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *IntBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *IntBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *IntBuilder) Release() {
	b.builder.Release()
}

func (b *IntBuilder) Append(v int64) {
	b.builder.Append(v)
}

func (b *IntBuilder) AppendValues(vs []int64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *IntBuilder) NewIntArray() *Int {
	return &Int{data: b.builder.NewInt64Array()}
}
