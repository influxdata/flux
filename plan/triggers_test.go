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
		name string
		node plan.Node
		want plan.TriggerSpec
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
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if err := plan.SetTriggerSpec(tc.node); err != nil {
				t.Fatalf("unexpected error setting trigger spec: %v", err)
			}
			n := tc.node.(*plan.PhysicalPlanNode)
			if !cmp.Equal(tc.want, n.TriggerSpec) {
				t.Fatalf("unexpected trigger spec: -want/+got\n%s", cmp.Diff(tc.want, n.TriggerSpec))
			}
		})
	}
}
