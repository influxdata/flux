package table

import (
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/executetest"
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

// Diff will perform a diff between two tables.
// If the tables are the same, the output will be an empty string.
// This will produce a fatal error if there was any problem reading
// either table.
func Diff(tb testing.TB, want, got flux.Table) string {
	tb.Helper()

	wantT, err := executetest.ConvertTable(want)
	if err != nil {
		tb.Fatalf("unexpected error reading want table: %s", err)
	}
	gotT, err := executetest.ConvertTable(got)
	if err != nil {
		tb.Fatalf("unexpected error reading got table: %s", err)
	}

	wantT.Normalize()
	gotT.Normalize()
	return cmp.Diff(wantT, gotT)
}
