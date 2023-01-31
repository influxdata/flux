package array

import "github.com/apache/arrow/go/v7/arrow/memory"

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	sv := stringValue{
		data: mem.Allocate(len(v)),
		mem:  mem,
		rc:   1,
	}
	copy(sv.data, v)
	return &String{
		value:  &sv,
		length: n,
	}
}
