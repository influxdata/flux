package array

import "github.com/apache/arrow/go/v7/arrow/memory"

func StringRepeat(v string, n int, mem memory.Allocator) *String {
	return &String{
		value:  v,
		length: n,
	}
}
