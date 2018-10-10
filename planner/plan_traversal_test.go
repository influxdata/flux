package planner_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/planner"
)

type SimpleRule struct {
	seenNodes []planner.NodeID
}

func (sr *SimpleRule) Pattern() planner.Pattern {
	return planner.Any()
}

func (sr *SimpleRule) Rewrite(node planner.PlanNode) (planner.PlanNode, bool) {
	sr.seenNodes = append(sr.seenNodes, node.ID())
	return node, false
}

func fluxToQueryPlan(fluxQuery string, a planner.Administration) (*planner.QueryPlan, error) {
	now := time.Now().UTC()
	spec, err := flux.Compile(context.Background(), fluxQuery, now)
	if err != nil {
		return nil, err
	}

	qp, err := planner.CreateLogicalPlan(spec, a)
	return qp, err
}

func TestPlanTraversal(t *testing.T) {

	testCases := []struct {
		name string
		fluxQuery string
		nodeIDs []planner.NodeID
	}{
		{
			name: "simple",
			fluxQuery: `from(bucket: "foo")`,
			nodeIDs: []planner.NodeID{"from0"},
		},
		{
			name: "from and filter",
			fluxQuery: `from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")`,
			nodeIDs: []planner.NodeID{"filter1", "from0"},
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
			//t.Parallel()

			simpleRule := SimpleRule{}
			aPlanner := planner.NewLogicalToPhysicalPlanner([]planner.Rule{&simpleRule})

			qp, err := fluxToQueryPlan(tc.fluxQuery, aPlanner)
			if err != nil {
				t.Fatalf("Could not convert Flux to logical plan: %v", err)
			}

			_, err = aPlanner.Plan(qp)
			if err != nil {
				t.Fatalf("Could not plan: %v", err)
			}

			if ! cmp.Equal(tc.nodeIDs, simpleRule.seenNodes) {
				t.Errorf("Traversal didn't match expected, -want/+got:\n%v",
					cmp.Diff(tc.nodeIDs, simpleRule.seenNodes))
			}
		})
	}
}
