package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type uints struct {
	data  []uint64
	alloc *memory.Allocator
}

func UInt(data []uint64) array.UInt {
	return &uints{data: data}
}

func (a *uints) Type() semantic.Type {
	return semantic.UInt
}

func (a *uints) IsNull(i int) bool {
	return false
}

func (a *uints) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *uints) Len() int {
	return len(a.data)
}

func (a *uints) NullN() int {
	return 0
}

func (a *uints) Value(i int) uint64 {
	return a.data[i]
}

func (a *uints) Copy() array.Base {
	panic("implement me")
}

func (a *uints) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * uint64Size)
	}
	a.data = nil
}

func (a *uints) Slice(start, stop int) array.BaseRef {
	return a.UIntSlice(start, stop)
}

func (a *uints) UIntSlice(start, stop int) array.UIntRef {
	return &uints{data: a.data[start:stop]}
}

func (a *uints) Uint64Values() []uint64 {
	return a.data
}

func UIntBuilder(a *memory.Allocator) array.UIntBuilder {
	return &uintBuilder{alloc: a}
}

type uintBuilder struct {
	data  []uint64
	alloc *memory.Allocator
}

func (b *uintBuilder) Type() semantic.Type {
	return semantic.UInt
}

func (b *uintBuilder) Len() int {
	return len(b.data)
}

func (b *uintBuilder) Cap() int {
	return cap(b.data)
}

func (b *uintBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * uint64Size); err != nil {
		panic(err)
	}
	data := make([]uint64, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * uint64Size)
	b.data = data
}

func (b *uintBuilder) BuildArray() array.Base {
	return b.BuildUIntArray()
}

func (b *uintBuilder) Free() {
	panic("implement me")
}

func (b *uintBuilder) Append(v uint64) {
	if len(b.data) == cap(b.data) {
		// Grow the slice in the same way as built-in append.
		n := len(b.data)
		if n == 0 {
			n = 2
		}
		b.Reserve(n)
	}
	b.data = append(b.data, v)
}

func (b *uintBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append(0)
}

func (b *uintBuilder) AppendValues(v []uint64, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *uintBuilder) BuildUIntArray() array.UInt {
	return &uints{
		data:  b.data,
		alloc: b.alloc,
	}
}
