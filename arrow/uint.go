package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type UInt struct {
	data *arrow.Uint64
}

func (a *UInt) Type() semantic.Type {
	return semantic.UInt
}

func (a *UInt) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *UInt) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *UInt) Len() int {
	return a.data.Len()
}

func (a *UInt) NullN() int {
	return a.data.NullN()
}

func (a *UInt) Retain() {
	a.data.Retain()
}

func (a *UInt) Release() {
	a.data.Release()
}

func (a *UInt) Value(i int) uint64 {
	return a.data.Value(i)
}

func (a *UInt) Uint64Values() []uint64 {
	return a.data.Uint64Values()
}

func NewUIntBuilder(a *memory.Allocator) *UIntBuilder {
	builder := arrow.NewUint64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &UIntBuilder{builder: builder}
}

type UIntBuilder struct {
	builder *arrow.Uint64Builder
}

func (b *UIntBuilder) Type() semantic.Type {
	return semantic.UInt
}

func (b *UIntBuilder) Len() int {
	return b.builder.Len()
}

func (b *UIntBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *UIntBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *UIntBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *UIntBuilder) Release() {
	b.builder.Release()
}

func (b *UIntBuilder) Append(v uint64) {
	b.builder.Append(v)
}

func (b *UIntBuilder) AppendValues(vs []uint64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *UIntBuilder) NewUIntArray() *UInt {
	return &UInt{data: b.builder.NewUint64Array()}
}
