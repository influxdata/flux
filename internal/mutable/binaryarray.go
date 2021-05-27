package mutable

import (
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/bitutil"
	"github.com/apache/arrow/go/arrow/memory"
)

// BinaryArray implements a mutable array using arrow buffers.
type BinaryArray struct {
	arrayBase
	dtype   arrow.BinaryDataType
	rawData []string
	size    int
}

// NewBinaryArray constructs a new BinaryArray.
func NewBinaryArray(mem memory.Allocator, typ arrow.BinaryDataType) *BinaryArray {
	return &BinaryArray{
		arrayBase: arrayBase{
			refCount: 1,
			mem:      mem,
		},
		dtype: typ,
	}
}

func (b *BinaryArray) AppendString(v string) {
	b.rawData = append(b.rawData, v)
	b.size += len(v)
}

func (b *BinaryArray) AppendStringValues(v []string) {
	b.rawData = append(b.rawData, v...)
	for i := range v {
		b.size += len(b.rawData[i])
	}
}

func (b *BinaryArray) AppendNull() {
	panic("implement me")
}

func (b *BinaryArray) Cap() int { return cap(b.rawData) }
func (b *BinaryArray) Len() int { return len(b.rawData) }

func (b *BinaryArray) NewArray() array.Interface {
	return b.NewBinaryArray()
}

func (b *BinaryArray) NewBinaryArray() *array.Binary {
	builder := array.NewBinaryBuilder(b.mem, b.dtype)
	builder.Reserve(len(b.rawData))
	builder.ReserveData(b.size)
	for _, v := range b.rawData {
		builder.AppendString(v)
	}
	b.reset()
	return builder.NewBinaryArray()
}

func (b *BinaryArray) reset() {
	b.rawData = b.rawData[0:0]
	b.length = 0
	b.size = 0
}

func (b *BinaryArray) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		b.reset()
	}
}

func (b *BinaryArray) Reserve(n int) {
	if len(b.rawData)+n > cap(b.rawData) {
		capacity := bitutil.NextPowerOf2(len(b.rawData) + n)
		newB := make([]string, len(b.rawData), capacity)
		copy(newB, b.rawData)
		b.rawData = newB
	}
}

func (b *BinaryArray) Resize(n int) {
	if n > cap(b.rawData) {
		capacity := bitutil.NextPowerOf2(n)
		newB := make([]string, n, capacity)
		copy(newB, b.rawData)
		b.rawData = newB
	} else {
		b.rawData = b.rawData[:n:cap(b.rawData)]
	}
}

func (b *BinaryArray) ValueString(i int) string {
	return b.rawData[i]
}

func (b *BinaryArray) SetString(i int, v string) {
	old := b.rawData[i]
	b.rawData[i] = v
	b.size += len(v) - len(old)
}

func (b *BinaryArray) Swap(i, j int) {
	b.rawData[i], b.rawData[j] = b.rawData[j], b.rawData[i]
}
