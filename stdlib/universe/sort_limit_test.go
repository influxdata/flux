package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestSortLimitRule(t *testing.T) {
	ctx, deps := dependency.Inject(context.Background(), executetest.NewTestExecuteDependencies())
	defer deps.Finish()

	from := &influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "testbucket"},
	}
	sort := &universe.SortProcedureSpec{
		Columns: []string{execute.DefaultValueColLabel},
	}
	limit0 := &universe.LimitProcedureSpec{N: 5}
	limit1 := &universe.LimitProcedureSpec{N: 1, Offset: 5}
	min := &universe.MinProcedureSpec{
		SelectorConfig: execute.SelectorConfig{
			Column: execute.DefaultValueColLabel,
		},
	}

	tests := []plantest.RuleTestCase{
		{
			Name:    "Default",
			Context: ctx,
			Rules: []plan.Rule{
				universe.SortLimitRule{},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("sort1", sort),
					plan.CreatePhysicalNode("limit2", limit0),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("merged_sort1_limit2", &universe.SortLimitProcedureSpec{
						SortProcedureSpec: sort,
						N:                 5,
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			SkipValidation: true,
		},
		{
			Name:    "WithOffset",
			Context: ctx,
			Rules: []plan.Rule{
				universe.SortLimitRule{},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("sort1", sort),
					plan.CreatePhysicalNode("limit2", limit1),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			NoChange:       true,
			SkipValidation: true,
		},
		{
			Name:    "MultipleSuccessors",
			Context: ctx,
			Rules: []plan.Rule{
				universe.SortLimitRule{},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("sort1", sort),
					plan.CreatePhysicalNode("limit2", limit0),
					plan.CreatePhysicalNode("min3", min),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("merged_sort1_limit2", &universe.SortLimitProcedureSpec{
						SortProcedureSpec: sort,
						N:                 5,
					}),
					plan.CreatePhysicalNode("min3", min),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
				},
			},
			SkipValidation: true,
		},
		{
			Name:    "SplitPattern",
			Context: ctx,
			Rules: []plan.Rule{
				universe.SortLimitRule{},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from0", from),
					plan.CreatePhysicalNode("sort1", sort),
					plan.CreatePhysicalNode("limit2", limit0),
					plan.CreatePhysicalNode("min3", min),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{1, 3},
				},
			},
			NoChange:       true,
			SkipValidation: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			plantest.PhysicalRuleTestHelper(t, &tc)
		})
	}
}
