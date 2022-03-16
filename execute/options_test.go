package execute

import (
	"context"
	"math"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
	planspec "github.com/influxdata/flux/plan/plantest/spec"
	"go.uber.org/zap/zaptest"
)

func TestExecuteOptions(t *testing.T) {
	type runWith struct {
		concurrencyQuota int
		memoryBytesQuota int64
	}

	testcases := []struct {
		name               string
		spec               *planspec.PlanSpec
		concurrencyLimit   int
		defaultMemoryLimit int64
		want               runWith
	}{
		{
			// If the concurrency quota and max bytes are set in the plan
			// resources, the execution state always uses those resources.
			// Historically, the values in the plan resources always took
			// precedence.
			name: "via-plan-no-options",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
				},
				Resources: flux.ResourceManagement{
					MemoryBytesQuota: 163484,
					ConcurrencyQuota: 4,
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			want: runWith{
				memoryBytesQuota: 163484,
				concurrencyQuota: 4,
			},
		},
		{
			// Use the plan resources even if the execution options are set.
			name: "via-plan-with-exec-options",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
				},
				Resources: flux.ResourceManagement{
					MemoryBytesQuota: 163484,
					ConcurrencyQuota: 4,
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			defaultMemoryLimit: 8,
			concurrencyLimit:   2,

			want: runWith{
				memoryBytesQuota: 163484,
				concurrencyQuota: 4,
			},
		},
		{
			// Choose resources based on the default execute options. We get
			// old behaviour of choosing concurrency quota based on the number
			// of roots in the plan.
			name: "defaults-one-root",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
					planspec.CreatePhysicalMockNode("2"),
					planspec.CreatePhysicalMockNode("3"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			want: runWith{
				memoryBytesQuota: math.MaxInt64,
				concurrencyQuota: 1,
			},
		},
		{
			// Again use the default execute options. Two roots in the plan
			// means we get a concurrency quota of two.
			name: "defaults-two-roots",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
					planspec.CreatePhysicalMockNode("root-0"),
					planspec.CreatePhysicalMockNode("root-1"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{1, 3},
				},
			},
			want: runWith{
				memoryBytesQuota: math.MaxInt64,
				concurrencyQuota: 2,
			},
		},
		{
			// Set the execute options. The memory limit passes in verbatim.
			// The concurrency limit is 16 and the new behaviour of choosing
			// the concurreny quota based on the number of non-source nodes is
			// active.
			name: "via-options-new-behaviour-non-source",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
					planspec.CreatePhysicalMockNode("2"),
					planspec.CreatePhysicalMockNode("3"),
					planspec.CreatePhysicalMockNode("root-0"),
					planspec.CreatePhysicalMockNode("root-1"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
					{3, 5},
				},
			},
			defaultMemoryLimit: 32768,
			concurrencyLimit:   16,
			want: runWith{
				memoryBytesQuota: 32768,
				concurrencyQuota: 5,
			},
		},
		{
			// Set the execute options. We want the new behaviour of setting
			// concurrency quota based on the number of non-source nodes (5),
			// but limited by the concurrency limit (4).
			name: "via-options-new-behaviour-limited",
			spec: &planspec.PlanSpec{
				Nodes: []plan.Node{
					planspec.CreatePhysicalMockNode("0"),
					planspec.CreatePhysicalMockNode("1"),
					planspec.CreatePhysicalMockNode("2"),
					planspec.CreatePhysicalMockNode("3"),
					planspec.CreatePhysicalMockNode("root-0"),
					planspec.CreatePhysicalMockNode("root-1"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
					{3, 5},
				},
			},
			defaultMemoryLimit: 32768,
			concurrencyLimit:   4,
			want: runWith{
				memoryBytesQuota: 32768,
				concurrencyQuota: 4,
			},
		},
	}

	for _, tc := range testcases {
		execDeps := NewExecutionDependencies(nil, nil, nil)
		ctx := execDeps.Inject(context.Background())

		inputPlan := planspec.CreatePlanSpec(tc.spec)

		thePlanner := plan.NewPhysicalPlanner()
		outputPlan, err := thePlanner.Plan(context.Background(), inputPlan)
		if err != nil {
			t.Fatalf("Physical planning failed: %v", err)
		}

		//
		// Modify the execution options. In practice, we would do this from
		// planner rules
		//
		if tc.defaultMemoryLimit != 0 {
			execDeps.ExecutionOptions.DefaultMemoryLimit = tc.defaultMemoryLimit
		}
		if tc.concurrencyLimit != 0 {
			execDeps.ExecutionOptions.ConcurrencyLimit = tc.concurrencyLimit
		}

		// Construct a basic execution state and choose the default resources.
		es := &executionState{
			p:         outputPlan,
			ctx:       ctx,
			resources: outputPlan.Resources,
			logger:    zaptest.NewLogger(t),
		}
		es.chooseDefaultResources(ctx, outputPlan)

		if err := es.validate(); err != nil {
			t.Fatalf("execution state failed validation: %s", err.Error())
		}

		if es.resources.MemoryBytesQuota != tc.want.memoryBytesQuota {
			t.Errorf("Expected memory quota of %v, but execution state has %v",
				tc.want.memoryBytesQuota, es.resources.MemoryBytesQuota)
		}

		if es.resources.ConcurrencyQuota != tc.want.concurrencyQuota {
			t.Errorf("Expected concurrency quota of %v, but execution state has %v",
				tc.want.concurrencyQuota, es.resources.ConcurrencyQuota)
		}
	}
}
