package planner_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
)

func TestRuleRegistration(t *testing.T) {
	var seenNodes []planner.NodeID

	// The rule created by this function matches any node, and just reports
	// what it has seen to seenNodes.
	simpleRuleCreate := plantest.CreateSimpleRuleFn(&seenNodes)

	// Register the rule,
	// then check seenNodes below to check that the rule was invoked.
	planner.RegisterLogicalRule(simpleRuleCreate)

	now := time.Now().UTC()
	fluxSpec, err := flux.Compile(context.Background(), `from(bucket: "telegraf") |> range(start: -5m)`, now)
	if err != nil {
		t.Fatalf("could not compile very simple Flux query: %v", err)
	}

	logicalPlanner := planner.NewLogicalPlanner()
	logicalPlanSpec, err := logicalPlanner.Plan(fluxSpec)
	if err != nil {
		t.Fatalf("could not do logical planning: %v", err)
	}

	wantSeenNodes := []planner.NodeID{"range1", "from0"}
	if !cmp.Equal(wantSeenNodes, seenNodes) {
		t.Errorf("did not find expected seen nodes, -want/+got:\n%v", cmp.Diff(wantSeenNodes, seenNodes))
	}

	// Test rule registration for the physical planner too.
	seenNodes = seenNodes[0:0]
	planner.RegisterPhysicalRule(simpleRuleCreate)

	physicalPlanner := planner.NewPhysicalPlanner()
	_, err = physicalPlanner.Plan(logicalPlanSpec)
	if err != nil {
		t.Fatalf("could not do physical planning: %v", err)
	}

	if !cmp.Equal(wantSeenNodes, seenNodes) {
		t.Errorf("did not find expected seen nodes, -want/+got:\n%v", cmp.Diff(wantSeenNodes, seenNodes))
	}
}
