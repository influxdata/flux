package array

import (
	"bytes"
	"sync/atomic"
	"unsafe"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/bitutil"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

type StringBuilder struct {
	values   *array.BinaryBuilder
	indices  *array.Int32Builder
	refCount int64

	hydratedArray bool
}

func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	return &StringBuilder{
		values:        array.NewBinaryBuilder(mem, StringType),
		indices:       array.NewInt32Builder(mem),
		refCount:      1,
		hydratedArray: false,
	}
}

func (b *StringBuilder) Retain() {
	atomic.AddInt64(&b.refCount, 1)
}
func (b *StringBuilder) Release() {
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.indices != nil {
			b.indices.Release()
			b.indices = nil
		}
		if b.values != nil {
			b.values.Release()
			b.values = nil
		}
	}
}
func (b *StringBuilder) Len() int {
	if b.hydratedArray {
		return b.values.Len()
	}
	return b.indices.Len()
}
func (b *StringBuilder) Cap() int {
	if b.hydratedArray {
		return b.values.Cap()
	}
	return b.indices.Cap()
}
func (b *StringBuilder) NullN() int {
	if b.hydratedArray {
		return b.values.NullN()
	}
	return b.indices.NullN()
}

func (b *StringBuilder) AppendBytes(buf []byte) {
	if !b.hydratedArray {
		if b.values.Len() == 0 {
			b.values.Append(buf)
			b.indices.Append(0)
			return
		}
		if bytes.Equal(buf, b.values.Value(0)) {
			b.indices.Append(0)
			return
		}
		// Need to add a new value to the values array, that means we
		// need to hydrate it.
		b.hydratedArray = true
		indices := b.indices.NewInt32Array()
		values := b.values.NewBinaryArray()
		b.values.Reserve(indices.Len())
		b.values.ReserveData(values.ValueLen(0) * indices.Len())
		for i := 0; i < indices.Len(); i++ {
			if indices.IsNull(i) {
				b.values.AppendNull()
			} else {
				b.values.Append(values.Value(int(indices.Value(i))))
			}
		}
		values.Release()
		indices.Release()
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
	if b.hydratedArray {
		b.values.AppendNull()
		return
	}
	b.indices.AppendNull()
}

func (b *StringBuilder) Reserve(n int) {
	if b.hydratedArray {
		b.values.Reserve(n)
		return
	}
	b.indices.Reserve(n)
}

func (b *StringBuilder) ReserveData(n int) {
	b.values.ReserveData(n)
}

func (b *StringBuilder) Resize(n int) {
	if b.hydratedArray {
		b.values.Resize(n)
		return
	}
	b.indices.Resize(n)
}

func (b *StringBuilder) NewArray() Array {
	return b.NewStringArray()
}

func (b *StringBuilder) NewStringArray() *String {
	values := b.values.NewBinaryArray()
	defer values.Release()
	if b.hydratedArray {
		b.hydratedArray = false
		return NewStringData(values.Data())
	}

	indices := b.indices.NewInt32Array()
	defer indices.Release()
	data := array.NewDataWithDictionary(
		StringDictionaryType,
		indices.Len(),
		indices.Data().Buffers(),
		indices.NullN(),
		0,
		values.Data().(*array.Data),
	)
	defer data.Release()
	return NewStringData(data)
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
