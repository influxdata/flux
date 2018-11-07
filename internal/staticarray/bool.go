package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type booleans struct {
	data  []bool
	alloc *memory.Allocator
}

func Boolean(data []bool) array.Boolean {
	return &booleans{data: data}
}

func (a *booleans) Type() semantic.Type {
	return semantic.Bool
}

func (a *booleans) IsNull(i int) bool {
	return false
}

func (a *booleans) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *booleans) Len() int {
	return len(a.data)
}

func (a *booleans) NullN() int {
	return 0
}

func (a *booleans) Value(i int) bool {
	return a.data[i]
}

func (a *booleans) Copy() array.Base {
	panic("implement me")
}

func (a *booleans) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * boolSize)
	}
	a.data = nil
}

func (a *booleans) Slice(start, stop int) array.BaseRef {
	return a.BooleanSlice(start, stop)
}

func (a *booleans) BooleanSlice(start, stop int) array.BooleanRef {
	return &booleans{data: a.data[start:stop]}
}

func (a *booleans) BoolValues() []bool {
	return a.data
}

func BooleanBuilder(a *memory.Allocator) array.BooleanBuilder {
	return &booleanBuilder{alloc: a}
}

type booleanBuilder struct {
	data  []bool
	alloc *memory.Allocator
}

func (b *booleanBuilder) Type() semantic.Type {
	return semantic.Bool
}

func (b *booleanBuilder) Len() int {
	return len(b.data)
}

func (b *booleanBuilder) Cap() int {
	return cap(b.data)
}

func (b *booleanBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * boolSize); err != nil {
		panic(err)
	}
	data := make([]bool, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * boolSize)
	b.data = data
}

func (b *booleanBuilder) BuildArray() array.Base {
	return b.BuildBooleanArray()
}

func (b *booleanBuilder) Free() {
	panic("implement me")
}

func (b *booleanBuilder) Append(v bool) {
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

func (b *booleanBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append(false)
}

func (b *booleanBuilder) AppendValues(v []bool, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *booleanBuilder) BuildBooleanArray() array.Boolean {
	return &booleans{
		data:  b.data,
		alloc: b.alloc,
	}
}
