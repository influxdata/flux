package staticarray

import "github.com/influxdata/flux/array"

var _ array.Int = Int(nil)

type Int []int64

func (a Int) IsNull(i int) bool {
	return false
}

func (a Int) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a Int) Len() int {
	return len(a)
}

func (a Int) NullN() int {
	return 0
}

func (a Int) Value(i int) int64 {
	return a[i]
}

func (a Int) Slice(start, stop int) array.Base {
	return a.IntSlice(start, stop)
}

func (a Int) IntSlice(start, stop int) array.Int {
	return Int(a[start:stop])
}

func (a Int) Int64Values() []int64 {
	return []int64(a)
}
