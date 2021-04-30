package table

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// Values returns the array from the column reader as an array.Interface.
func Values(cr flux.ColReader, j int) array.Interface {
	switch typ := cr.Cols()[j].Type; typ {
	case flux.TInt:
		return cr.Ints(j)
	case flux.TUInt:
		return cr.UInts(j)
	case flux.TFloat:
		return cr.Floats(j)
	case flux.TString:
		return cr.Strings(j)
	case flux.TBool:
		return cr.Bools(j)
	case flux.TTime:
		return cr.Times(j)
	default:
		panic(errors.Newf(codes.Internal, "unimplemented column type: %s", typ))
	}
}
