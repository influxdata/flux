package plan_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
)

type triggerAwareProcedureSpec struct {
	plan.DefaultCost
}

func (s triggerAwareProcedureSpec) Copy() plan.ProcedureSpec {
	return s
}

func (s triggerAwareProcedureSpec) Kind() plan.ProcedureKind {
	return "TriggerAwareProcedure"
}

func (s triggerAwareProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func TestTriggers(t *testing.T) {
	testcases := []struct {
		name    string
		node    plan.PlanNode
		want    plan.TriggerSpec
		wantErr bool
	}{
		{
			name: "default trigger spec",
			node: &plan.PhysicalPlanNode{},
			want: plan.AfterWatermarkTriggerSpec{},
		},
		{
			name: "trigger aware procedure",
			node: &plan.PhysicalPlanNode{
				Spec: triggerAwareProcedureSpec{},
			},
			want: plan.NarrowTransformationTriggerSpec{},
		},
		{
			name:    "cannot set trigger on logical node",
			node:    &plan.LogicalPlanNode{},
			wantErr: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := plan.SetTriggerSpec(tc.node)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error setting default triggers: %v", err)
			}
			if tc.wantErr && err == nil {
				t.Fatal("expected error setting default triggers, but got nothing")
			}
			if tc.wantErr == (err != nil) {
				return
			}
			n := tc.node.(*plan.PhysicalPlanNode)
			if !cmp.Equal(tc.want, n.TriggerSpec) {
				t.Fatalf("unexpected trigger spec: -want/+got\n%s", cmp.Diff(tc.want, n.TriggerSpec))
			}
		})
	}
}
