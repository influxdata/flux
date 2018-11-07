package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type strings struct {
	data  []string
	alloc *memory.Allocator
}

func String(data []string) array.String {
	return &strings{data: data}
}

func (a *strings) Type() semantic.Type {
	return semantic.String
}

func (a *strings) IsNull(i int) bool {
	return false
}

func (a *strings) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *strings) Len() int {
	return len(a.data)
}

func (a *strings) NullN() int {
	return 0
}

func (a *strings) Value(i int) string {
	return a.data[i]
}

func (a *strings) Copy() array.Base {
	panic("implement me")
}

func (a *strings) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * stringSize)
	}
	a.data = nil
}

func (a *strings) Slice(start, stop int) array.BaseRef {
	return a.StringSlice(start, stop)
}

func (a *strings) StringSlice(start, stop int) array.StringRef {
	return &strings{data: a.data[start:stop]}
}

func (a *strings) StringValues() []string {
	return a.data
}

func StringBuilder(a *memory.Allocator) array.StringBuilder {
	return &stringBuilder{alloc: a}
}

type stringBuilder struct {
	data  []string
	alloc *memory.Allocator
}

func (b *stringBuilder) Type() semantic.Type {
	return semantic.String
}

func (b *stringBuilder) Len() int {
	return len(b.data)
}

func (b *stringBuilder) Cap() int {
	return cap(b.data)
}

func (b *stringBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * stringSize); err != nil {
		panic(err)
	}
	data := make([]string, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * stringSize)
	b.data = data
}

func (b *stringBuilder) BuildArray() array.Base {
	return b.BuildStringArray()
}

func (b *stringBuilder) Free() {
	panic("implement me")
}

func (b *stringBuilder) Append(v string) {
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

func (b *stringBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append("")
}

func (b *stringBuilder) AppendValues(v []string, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *stringBuilder) BuildStringArray() array.String {
	return &strings{
		data:  b.data,
		alloc: b.alloc,
	}
}
