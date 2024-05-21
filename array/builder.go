package array

import (
	"bytes"

	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

type StringBuilder struct {
	mem      memory.Allocator
	builder  *array.BinaryBuilder
	constant bool
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:      mem,
		builder:  array.NewBinaryBuilder(mem, StringType),
		constant: true,
	}
}

func (b *StringBuilder) Retain() {
	b.builder.Retain()
}
func (b *StringBuilder) Release() {
	b.builder.Release()
}
func (b *StringBuilder) Len() int {
	return b.builder.Len()
}
func (b *StringBuilder) Cap() int {
	return b.builder.Cap()
}
func (b *StringBuilder) NullN() int {
	return b.builder.NullN()
}

func (b *StringBuilder) AppendBytes(buf []byte) {
	if b.builder.Len() > 0 {
		b.constant = b.constant && bytes.Equal(buf, b.builder.Value(0))
	}
	b.builder.Append(buf)
}

// Append appends a string to the array being built. The input string
// will always be copied.
func (b *StringBuilder) Append(v string) {
	b.AppendBytes([]byte(v))
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
	b.constant = false
	b.builder.AppendNull()
}

func (b *StringBuilder) UnsafeAppendBoolToBitmap(isValid bool) {
	b.builder.UnsafeAppendBoolToBitmap(isValid)
}

func (b *StringBuilder) Reserve(n int) {
	b.builder.Reserve(n)
}

func (b *StringBuilder) ReserveData(n int) {
	b.builder.ReserveData(n)
}

func (b *StringBuilder) Resize(n int) {
	b.builder.Resize(n)
}

func (b *StringBuilder) NewArray() Array {
	return b.NewStringArray()
}

func (b *StringBuilder) NewStringArray() *String {
	arr := b.builder.NewBinaryArray()
	if !b.constant || arr.Len() < 1 {
		b.constant = true
		return &String{arr}
	}
	defer arr.Release()
	return StringRepeat(arr.ValueString(0), arr.Len(), b.mem)
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
			b.Append(values.Value(i))
		}
	}
}

// Copy of Array.IsValid from arrow, allowing the IsValid check to be done without going through an interface
func isValid(nullBitmapBytes []byte, offset int, i int) bool {
	return len(nullBitmapBytes) == 0 || bitutil.BitIsSet(nullBitmapBytes, offset+i)
}
