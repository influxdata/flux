package planner_test

import (
	"context"
	"fmt"
	"github.com/influxdata/flux/planner"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/semantic"
)

var create = map[flux.OperationKind]planner.CreateLogicalProcedureSpec{
	// Take a FromOpSpec and translate it to a FromProcedureSpec
	inputs.FromKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*inputs.FromOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.FromProcedureSpec{
			Bucket:   spec.Bucket,
			BucketID: spec.BucketID,
		}, nil
	},
	// Take a RangeOpSpec and convert it to a RangeProcedureSpec
	transformations.RangeKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.RangeOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		if spec.TimeCol == "" {
			spec.TimeCol = execute.DefaultTimeColLabel
		}

		return &planner.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: spec.Start,
				Stop:  spec.Stop,
			},
			TimeCol:  spec.TimeCol,
			StartCol: spec.StartCol,
			StopCol:  spec.StopCol,
		}, nil
	},
	// Take a FilterOpSpec and translate it to a FilterProcedureSpec
	transformations.FilterKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.FilterOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.FilterProcedureSpec{
			Fn: spec.Fn.Copy().(*semantic.FunctionExpression),
		}, nil
	},
	transformations.YieldKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.YieldOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.YieldProcedureSpec{
			Name: spec.Name,
		}, nil
	},
	transformations.JoinKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.JoinOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.JoinProcedureSpec{
			On: spec.On,
		}, nil
	},
}

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

func fluxToQueryPlan(fluxQuery string) (*planner.QueryPlan, error) {
	now := time.Now().UTC()
	spec, err := flux.Compile(context.Background(), fluxQuery, now)
	if err != nil {
		return nil, err
	}

	qp, err := planner.CreateLogicalPlan(spec, create)
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
			planner := planner.NewLogicalToPhysicalPlanner([]planner.Rule{&simpleRule})

			qp, err := fluxToQueryPlan(tc.fluxQuery)
			if err != nil {
				t.Fatalf("Could not convert Flux to logical plan: %v", err)
			}

			_, err = planner.Plan(qp)
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
