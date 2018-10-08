package planner

import (
	"fmt"
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/semantic"
)

var create = map[flux.OperationKind]CreateLogicalProcedureSpec{
	// Take a FromOpSpec and translate it to a FromProcedureSpec
	functions.FromKind: func(op flux.OperationSpec) (LogicalProcedureSpec, error) {
		spec, ok := op.(*functions.FromOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &FromProcedureSpec{
			Bucket:   spec.Bucket,
			BucketID: spec.BucketID,
		}, nil
	},
	// Take a RangeOpSpec and convert it to a RangeProcedureSpec
	functions.RangeKind: func(op flux.OperationSpec) (LogicalProcedureSpec, error) {
		spec, ok := op.(*functions.RangeOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		if spec.TimeCol == "" {
			spec.TimeCol = execute.DefaultTimeColLabel
		}

		return &RangeProcedureSpec{
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
	functions.FilterKind: func(op flux.OperationSpec) (LogicalProcedureSpec, error) {
		spec, ok := op.(*functions.FilterOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &FilterProcedureSpec{
			Fn: spec.Fn.Copy().(*semantic.FunctionExpression),
		}, nil
	},
	functions.YieldKind: func(op flux.OperationSpec) (LogicalProcedureSpec, error) {
		spec, ok := op.(*functions.YieldOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &YieldProcedureSpec{
			Name: spec.Name,
		}, nil
	},
	functions.JoinKind: func(op flux.OperationSpec) (LogicalProcedureSpec, error) {
		spec, ok := op.(*functions.JoinOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &JoinProcedureSpec{
			On: spec.On,
		}, nil
	},
}

type SimpleRule struct {
	seenNodes []NodeID
}

func (sr *SimpleRule) Pattern() Pattern {
	return Any()
}

func (sr *SimpleRule) Rewrite(node PlanNode) (PlanNode, bool) {
	sr.seenNodes = append(sr.seenNodes, node.ID())
	return node, false
}

func fluxToQueryPlan(fluxQuery string) (*QueryPlan, error) {
	now := time.Now().UTC()
	spec, err := flux.Compile(context.Background(), fluxQuery, now)
	if err != nil {
		return nil, err
	}

	qp, err := CreateLogicalPlan(spec, create)
	return qp, err
}

func TestPlanTraversal(t *testing.T) {

	testCases := []struct {
		name string
		fluxQuery string
		nodeIDs []NodeID
	}{
		{
			name: "simple",
			fluxQuery: `from(bucket: "foo")`,
			nodeIDs: []NodeID{"from0"},
		},
		{
			name: "from and filter",
			fluxQuery: `from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")`,
			nodeIDs: []NodeID{"filter1", "from0"},
		},
		{
			name: "join",
			fluxQuery: `
			    left = from(bucket: "foo") |> filter(fn: (r) => r._field == "cpu")
                right = from(bucket: "foo") |> range(start: -1d)
                join(tables: {l: left, r: right}, on: ["key"]) |> yield()`,
			nodeIDs: []NodeID{"yield5", "join4", "filter1", "from0", "range3", "from2"},
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
			nodeIDs: []NodeID{"join7", "filter6", "join4", "filter1", "from0", "range3", "from2", "range5"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simpleRule := SimpleRule{}
			planner := NewLogicalToPhysicalPlanner([]Rule{&simpleRule})

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
