package staticarray

import "github.com/influxdata/flux/array"

var _ array.String = String(nil)

type String []string

func (a String) IsNull(i int) bool {
	return false
}

func (a String) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a String) Len() int {
	return len(a)
}

func (a String) NullN() int {
	return 0
}

func (a String) Value(i int) string {
	return a[i]
}

func (a String) Slice(start, stop int) array.Base {
	return a.StringSlice(start, stop)
}

func (a String) StringSlice(start, stop int) array.String {
	return String(a[start:stop])
}
