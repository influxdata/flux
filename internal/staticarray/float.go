package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type floats struct {
	data  []float64
	alloc *memory.Allocator
}

func Float(data []float64) array.Float {
	return &floats{data: data}
}

func (a *floats) Type() semantic.Type {
	return semantic.Float
}

func (a *floats) IsNull(i int) bool {
	return false
}

func (a *floats) IsValid(i int) bool {
	return i >= 0 && i < len(a.data)
}

func (a *floats) Len() int {
	return len(a.data)
}

func (a *floats) NullN() int {
	return 0
}

func (a *floats) Value(i int) float64 {
	return a.data[i]
}

func (a *floats) Copy() array.Base {
	panic("implement me")
}

func (a *floats) Free() {
	if a.alloc != nil {
		a.alloc.Free(cap(a.data) * float64Size)
	}
	a.data = nil
}

func (a *floats) Slice(start, stop int) array.BaseRef {
	return a.FloatSlice(start, stop)
}

func (a *floats) FloatSlice(start, stop int) array.FloatRef {
	return Float(a.data[start:stop])
}

func (a *floats) Float64Values() []float64 {
	return a.data
}

func FloatBuilder(a *memory.Allocator) array.FloatBuilder {
	return &floatBuilder{alloc: a}
}

type floatBuilder struct {
	data  []float64
	alloc *memory.Allocator
}

func (b *floatBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *floatBuilder) Len() int {
	if b == nil {
		return 0
	}
	return len(b.data)
}

func (b *floatBuilder) Cap() int {
	return cap(b.data)
}

func (b *floatBuilder) Reserve(n int) {
	newCap := len(b.data) + n
	if newCap := len(b.data) + n; newCap <= cap(b.data) {
		return
	}
	if err := b.alloc.Allocate(newCap * float64Size); err != nil {
		panic(err)
	}
	data := make([]float64, len(b.data), newCap)
	copy(data, b.data)
	b.alloc.Free(cap(b.data) * float64Size)
	b.data = data
}

func (b *floatBuilder) BuildArray() array.Base {
	return b.BuildFloatArray()
}

func (b *floatBuilder) Free() {
	panic("implement me")
}

func (b *floatBuilder) Append(v float64) {
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

func (b *floatBuilder) AppendNull() {
	// The staticarray does not support nulls so it will do the current behavior of just appending
	// the zero value.
	b.Append(0)
}

func (b *floatBuilder) AppendValues(v []float64, valid ...[]bool) {
	if newCap := len(b.data) + len(v); newCap > cap(b.data) {
		b.Reserve(newCap - cap(b.data))
	}
	b.data = append(b.data, v...)
}

func (b *floatBuilder) BuildFloatArray() array.Float {
	return &floats{
		data:  b.data,
		alloc: b.alloc,
	}
}
