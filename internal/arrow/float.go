package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Float64 struct {
	data *arrow.Float64
}

func (a *Float64) Type() semantic.Type {
	return semantic.Float
}

func (a *Float64) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Float64) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Float64) Len() int {
	return a.data.Len()
}

func (a *Float64) NullN() int {
	return a.data.NullN()
}

func (a *Float64) Slice(start, stop int) array.BaseRef {
	return a.FloatSlice(start, stop)
}

func (a *Float64) Copy() array.Base {
	a.data.Retain()
	return a
}

func (a *Float64) Value(i int) float64 {
	return a.data.Value(i)
}

func (a *Float64) FloatSlice(start, stop int) array.FloatRef {
	panic("implement me")
}

func (a *Float64) Float64Values() []float64 {
	return a.data.Float64Values()
}

func (a *Float64) Free() {
	a.data.Release()
}

func FloatBuilder(a *memory.Allocator) array.FloatBuilder {
	builder := arrow.NewFloat64Builder(&Allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &floatBuilder{builder: builder}
}

type floatBuilder struct {
	builder *arrow.Float64Builder
}

func (b *floatBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *floatBuilder) Len() int {
	return b.builder.Len()
}

func (b *floatBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *floatBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *floatBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *floatBuilder) BuildArray() array.Base {
	return b.BuildFloatArray()
}

func (b *floatBuilder) Free() {
	b.builder.Release()
}

func (b *floatBuilder) Append(v float64) {
	b.builder.Append(v)
}

func (b *floatBuilder) AppendValues(vs []float64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *floatBuilder) BuildFloatArray() array.Float {
	return &Float64{data: b.builder.NewFloat64Array()}
}
