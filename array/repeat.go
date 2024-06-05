package array

import (
	"unsafe"

	"github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	buf := memory.NewResizableBuffer(mem)
	buf.Resize(len(v))
	copy(buf.Bytes(), v)
	return &String{
		binaryArray: &repeatedBinary{
			len: n,
			buf: buf,
		},
	}
}

type repeatedBinary struct {
	len int
	buf *memory.Buffer
}

func (*repeatedBinary) Data() arrow.ArrayData {
	return nil
}

func (*repeatedBinary) NullN() int {
	return 0
}

func (*repeatedBinary) NullBitmapBytes() []byte {
	return nil
}

func (*repeatedBinary) IsNull(i int) bool {
	return false
}

func (b *repeatedBinary) IsValid(i int) bool {
	return i < b.len
}

func (b *repeatedBinary) Len() int {
	return b.len
}

func (b *repeatedBinary) ValueBytes() []byte {
	return b.buf.Bytes()
}

func (b *repeatedBinary) ValueLen(int) int {
	return b.buf.Len()
}

func (*repeatedBinary) ValueOffset(int) int {
	return 0
}

func (b *repeatedBinary) ValueString(int) string {
	return unsafe.String(unsafe.SliceData(b.buf.Bytes()), b.buf.Len())
}

func (b *repeatedBinary) Retain() {
	b.buf.Retain()
}

func (b *repeatedBinary) Release() {
	b.buf.Release()
}

func (b *repeatedBinary) IsConstant() bool {
	return true
}

func (b *repeatedBinary) Slice(i, j int) binaryArray {
	b.buf.Retain()
	return &repeatedBinary{
		len: j - i,
		buf: b.buf,
	}
}
