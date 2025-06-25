package array

import (
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	db := array.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
	db.AppendString(v)
	dict := db.NewArray()
	db.Release()

	ib := array.NewInt32Builder(mem)
	for i := 0; i < n; i++ {
		ib.Append(0)
	}
	indices := ib.NewArray()
	ib.Release()

	data := array.NewDataWithDictionary(
		StringDictionaryType,
		indices.Len(),
		indices.Data().Buffers(),
		indices.Data().NullN(),
		indices.Data().Offset(),
		dict.Data().(*array.Data),
	)
	defer data.Release()
	indices.Release()
	dict.Release()

	return NewStringData(data)
}
