package array

import (
	"bytes"
	"sync/atomic"

	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

type StringBuilder struct {
	refCount     int64
	builder      *array.BinaryBuilder
	mem          memory.Allocator
	value        *stringValue
	length       int
	capacity     int
	dataCapacity int
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:      mem,
		refCount: 1,
	}
}
func (b *StringBuilder) init() {
	if b.builder != nil {
		return
	}
	builder := array.NewBinaryBuilder(b.mem, StringType)
	if capacity := b.Cap(); capacity > 0 {
		builder.Resize(capacity)
		dataCapacity := b.value.Len() * capacity
		if dataCapacity < b.dataCapacity {
			dataCapacity = b.dataCapacity
		}
		builder.ReserveData(dataCapacity)
	}
	if b.length > 0 {
		for i := 0; i < b.length; i++ {
			builder.Append(b.value.Bytes())
		}
	}
	b.builder = builder
	b.value.Release()
	b.value = nil
}
func (b *StringBuilder) reset() {
	b.builder = nil
	b.length = 0
	b.value = nil
	b.capacity = 0
	b.dataCapacity = 0
}
func (b *StringBuilder) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}
func (b *StringBuilder) Release() {
	if atomic.AddInt64(&b.refCount, -1) != 0 {
		return
	}
	if b.builder != nil {
		b.builder.Release()
		b.builder = nil
	}
	if b.value != nil {
		b.value.Release()
		b.value = nil
	}
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

func (b *StringBuilder) AppendBytes(buf []byte) {
	if b.builder == nil && (b.length == 0 || bytes.Equal(buf, b.value.Bytes())) {
		if b.value == nil {
			b.initValue(len(buf))
			copy(b.value.data, buf)
		}
		b.length++
		return
	}
	if b.value != nil {
		b.init()
	}
	b.builder.Append(buf)
}

// Append appends a string to the array being built. The input string
// will always be copied.
func (b *StringBuilder) Append(v string) {
	if b.builder == nil && (b.length == 0 || v == string(b.value.Bytes())) {
		if b.value == nil {
			b.initValue(len(v))
			copy(b.value.data, v)
		}
		b.length++
		return
	}
	if b.value != nil {
		b.init()
	}
	b.builder.AppendString(v)
}

func (b *StringBuilder) initValue(size int) {
	b.value = &stringValue{
		data: b.mem.Allocate(size),
		mem:  b.mem,
		rc:   1,
	}
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
func (b *StringBuilder) NewArray() Array {
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
func (b *StringBuilder) CopyValidValues(values *String, nullCheckArray Array) {
	if values.Len() != nullCheckArray.Len() {
		panic("Length mismatch between the value array and the null check array")
	}
	b.Reserve(values.Len() - nullCheckArray.NullN())

	nullBitMapBytes := nullCheckArray.NullBitmapBytes()
	nullOffset := nullCheckArray.Data().Offset()
	for i := 0; i < values.Len(); i++ {
		if isValid(nullBitMapBytes, nullOffset, i) {
			b.AppendBytes(values.ValueBytes(i))
		}
	}
}

// Copy of Array.IsValid from arrow, allowing the IsValid check to be done without going through an interface
func isValid(nullBitmapBytes []byte, offset int, i int) bool {
	return len(nullBitmapBytes) == 0 || bitutil.BitIsSet(nullBitmapBytes, offset+i)
}
