package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestYield_NewQuery(t *testing.T) {
	testcases := []querytest.NewQueryTestCase{
		{
			Name: "mutliple yields",
			Raw: `
				from(bucket: "foo") |> range(start:-1h) |> yield(name: "1")
				from(bucket: "foo") |> range(start:-2h) |> yield(name: "2")
			`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "foo"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "yield2",
						Spec: &universe.YieldOpSpec{
							Name: "1",
						},
					},
					{
						ID: "from3",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "foo"},
						},
					},
					{
						ID: "range4",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "yield5",
						Spec: &universe.YieldOpSpec{
							Name: "2",
						},
					},
				},
				Edges: []operation.Edge{
					{
						Parent: operation.NodeID("from0"),
						Child:  operation.NodeID("range1"),
					},
					{
						Parent: operation.NodeID("range1"),
						Child:  operation.NodeID("yield2"),
					},
					{
						Parent: operation.NodeID("from3"),
						Child:  operation.NodeID("range4"),
					},
					{
						Parent: operation.NodeID("range4"),
						Child:  operation.NodeID("yield5"),
					},
				},
			},
		},
		{
			Name: "yield in sub-block",
			Raw: `
				f = () => {
					g = () => from(bucket: "foo") |> range(start:-1h) |> yield()
					return g
				}
				f()()
			`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "foo"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "yield2",
						Spec: &universe.YieldOpSpec{
							Name: "_result",
						},
					},
				},
				Edges: []operation.Edge{
					{
						Parent: operation.NodeID("from0"),
						Child:  operation.NodeID("range1"),
					},
					{
						Parent: operation.NodeID("range1"),
						Child:  operation.NodeID("yield2"),
					},
				},
			},
		},
		{
			Name: "sub-yield",
			Raw: `
				from(bucket: "foo") |> range(start:-1h) |> yield(name: "1") |> sum() |> yield(name: "2")
			`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "foo"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "yield2",
						Spec: &universe.YieldOpSpec{
							Name: "1",
						},
					},
					{
						ID: "sum3",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.SimpleAggregateConfig{
								Columns: []string{"_value"},
							},
						},
					},
					{
						ID: "yield4",
						Spec: &universe.YieldOpSpec{
							Name: "2",
						},
					},
				},
				Edges: []operation.Edge{
					{
						Parent: operation.NodeID("from0"),
						Child:  operation.NodeID("range1"),
					},
					{
						Parent: operation.NodeID("range1"),
						Child:  operation.NodeID("yield2"),
					},
					{
						Parent: operation.NodeID("yield2"),
						Child:  operation.NodeID("sum3"),
					},
					{
						Parent: operation.NodeID("sum3"),
						Child:  operation.NodeID("yield4"),
					},
				},
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
