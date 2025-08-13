package array

import (
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	db := array.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
	db.AppendString(v)
	values := db.NewArray()
	db.Release()

	reb := array.NewInt32Builder(mem)
	reb.Append(int32(n))
	runEnds := reb.NewArray()
	reb.Release()

	arr := array.NewRunEndEncodedArray(runEnds, values, n, 0)
	defer arr.Release()
	runEnds.Release()
	values.Release()

	return NewStringData(arr.Data())
}
