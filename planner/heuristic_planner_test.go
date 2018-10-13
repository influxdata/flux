package planner_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
)

func TestPlanTraversal(t *testing.T) {

	testCases := []struct {
		name      string
		fluxQuery string
		nodeIDs   []planner.NodeID
	}{
		{
			name:      "simple",
			fluxQuery: `from(bucket: "foo")`,
			nodeIDs:   []planner.NodeID{"from0"},
		},
		{
			name:      "from and filter",
			fluxQuery: `from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")`,
			nodeIDs:   []planner.NodeID{"filter1", "from0"},
		},
		//{
		//	name: "multi-root",
		//	fluxQuery: `
		//		from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu") |> yield(name: "1")
		//		from(bucket: "foo") |> filter(fn: (r) => r._field == "fan") |> yield(name: "2")`,
		//	nodeIDs: []planner.NodeID{"filter1", "from0", "filter3", "from2"},
		//},
		{
			name: "join",
			fluxQuery: `
			    left = from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")
                right = from(bucket: "foo") |> range(start: -1d)
                join(tables: {l: left, r: right}, on: ["key"]) |> yield()`,
			nodeIDs: []planner.NodeID{"yield5", "join4", "filter1", "from0", "range3", "from2"},
		},
		{
			name: "diamond",
			//               join7
			//              /     \
			//        filter6     range5
			//              \     /
			//               join4
			//              /     \
			//        filter1      range3
			//          |            |
			//         from0        from2
			fluxQuery: `
				left = from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")
				right = from(bucket: "foo") |> range(start: -1d)
				j = join(tables: {l: left, r: right}, on: ["key"])
				right2 = range(start: -1y, table: j)
				left2 = filter(fn: (r) => r._value > 1.0, table: j)
				join(tables: {l: left2, r: right2}, on: ["key"])`,
			nodeIDs: []planner.NodeID{"join7", "filter6", "join4", "filter1", "from0", "range3", "from2", "range5"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now().UTC()
			spec, err := flux.Compile(context.Background(), tc.fluxQuery, now)
			if err != nil {
				t.Fatalf("Failed to create flux.Spec from text: %v", err)
			}

			var seenNodes []planner.NodeID
			simpleRule := plantest.CreateSimpleRuleFn(&seenNodes)()
			thePlanner := planner.NewLogicalPlanner(planner.WithLogicalRule(simpleRule))
			_, err = thePlanner.Plan(spec)
			if err != nil {
				t.Fatalf("Could not plan: %v", err)
			}

			if !cmp.Equal(tc.nodeIDs, seenNodes) {
				t.Errorf("Traversal didn't match expected, -want/+got:\n%v",
					cmp.Diff(tc.nodeIDs, seenNodes))
			}
		})
	}
}
