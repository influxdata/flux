package staticarray

import "github.com/influxdata/flux/array"

var _ array.Float = Float(nil)

type Float []float64

func (a Float) IsNull(i int) bool {
	return false
}

func (a Float) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a Float) Len() int {
	return len(a)
}

func (a Float) NullN() int {
	return 0
}

func (a Float) Value(i int) float64 {
	return a[i]
}

func (a Float) Slice(start, stop int) array.Base {
	return a.FloatSlice(start, stop)
}

func (a Float) FloatSlice(start, stop int) array.Float {
	return Float(a[start:stop])
}

func (a Float) Float64Values() []float64 {
	return []float64(a)
}
