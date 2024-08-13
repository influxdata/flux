package array

import (
	"bytes"
	"sync/atomic"
	"unsafe"

	"github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

type StringBuilder struct {
	mem         memory.Allocator
	len         int
	cap         int
	reserveData int
	buffer      *memory.Buffer
	builder     *array.BinaryBuilder
	refCount    int64
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:         mem,
		len:         0,
		cap:         0,
		reserveData: 0,
		buffer:      nil,
		builder:     nil,
		refCount:    1,
	}
}

func (b *StringBuilder) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}
func (b *StringBuilder) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.buffer != nil {
			b.buffer.Release()
		}
		if b.builder != nil {
			b.builder.Release()
		}
	}
}
func (b *StringBuilder) Len() int {
	if b.builder != nil {
		return b.builder.Len()
	}
	return b.len
}
func (b *StringBuilder) Cap() int {
	if b.builder != nil {
		return b.builder.Cap()
	}
	if b.cap > b.len {
		return b.cap
	}
	return b.len
}
func (b *StringBuilder) NullN() int {
	if b.builder != nil {
		return b.builder.NullN()
	}
	return 0
}

func (b *StringBuilder) AppendBytes(buf []byte) {
	if b.builder != nil {
		b.builder.Append(buf)
		return
	}
	if b.len == 0 {
		b.buffer = memory.NewResizableBuffer(b.mem)
		b.buffer.Resize(len(buf))
		copy(b.buffer.Bytes(), buf)
		b.len = 1
		return
	}
	if bytes.Equal(b.buffer.Bytes(), buf) {
		b.len++
		return
	}
	b.makeBuilder(buf)

}

// Append appends a string to the array being built. A reference
// to the input string will not be retained by the builder. The
// string will be copied, if necessary.
func (b *StringBuilder) Append(v string) {
	// Avoid copying the input string as AppendBytes
	// will never keep a reference or modify the input.
	bytes := unsafe.Slice(unsafe.StringData(v), len(v))
	b.AppendBytes(bytes)
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
	if b.builder == nil {
		b.makeBuilder(nil)
	}
	b.builder.AppendNull()
}

func (b *StringBuilder) Reserve(n int) {
	if b.builder != nil {
		b.builder.Reserve(n)
		return
	}
	if b.len+n > b.cap {
		b.cap = b.len + n
	}
}

func (b *StringBuilder) ReserveData(n int) {
	if b.builder != nil {
		b.builder.ReserveData(n)
		return
	}
	b.reserveData = n
}

func (b *StringBuilder) Resize(n int) {
	if b.builder != nil {
		b.builder.Resize(n)
	}
	b.cap = n
	if b.len > n {
		b.len = n
	}
}

func (b *StringBuilder) NewArray() Array {
	return b.NewStringArray()
}

func (b *StringBuilder) NewStringArray() *String {
	if b.builder != nil {
		arr := &String{b.builder.NewBinaryArray()}
		b.builder.Release()
		b.builder = nil
		return arr
	}
	if b.buffer != nil {
		arr := &String{&repeatedBinary{
			len: b.len,
			buf: b.buffer,
		}}
		b.buffer = nil
		b.len = 0
		b.cap = 0
		return arr
	}
	// getting this far means we have an empty array.
	arr := StringRepeat("", b.len, b.mem)
	b.len = 0
	b.cap = 0
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
			b.Append(values.Value(i))
		}
	}
}

func (b *StringBuilder) makeBuilder(value []byte) {
	bufferLen := 0
	if b.buffer != nil {
		bufferLen = b.buffer.Len()
	}
	size := b.len
	if b.cap > b.len {
		size = b.cap
	}
	dataSize := b.len * bufferLen
	if value != nil {
		if b.cap <= b.len {
			size++
		}
		dataSize += len(value)
	}
	if b.reserveData > dataSize {
		dataSize = b.reserveData
	}
	b.builder = array.NewBinaryBuilder(b.mem, arrow.BinaryTypes.String)
	b.builder.Resize(size)
	b.builder.ReserveData(dataSize)
	for i := 0; i < b.len; i++ {
		b.builder.Append(b.buffer.Bytes())
	}
	if value != nil {
		b.builder.Append(value)
	}
	if b.buffer != nil {
		b.buffer.Release()
		b.buffer = nil
	}
	b.len = 0
	b.cap = 0
	b.reserveData = 0
}

// Copy of Array.IsValid from arrow, allowing the IsValid check to be done without going through an interface
func isValid(nullBitmapBytes []byte, offset int, i int) bool {
	return len(nullBitmapBytes) == 0 || bitutil.BitIsSet(nullBitmapBytes, offset+i)
}
