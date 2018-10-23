package staticarray

import "github.com/influxdata/flux/array"

var _ array.Boolean = Boolean(nil)

type Boolean []bool

func (a Boolean) IsNull(i int) bool {
	return false
}

func (a Boolean) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a Boolean) Len() int {
	return len(a)
}

func (a Boolean) NullN() int {
	return 0
}

func (a Boolean) Value(i int) bool {
	return a[i]
}

func (a Boolean) Slice(start, stop int) array.Base {
	return a.BooleanSlice(start, stop)
}

func (a Boolean) BooleanSlice(start, stop int) array.Boolean {
	return Boolean(a[start:stop])
}
