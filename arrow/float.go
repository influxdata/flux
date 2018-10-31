package arrow

import (
	arrow "github.com/apache/arrow/go/arrow/array"
	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
)

type Float struct {
	data *arrow.Float64
}

func (a *Float) Type() semantic.Type {
	return semantic.Float
}

func (a *Float) IsNull(i int) bool {
	return a.data.IsNull(i)
}

func (a *Float) IsValid(i int) bool {
	return a.data.IsValid(i)
}

func (a *Float) Len() int {
	return a.data.Len()
}

func (a *Float) NullN() int {
	return a.data.NullN()
}

func (a *Float) Retain() {
	a.data.Retain()
}

func (a *Float) Release() {
	a.data.Release()
}

func (a *Float) Value(i int) float64 {
	return a.data.Value(i)
}

func (a *Float) Float64Values() []float64 {
	return a.data.Float64Values()
}

func NewFloatBuilder(a *memory.Allocator) *FloatBuilder {
	builder := arrow.NewFloat64Builder(&allocator{
		Allocator: arrowmemory.NewGoAllocator(),
		alloc:     a,
	})
	return &FloatBuilder{builder: builder}
}

type FloatBuilder struct {
	builder *arrow.Float64Builder
}

func (b *FloatBuilder) Type() semantic.Type {
	return semantic.Float
}

func (b *FloatBuilder) Len() int {
	return b.builder.Len()
}

func (b *FloatBuilder) Cap() int {
	return b.builder.Cap()
}

func (b *FloatBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *FloatBuilder) AppendNull() {
	b.builder.AppendNull()
}

func (b *FloatBuilder) Release() {
	b.builder.Release()
}

func (b *FloatBuilder) Append(v float64) {
	b.builder.Append(v)
}

func (b *FloatBuilder) AppendValues(vs []float64, valid ...[]bool) {
	if len(valid) > 0 {
		b.builder.AppendValues(vs, valid[0])
		return
	}

	b.builder.Reserve(len(vs))
	for _, v := range vs {
		b.builder.UnsafeAppend(v)
	}
}

func (b *FloatBuilder) NewFloatArray() *Float {
	return &Float{data: b.builder.NewFloat64Array()}
}
