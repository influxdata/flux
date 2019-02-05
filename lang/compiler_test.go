package lang_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestASTCompiler(t *testing.T) {
	script := `
import "csv"
option now = () => 2017-10-10T00:01:00Z
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`
	astPkg, err := flux.Parse(script)
	if err != nil {
		t.Fatalf("failed to parse script: %v", err)
	}

	c := lang.ASTCompiler{
		AST: astPkg,
	}
	gotSpec, err := c.Compile(context.Background())
	if err != nil {
		t.Fatalf("failed to compile AST: %v", err)
	}

	wantSpec := &flux.Spec{
		Operations: []*flux.Operation{
			{
				ID:   flux.OperationID("fromCSV0"),
				Spec: &csv.FromCSVOpSpec{CSV: "foo,bar"},
			},
			{
				ID: flux.OperationID("range1"),
				Spec: &universe.RangeOpSpec{
					Start:      flux.Time{Absolute: time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC)},
					Stop:       flux.Time{IsRelative: true},
					TimeColumn: "_time", StartColumn: "_start", StopColumn: "_stop"},
			},
		},
		Edges: []flux.Edge{{Parent: flux.OperationID("fromCSV0"), Child: flux.OperationID("range1")}},
		Now:   time.Date(2017, 10, 10, 0, 1, 0, 0, time.UTC),
	}

	cmpOpts := cmpopts.IgnoreUnexported(flux.Spec{})
	if !cmp.Equal(wantSpec, gotSpec, cmpOpts) {
		t.Fatalf("compiler produced unexpected spec; -want/+got:\n%v\n", cmp.Diff(wantSpec, gotSpec, cmpOpts))
	}
}
