package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type times struct {
	data  []values.Time
	alloc *memory.Allocator
}

func Time(data []values.Time) array.Time {
	return &times{data: data}
}

func (a *times) Type() semantic.Type {
	return semantic.Time
}

func (a *times) IsNull(i int) bool {
	return false
}

func (a *times) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *times) Len() int {
	return len(a.data)
}

func (a *times) NullN() int {
	return 0
}

func (a *times) Value(i int) values.Time {
	return a.data[i]
}

func (a *times) Copy() array.Base {
	panic("implement me")
}

func (a *times) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * timeSize)
	}
	a.data = nil
}

func (a *times) Slice(start, stop int) array.BaseRef {
	return a.TimeSlice(start, stop)
}

func (a *times) TimeSlice(start, stop int) array.TimeRef {
	return &times{data: a.data[start:stop]}
}

func (a *times) TimeValues() []values.Time {
	return a.data
}

func TimeBuilder(a *memory.Allocator) array.TimeBuilder {
	return &timeBuilder{alloc: a}
}

type timeBuilder struct {
	data  []values.Time
	alloc *memory.Allocator
}

func (b *timeBuilder) Type() semantic.Type {
	return semantic.Time
}

func (b *timeBuilder) Len() int {
	return len(b.data)
}

func (b *timeBuilder) Cap() int {
	return cap(b.data)
}

func (b *timeBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * timeSize); err != nil {
		panic(err)
	}
	data := make([]values.Time, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * timeSize)
	b.data = data
}

func (b *timeBuilder) BuildArray() array.Base {
	return b.BuildTimeArray()
}

func (b *timeBuilder) Free() {
	panic("implement me")
}

func (b *timeBuilder) Append(v values.Time) {
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

func (b *timeBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append(0)
}

func (b *timeBuilder) AppendValues(v []values.Time, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *timeBuilder) BuildTimeArray() array.Time {
	return &times{
		data:  b.data,
		alloc: b.alloc,
	}
}
