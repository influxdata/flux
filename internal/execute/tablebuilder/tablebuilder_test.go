package tablebuilder_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/execute/tablebuilder"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func newUnlimitedAllocator() *memory.Allocator {
	return &memory.Allocator{}
}

func TestTableBuilder_WithGroupKey(t *testing.T) {
	tb, err := tablebuilder.New(newUnlimitedAllocator()).WithGroupKey(
		execute.NewGroupKey(
			[]flux.ColMeta{
				{
					Label: "_measurement",
					Type:  flux.TString,
				},
				{
					Label: "_field",
					Type:  flux.TString,
				},
			},
			[]values.Value{
				values.NewString("cpu"),
				values.NewString("usage_user"),
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := tb.Floats("_value").Do(func(b array.FloatBuilder) error {
		b.Append(2.0)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	table, err := tb.Table()
	if err != nil {
		t.Fatal(err)
	}

	exp := &executetest.Table{
		KeyCols: []string{"_measurement", "_field"},
		ColMeta: []flux.ColMeta{
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
		Data: [][]interface{}{
			{"cpu", "usage_user", 2.0},
		},
	}
	exp.Normalize()

	got, err := executetest.ConvertTable(table)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(exp, got) {
		t.Fatalf("unexpected table -want/+got:\n%s", cmp.Diff(exp, got))
	}
}
