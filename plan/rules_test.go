package plan_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

func init() {
	plan.RegisterLogicalRules(
		influxdb.DefaultFromAttributes{
			Org:  &influxdb.NameOrID{Name: "influxdata"},
			Host: func(v string) *string { return &v }("http://localhost:9999"),
		},
	)
}

func TestRuleRegistration(t *testing.T) {
	simpleRule := plantest.SimpleRule{}

	// Register the rule,
	// then check seenNodes below to check that the rule was invoked.
	plan.RegisterLogicalRules(&simpleRule)

	now := time.Now().UTC()
	fluxSpec, err := spec.FromScript(dependenciestest.Default().Inject(context.Background()), now, `from(bucket: "telegraf") |> range(start: -5m)`)
	if err != nil {
		t.Fatalf("could not compile very simple Flux query: %v", err)
	}

	logicalPlanner := plan.NewLogicalPlanner()
	initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
	if err != nil {
		t.Fatal(err)
	}
	logicalPlanSpec, err := logicalPlanner.Plan(initPlan)
	if err != nil {
		t.Fatalf("could not do logical planning: %v", err)
	}

	wantSeenNodes := []plan.NodeID{"generated_yield", "range1", "from0"}
	if !cmp.Equal(wantSeenNodes, simpleRule.SeenNodes) {
		t.Errorf("did not find expected seen nodes, -want/+got:\n%v", cmp.Diff(wantSeenNodes, simpleRule.SeenNodes))
	}

	// Test rule registration for the physical plan too.
	simpleRule.SeenNodes = simpleRule.SeenNodes[0:0]
	plan.RegisterPhysicalRules(&simpleRule)

	physicalPlanner := plan.NewPhysicalPlanner()
	_, err = physicalPlanner.Plan(logicalPlanSpec)
	if err != nil {
		t.Fatalf("could not do physical planning: %v", err)
	}

	// This test will be fragile if we lock down the actual nodes seen,
	// so just pass if we saw anything.
	if len(simpleRule.SeenNodes) == 0 {
		t.Errorf("expected simpleRule to have been registered and have seen some nodes")
	}
}
