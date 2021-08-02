package arrow

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/values"
)

// Repeat will construct an arrow array that repeats
// the value n times.
func Repeat(v values.Value, n int, mem memory.Allocator) array.Interface {
	switch v := values.Unwrap(v).(type) {
	case values.Time:
		return array.IntRepeat(int64(v), n, mem)
	default:
		return array.Repeat(v, n, mem)
	}
}
