package array

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

type StringBuilder struct {
	builder      *array.BinaryBuilder
	mem          memory.Allocator
	value        string
	length       int
	capacity     int
	dataCapacity int
	refCount     int
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:      mem,
		refCount: 1,
	}
}
func (b *StringBuilder) init() {
	if b.builder == nil {
		if b.refCount <= 0 {
			return
		}

		builder := array.NewBinaryBuilder(b.mem, StringType)
		if capacity := b.Cap(); capacity > 0 {
			builder.Resize(capacity)
			dataCapacity := len(b.value) * capacity
			if dataCapacity < b.dataCapacity {
				dataCapacity = b.dataCapacity
			}
			builder.ReserveData(dataCapacity)
		}
		if b.length > 0 {
			for i := 0; i < b.length; i++ {
				builder.AppendString(b.value)
			}
		}

		for i := 1; i < b.refCount; i++ {
			builder.Retain()
		}
		b.builder = builder
	}
}
func (b *StringBuilder) reset() {
	b.builder = nil
	b.length = 0
	b.capacity = 0
	b.dataCapacity = 0
	b.value = ""
}
func (b *StringBuilder) Retain() {
	if b.builder != nil {
		b.builder.Retain()
		return
	}
	b.refCount++
}
func (b *StringBuilder) Release() {
	if b.builder != nil {
		b.builder.Release()
		return
	}
	b.refCount--
}
func (b *StringBuilder) Len() int {
	if b.builder != nil {
		return b.builder.Len()
	}
	return b.length
}
func (b *StringBuilder) Cap() int {
	if b.builder != nil {
		return b.builder.Cap()
	}

	capacity := b.capacity
	if capacity < b.length {
		capacity = b.length
	}
	return capacity
}
func (b *StringBuilder) NullN() int {
	if b.builder != nil {
		return b.builder.NullN()
	}
	return 0
}
func (b *StringBuilder) Append(v string) {
	if b.builder == nil && (b.length == 0 || v == b.value) {
		b.value = v
		b.length++
		return
	}
	b.init()
	b.builder.AppendString(v)
}
func (b *StringBuilder) AppendValues(v []string, valid []bool) {
	for i, val := range v {
		if len(valid) != 0 && valid[i] {
			b.AppendNull()
			continue
		}
		b.Append(val)
	}
}
func (b *StringBuilder) AppendNull() {
	b.init()
	b.builder.AppendNull()
}
func (b *StringBuilder) UnsafeAppendBoolToBitmap(isValid bool) {
	b.init()
	b.builder.UnsafeAppendBoolToBitmap(isValid)
}
func (b *StringBuilder) Reserve(n int) {
	if b.builder != nil {
		b.builder.Reserve(n)
		return
	}
	b.capacity = n
}
func (b *StringBuilder) ReserveData(n int) {
	if b.builder != nil {
		b.builder.ReserveData(n)
		return
	}
	b.dataCapacity = n
}
func (b *StringBuilder) Resize(n int) {
	if b.builder != nil {
		b.builder.Resize(n)
		return
	}
	// In arrow, resize and reserve both affect
	// the capacity. Neither of them change the
	// length of the built array.
	b.capacity = n
}
func (b *StringBuilder) NewArray() Interface {
	return b.NewStringArray()
}
func (b *StringBuilder) NewStringArray() *String {
	arr := &String{}
	if b.builder == nil {
		arr.value, arr.length = b.value, b.length
	} else {
		arr.data = b.builder.NewBinaryArray()
	}
	b.reset()
	return arr
}
