package arrow

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/values"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

// Repeat will construct an arrow array that repeats
// the value n times.
func Repeat(colType flux.ColType, v values.Value, n int, mem memory.Allocator) array.Array {
	switch colType {
	case flux.TInt:
		var ival int64
		if !v.IsNull() {
			ival = v.Int()
		}
		return array.IntRepeat(ival, v.IsNull(), n, mem)
	case flux.TUInt:
		var uival uint64
		if !v.IsNull() {
			uival = v.UInt()
		}
		return array.UintRepeat(uival, v.IsNull(), n, mem)
	case flux.TFloat:
		var fval float64
		if !v.IsNull() {
			fval = v.Float()
		}
		return array.FloatRepeat(fval, v.IsNull(), n, mem)
	case flux.TBool:
		var bval bool
		if !v.IsNull() {
			bval = v.Bool()
		}
		return array.BooleanRepeat(bval, v.IsNull(), n, mem)
	case flux.TString:
		var sval string
		if !v.IsNull() {
			sval = v.Str()
		}
		return array.StringRepeat(sval, n, mem)
	case flux.TTime:
		var tval values.Time
		if !v.IsNull() {
			tval = v.Time()
		}
		return array.IntRepeat(int64(tval), v.IsNull(), n, mem)
	default:
		panic(errors.Newf(codes.Internal, "invalid arrow primitive type: %T", colType))
	}
}
