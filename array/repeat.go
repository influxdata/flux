package array

import "github.com/apache/arrow/go/arrow/memory"

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	return &String{
		value:  v,
		length: n,
	}
}
