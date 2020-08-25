package plan_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/values"
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
	plan.ClearRegisteredRules()

	simpleRule := plantest.SimpleRule{}

	// Register the rule,
	// then check seenNodes below to check that the rule was invoked.
	plan.RegisterLogicalRules(&simpleRule)

	now := time.Now().UTC()
	fluxSpec, err := spec.FromScript(dependenciestest.Default().Inject(context.Background()), runtime.Default, now,
		`from(host: "http://localhost:9999", bucket: "telegraf") |> range(start: -5m)`)
	if err != nil {
		t.Fatalf("could not compile very simple Flux query: %v", err)
	}

	logicalPlanner := plan.NewLogicalPlanner()
	initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
	if err != nil {
		t.Fatal(err)
	}
	logicalPlanSpec, err := logicalPlanner.Plan(context.Background(), initPlan)
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
	_, err = physicalPlanner.Plan(context.Background(), logicalPlanSpec)
	if err != nil {
		t.Fatalf("could not do physical planning: %v", err)
	}

	// This test will be fragile if we lock down the actual nodes seen,
	// so just pass if we saw anything.
	if len(simpleRule.SeenNodes) == 0 {
		t.Errorf("expected simpleRule to have been registered and have seen some nodes")
	}
}

func TestRewriteWithContext(t *testing.T) {
	plan.ClearRegisteredRules()

	var (
		ctxKey  = "contextKey"
		rewrite = false
		value   interface{}
	)
	functionRule := plantest.FunctionRule{
		RewriteFn: func(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
			rewrite = true
			value = ctx.Value(ctxKey)
			return node, false, nil
		},
	}

	// Define the context after the above to ensure we don't end up accidentally reading
	// from the outer context rather than the one passed to the function.
	ctx := context.WithValue(context.Background(), ctxKey, true)
	// Register the rule.
	plan.RegisterLogicalRules(&functionRule)

	now := time.Now().UTC()
	fluxSpec, err := spec.FromScript(dependenciestest.Default().Inject(ctx), runtime.Default, now,
		`from(host: "http://localhost:9999", bucket: "telegraf") |> range(start: -5m)`)
	if err != nil {
		t.Fatalf("could not compile very simple Flux query: %v", err)
	}

	logicalPlanner := plan.NewLogicalPlanner()
	initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
	if err != nil {
		t.Fatal(err)
	}
	logicalPlanSpec, err := logicalPlanner.Plan(ctx, initPlan)
	if err != nil {
		t.Fatalf("could not do logical planning: %v", err)
	}

	if !rewrite {
		t.Fatal("logical planning did not call rewrite on the function rule")
	} else if value == nil {
		t.Fatal("value wasn't present in the context")
	}

	// Reset the values that were modified.
	rewrite, value = false, nil

	// Register the same rule with the physical planner.
	plan.RegisterPhysicalRules(&functionRule)

	physicalPlanner := plan.NewPhysicalPlanner()
	_, err = physicalPlanner.Plan(ctx, logicalPlanSpec)
	if err != nil {
		t.Fatalf("could not do physical planning: %v", err)
	}

	if !rewrite {
		t.Fatal("physical planning did not call rewrite on the function rule")
	} else if value == nil {
		t.Fatal("value wasn't present in the context")
	}
}

func TestRewriteWithContext_TableObjectCompiler(t *testing.T) {
	plan.ClearRegisteredRules()

	var (
		ctxKey  = "contextKey"
		rewrite = false
		value   interface{}
	)
	functionRule := plantest.FunctionRule{
		RewriteFn: func(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
			rewrite = true
			value = ctx.Value(ctxKey)
			return node, false, nil
		},
	}

	// Define the context after the above to ensure we don't end up accidentally reading
	// from the outer context rather than the one passed to the function.
	ctx := context.WithValue(context.Background(), ctxKey, true)
	// Register the rule.
	plan.RegisterLogicalRules(&functionRule)

	prelude := runtime.Prelude()
	buckets, _ := prelude.Lookup("buckets")
	args := values.NewObjectWithValues(map[string]values.Value{
		"host": values.NewString("http://localhost:9999"),
	})
	res, err := buckets.Function().Call(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	to := res.(*flux.TableObject)

	now := time.Now().UTC()
	if _, err := lang.CompileTableObject(ctx, to, now); err != nil {
		t.Fatal(err)
	}

	if !rewrite {
		t.Fatal("planning did not call rewrite on the function rule")
	} else if value == nil {
		t.Fatal("value wasn't present in the context")
	}
}

func TestMultiRootMatch(t *testing.T) {
	plan.ClearRegisteredRules()

	multiRootRule := plantest.MultiRootRule{}

	// Register the rule,
	// then check seenNodes below to check that the rule was invoked.
	plan.RegisterLogicalRules(&multiRootRule)

	now := time.Now().UTC()
	fluxSpec, err := spec.FromScript(dependenciestest.Default().Inject(context.Background()), runtime.Default, now,
		`from(host: "http://localhost:9999", bucket: "telegraf") |> range(start: -5m) |> min() |> max() |> mean()`)
	if err != nil {
		t.Fatalf("could not compile very simple Flux query: %v", err)
	}

	logicalPlanner := plan.NewLogicalPlanner()
	initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
	if err != nil {
		t.Fatal(err)
	}
	_, err = logicalPlanner.Plan(context.Background(), initPlan)
	if err != nil {
		t.Fatalf("could not do logical planning: %v", err)
	}

	// Only the min/max/means should match, not the from or range.
	wantSeenNodes := []plan.NodeID{"mean4", "max3", "min2"}
	if !cmp.Equal(wantSeenNodes, multiRootRule.SeenNodes) {
		t.Errorf("did not find expected seen nodes, -want/+got:\n%v", cmp.Diff(wantSeenNodes, multiRootRule.SeenNodes))
	}
}
