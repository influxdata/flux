package join_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/join"
	"github.com/influxdata/flux/stdlib/universe"
)

func compile(fluxText string, now time.Time) (*flux.Spec, error) {
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	return spec.FromScript(ctx, runtime.Default, now, fluxText)
}

func TestEquiJoinPredicateRule(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name      string
		flux      string
		wantErr   error
		wantPairs []join.ColumnPair
		wantPlan  *plantest.PlanSpec
	}{
		{
			name: "single comparison",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => l.a == r.b,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantPairs: []join.ColumnPair{
				join.ColumnPair{Left: "a", Right: "b"},
			},
			wantPlan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from0", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter1", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("from2", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter3", &universe.FilterProcedureSpec{}),
					plan.CreatePhysicalNode("join.tables4", &join.EquiJoinProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
				},
				Now: now,
			},
		},
		{
			name: "multiple comparisons",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => l.a == r.b and l.e == r.f,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantPairs: []join.ColumnPair{
				join.ColumnPair{Left: "a", Right: "b"},
				join.ColumnPair{Left: "e", Right: "f"},
			},
			wantPlan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from0", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter1", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("from2", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter3", &universe.FilterProcedureSpec{}),
					plan.CreatePhysicalNode("join.tables", &join.EquiJoinProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
				},
				Now: now,
			},
		},
		{
			name: "reject or",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => l.a == r.b or l.e == r.f,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantErr: errors.New(
				codes.Invalid,
				"error in join function - some expressions are not yet supported in the `on` parameter: unsupported operator in join predicate: or",
			),
		},
		{
			name: "reject intratable comparisons",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => l.a == l.b,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantErr: errors.New(
				codes.Invalid,
				"error in join function - some expressions are not yet supported in the `on` parameter: binary expression operands must reference `l` or `r` only, and may not reference the same object",
			),
		},
		{
			name: "r == l",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => r.a == l.b,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantPairs: []join.ColumnPair{
				join.ColumnPair{Left: "b", Right: "a"},
			},
			wantPlan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from0", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter1", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("from2", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter3", &universe.FilterProcedureSpec{}),
					plan.CreatePhysicalNode("join.tables", &join.EquiJoinProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
				},
				Now: now,
			},
		},
		{
			name: "multiple comparison same column",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => r.a == l.b and r.d == l.b,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantPairs: []join.ColumnPair{
				join.ColumnPair{Left: "b", Right: "a"},
				join.ColumnPair{Left: "b", Right: "d"},
			},
			wantPlan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from0", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter1", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("from2", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter3", &universe.FilterProcedureSpec{}),
					plan.CreatePhysicalNode("join.tables4", &join.EquiJoinProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
				},
				Now: now,
			},
		},
		{
			name: "illegal expression",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => true,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantErr: errors.New(
				codes.Invalid,
				"error in join function - some expressions are not yet supported in the `on` parameter: illegal expression type in join predicate: BooleanLiteral",
			),
		},
		{
			name: "multiple statements",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => {
					v = true
					return v
				},
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantErr: errors.New(
				codes.Invalid,
				"error in join function - some expressions are not yet supported in the `on` parameter: function body should be a single logical expression that compares columns from each table",
			),
		},
		{
			name: "multiple comparisons invalid operator",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.tables(
				left: left,
				right: right,
				on: (l, r) => l.a == r.b and l.e != r.f,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantErr: errors.New(
				codes.Invalid,
				"error in join function - some expressions are not yet supported in the `on` parameter: unsupported operator in join predicate: !=",
			),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fluxSpec, err := compile(tc.flux, now)
			if err != nil {
				t.Fatalf("could not compile flux query: %v", err)
			}

			logicalPlanner := plan.NewLogicalPlanner()
			initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
			if err != nil {
				t.Fatal(err)
			}
			logicalPlan, err := logicalPlanner.Plan(context.Background(), initPlan)
			if err != nil {
				t.Fatal(err)
			}
			physicalPlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(&join.EquiJoinPredicateRule{}))
			physicalPlan, err := physicalPlanner.Plan(context.Background(), logicalPlan)
			if err != nil {
				if tc.wantErr != nil {
					if tc.wantErr.Error() != err.Error() {
						t.Fatalf("expected error: %s - got %s", tc.wantErr, err)
					}
					return
				} else {
					t.Fatalf("got unexpected error: %s", err)
				}
			} else {
				if tc.wantErr != nil {
					t.Fatalf("expected error `%s` - got none", tc.wantErr)
				}
			}

			var pairs []join.ColumnPair
			for node := range physicalPlan.Roots {
				spec, ok := node.ProcedureSpec().(*join.EquiJoinProcedureSpec)
				if !ok {
					continue
				}
				pairs = spec.On
				break
			}
			if diff := cmp.Diff(tc.wantPairs, pairs); diff != "" {
				t.Errorf("unexpected column pairs; -want/+got:\n%v", diff)
			}

			wantPlan := plantest.CreatePlanSpec(tc.wantPlan)
			if err := plantest.ComparePlansShallow(wantPlan, physicalPlan); err != nil {
				t.Error(err)
			}
		})
	}
}
