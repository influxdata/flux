package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type UInt64 struct {
	data *arrow.Uint64
}

func (a *UInt64) Type() semantic.Type {
	return semantic.UInt
}

func (a *UInt64) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *UInt64) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *UInt64) Len() int {
	return a.data.Len()
}

func (a *UInt64) NullN() int {
	return a.data.NullN()
}

func (a *UInt64) Slice(start, stop int) array.BaseRef {
	return a.UIntSlice(start, stop)
}

func (a *UInt64) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *UInt64) Value(i int) uint64 {
	return a.data.Value(i)
}

func (a *UInt64) UIntSlice(start, stop int) array.UIntRef {
	panic("implement me")
}

func (a *UInt64) Uint64Values() []uint64 {
	return a.data.Uint64Values()
}

func (a *UInt64) Free() {
	a.data.Release()
}

func UIntBuilder(a *memory.Allocator) array.UIntBuilder {
	builder := arrow.NewUint64Builder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &uintBuilder{builder: builder}
}

type uintBuilder struct {
	builder *arrow.Uint64Builder
}

func (b *uintBuilder) Type() semantic.Type {
	return semantic.UInt
}

func (b *uintBuilder) Len() int {
	return b.builder.Len()
}

func (b *uintBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *uintBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *uintBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *uintBuilder) BuildArray() array.Base {
	return b.BuildUIntArray()
}

func (b *uintBuilder) Free() {
	b.builder.Release()
}

func (b *uintBuilder) Append(v uint64) {
	b.builder.Append(v)
}

func (b *uintBuilder) AppendValues(vs []uint64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *uintBuilder) BuildUIntArray() array.UInt {
	return &UInt64{data: b.builder.NewUint64Array()}
}
