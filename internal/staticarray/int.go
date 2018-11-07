package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type ints struct {
	data  []int64
	alloc *memory.Allocator
}

func Int(data []int64) array.Int {
	return &ints{data: data}
}

func (a *ints) Type() semantic.Type {
	return semantic.Int
}

func (a *ints) IsNull(i int) bool {
	return false
}

func (a *ints) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *ints) Len() int {
	return len(a.data)
}

func (a *ints) NullN() int {
	return 0
}

func (a *ints) Value(i int) int64 {
	return a.data[i]
}

func (a *ints) Copy() array.Base {
	panic("implement me")
}

func (a *ints) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * int64Size)
	}
	a.data = nil
}

func (a *ints) Slice(start, stop int) array.BaseRef {
	return a.IntSlice(start, stop)
}

func (a *ints) IntSlice(start, stop int) array.IntRef {
	return &ints{data: a.data[start:stop]}
}

func (a *ints) Int64Values() []int64 {
	return a.data
}

func IntBuilder(a *memory.Allocator) array.IntBuilder {
	return &intBuilder{alloc: a}
}

type intBuilder struct {
	data  []int64
	alloc *memory.Allocator
}

func (b *intBuilder) Type() semantic.Type {
	return semantic.Int
}

func (b *intBuilder) Len() int {
	return len(b.data)
}

func (b *intBuilder) Cap() int {
	return cap(b.data)
}

func (b *intBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * int64Size); err != nil {
		panic(err)
	}
	data := make([]int64, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * int64Size)
	b.data = data
}

func (b *intBuilder) BuildArray() array.Base {
	return b.BuildIntArray()
}

func (b *intBuilder) Free() {
	panic("implement me")
}

func (b *intBuilder) Append(v int64) {
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

func (b *intBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append(0)
}

func (b *intBuilder) AppendValues(v []int64, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *intBuilder) BuildIntArray() array.Int {
	return &ints{
		data:  b.data,
		alloc: b.alloc,
	}
}
