package array

import (
	"bytes"
	"sync/atomic"
	"unsafe"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/bitutil"
	"github.com/apache/arrow-go/v18/arrow/memory"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

type StringBuilder struct {
	mem      memory.Allocator
	values   *array.BinaryBuilder
	runSize  int32
	refCount int64
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		mem:      mem,
		values:   array.NewBinaryBuilder(mem, StringType),
		runSize:  0,
		refCount: 1,
	}
}

func (b *StringBuilder) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}
func (b *StringBuilder) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.values != nil {
			b.values.Release()
			b.values = nil
		}
	}
}
func (b *StringBuilder) Len() int {
	if b.runSize > -1 {
		return int(b.runSize)
	}
	return b.values.Len()
}
func (b *StringBuilder) Cap() int {
	return b.values.Cap()
}
func (b *StringBuilder) NullN() int {
	if b.runSize > -1 {
		return 0
	}
	return b.values.NullN()
}

func (b *StringBuilder) hydrate() {
	if b.runSize < 0 {
		panic(errors.New(codes.Internal, "attempting to hydrate already hydrated string builder"))
	}
	values := b.values.NewBinaryArray()
	if b.runSize > 0 {
		b.values.Reserve(int(b.runSize))
		b.values.ReserveData(values.ValueLen(0) * int(b.runSize))
	}
	for i := 0; i < int(b.runSize); i++ {
		b.values.Append(values.Value(0))
	}
	values.Release()
	b.runSize = -1
}

func (b *StringBuilder) AppendBytes(buf []byte) {
	if b.runSize == 0 {
		b.values.Append(buf)
		b.runSize = 1
		return
	} else if b.runSize > 0 {
		if bytes.Equal(buf, b.values.Value(0)) {
			b.runSize += 1
			return
		}
		// Need to add a new value to the values array, that means we
		// need to hydrate it.
		b.hydrate()
	}
	b.values.Append(buf)
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
	if b.runSize > -1 {
		b.hydrate()
	}
	b.values.AppendNull()
}

func (b *StringBuilder) Reserve(n int) {
	b.values.Reserve(n)
}

func (b *StringBuilder) ReserveData(n int) {
	b.values.ReserveData(n)
}

func (b *StringBuilder) Resize(n int) {
	b.values.Resize(n)
}

func (b *StringBuilder) NewArray() Array {
	return b.NewStringArray()
}

func (b *StringBuilder) NewStringArray() *String {
	values := b.values.NewBinaryArray()
	defer values.Release()
	if b.runSize < 0 {
		b.runSize = 0
		return NewStringData(values.Data())
	}

	reb := array.NewInt32Builder(b.mem)
	defer reb.Release()
	reb.Append(b.runSize)
	runEnds := reb.NewInt32Array()
	defer runEnds.Release()

	arr := array.NewRunEndEncodedArray(runEnds, values, int(b.runSize), 0)
	defer arr.Release()
	b.runSize = 0
	return NewStringData(arr.Data())
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
