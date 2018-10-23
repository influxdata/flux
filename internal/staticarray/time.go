package staticarray

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/values"
)

var _ array.Time = Time(nil)

type Time []values.Time

func (a Time) IsNull(i int) bool {
	return false
}

func (a Time) IsValid(i int) bool {
	return i >= 0 && i < len(a)
}

func (a Time) Len() int {
	return len(a)
}

func (a Time) NullN() int {
	return 0
}

func (a Time) Value(i int) values.Time {
	return a[i]
}

func (a Time) Slice(start, stop int) array.Base {
	return a.TimeSlice(start, stop)
}

func (a Time) TimeSlice(start, stop int) array.Time {
	return Time(a[start:stop])
}

func (a Time) TimeValues() []values.Time {
	return []values.Time(a)
}
