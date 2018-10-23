package staticarray

import "github.com/influxdata/flux/array"

var _ array.UInt = UInt(nil)

type UInt []uint64

func (a UInt) IsNull(i int) bool {
	return false
}

func (a UInt) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a UInt) Len() int {
	return len(a)
}

func (a UInt) NullN() int {
	return 0
}

func (a UInt) Value(i int) uint64 {
	return a[i]
}

func (a UInt) Slice(start, stop int) array.Base {
	return a.UIntSlice(start, stop)
}

func (a UInt) UIntSlice(start, stop int) array.UInt {
	return UInt(a[start:stop])
}

func (a UInt) Uint64Values() []uint64 {
	return []uint64(a)
}
