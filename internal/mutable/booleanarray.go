package mutable

import (
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/bitutil"
	"github.com/apache/arrow/go/arrow/memory"
)

// BooleanArray is an array of bool values.
type BooleanArray struct {
	arrayBase
	data     *memory.Buffer
	capacity int
}

// NewBooleanArray constructs a new BooleanArray.
func NewBooleanArray(mem memory.Allocator) *BooleanArray {
	return &BooleanArray{
		arrayBase: arrayBase{
			refCount: 1,
			mem:      mem,
		},
	}
}

func (b *BooleanArray) Append(v bool) {
	b.Reserve(1)
	i := b.length
	b.length++
	b.Set(i, v)
}

func (b *BooleanArray) AppendNull() {
	panic("implement me")
}

func (b *BooleanArray) AppendValues(v []bool) {
	b.Reserve(len(v))
	offset := b.length
	b.length += len(v)
	for i := range v {
		b.Set(offset+i, v[i])
	}
}

func (b *BooleanArray) Cap() int { return b.capacity }

func (b *BooleanArray) NewArray() array.Interface {
	return b.NewBooleanArray()
}

func (b *BooleanArray) NewBooleanArray() *array.Boolean {
	data := array.NewData(
		arrow.FixedWidthTypes.Boolean,
		b.length,
		[]*memory.Buffer{nil, b.data},
		nil, 0, 0,
	)
	b.reset()

	a := array.NewBooleanData(data)
	data.Release()
	return a
}

func (b *BooleanArray) init() {
	b.data = memory.NewResizableBuffer(b.mem)
}

func (b *BooleanArray) reset() {
	b.data.Release()
	b.data = nil
	b.length = 0
	b.capacity = 0
}

// Release will release any reference to data buffers.
func (b *BooleanArray) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.data != nil {
			b.reset()
		}
	}
}

func (b *BooleanArray) Reserve(n int) {
	if b.length+n > b.capacity {
		b.resize(b.length + n)
	}
}

func (b *BooleanArray) Resize(n int) {
	if n > b.capacity {
		b.resize(n)
	}
	b.length = n
}

func (b *BooleanArray) resize(n int) {
	if b.data == nil {
		b.init()
	}
	capacity := bitutil.CeilByte(n)
	b.data.Resize(capacity / 8)
	b.capacity = b.data.Cap() * 8
}

func (b *BooleanArray) Value(i int) bool {
	return bitutil.BitIsSet(b.data.Buf(), i)
}

func (b *BooleanArray) Set(i int, v bool) {
	bitutil.SetBitTo(b.data.Buf(), i, v)
}

func (b *BooleanArray) Swap(i, j int) {
	v := b.Value(i)
	b.Set(i, b.Value(j))
	b.Set(j, v)
}
