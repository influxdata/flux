package csv_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestFromCSV_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `import "csv" csv.from()`,
			WantErr: true,
		},
		{
			Name:    "from conflicting args",
			Raw:     `import "csv" csv.from(csv:"d", file:"b")`,
			WantErr: true,
		},
		{
			Name:    "from repeat arg",
			Raw:     `import "csv" csv.from(csv:"telegraf", csv:"oops")`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `import "csv" csv.from(csv:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "fromCSV text",
			Raw:  `import "csv" csv.from(csv: "1,2") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "fromCSV0",
						Spec: &csv.FromCSVOpSpec{
							CSV: "1,2",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "fromCSV0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFromCSVOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"fromCSV","kind":"fromCSV","spec":{"csv":"1,2"}}`)
	op := &flux.Operation{
		ID: "fromCSV",
		Spec: &csv.FromCSVOpSpec{
			CSV: "1,2",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
