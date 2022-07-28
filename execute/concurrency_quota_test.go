package execute_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux/plan"
	"go.uber.org/zap/zaptest"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	_ "github.com/influxdata/flux/fluxinit/static"
	fluxfeature "github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/internal/pkg/feature"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

//
// These test cases verify that execution engine concurrencyQuota is computed
// correctly.
//

// ParallelFromRemoteProcedureSpec implements a parallel from-remote procedure
// spec. Parallel factor can be set and it reports that it has parallel run
// attributes. Using this to create a mock plan with parallelization, for the
// purpose of testing the concurrency quota computation. The parallel plans
// here are not actually sensible parallelization strategies.
type ParallelFromRemoteProcedureSpec struct {
	*influxdb.FromRemoteProcedureSpec
	factor int
}

const ParallelFromRemoteTestKind = "parallel-from-remote-test"

func (src *ParallelFromRemoteProcedureSpec) OutputAttributes() plan.PhysicalAttributes {
	if src.factor > 1 {
		return plan.PhysicalAttributes{
			plan.ParallelRunKey: plan.ParallelRunAttribute{Factor: src.factor},
		}
	}
	return nil
}

func (src *ParallelFromRemoteProcedureSpec) Kind() plan.ProcedureKind {
	return ParallelFromRemoteTestKind
}

func (src *ParallelFromRemoteProcedureSpec) Copy() plan.ProcedureSpec {
	return src
}

func (src *ParallelFromRemoteProcedureSpec) Cost(inStats []plan.Statistics) (plan.Cost, plan.Statistics) {
	return plan.Cost{}, plan.Statistics{}
}

type parallelizeFromTo struct {
	Factor int
}

// parallelizeFromTo is a planner rule that adds a mock form of parallel
// execution, only for the purpose of testing the concurrency quota computation
// when parallelization is present. The test cases here compile flux from
// source code for the sake of convenience, so we need a planner rule that
// introduces parallelization.
func (parallelizeFromTo) Name() string {
	return "parallelizeFromTo"
}

func (parallelizeFromTo) Pattern() plan.Pattern {
	return plan.MultiSuccessor(influxdb.ToKind, plan.SingleSuccessor(influxdb.FromRemoteKind))
}

func (rule parallelizeFromTo) Rewrite(ctx context.Context, pn plan.Node) (plan.Node, bool, error) {
	toNode := pn
	toSpec := toNode.ProcedureSpec().(*influxdb.ToProcedureSpec)

	fromNode := toNode.Predecessors()[0]
	fromSpec := fromNode.ProcedureSpec().(*influxdb.FromRemoteProcedureSpec)

	physicalFromNode := fromNode.(*plan.PhysicalPlanNode)
	if attr := plan.GetOutputAttribute(physicalFromNode, plan.ParallelRunKey); attr != nil {
		return pn, false, nil
	}

	newFromNode := plan.CreateUniquePhysicalNode(ctx, "from", &ParallelFromRemoteProcedureSpec{
		FromRemoteProcedureSpec: fromSpec.Copy().(*influxdb.FromRemoteProcedureSpec),
		factor:                  rule.Factor,
	})
	newToNode := plan.CreateUniquePhysicalNode(ctx, "to", toSpec.Copy().(*influxdb.ToProcedureSpec))
	mergeNode := plan.CreateUniquePhysicalNode(ctx, "partitionMerge", &universe.PartitionMergeProcedureSpec{Factor: rule.Factor})

	newFromNode.AddSuccessors(newToNode)
	newToNode.AddPredecessors(newFromNode)

	newToNode.AddSuccessors(mergeNode)
	mergeNode.AddPredecessors(newToNode)

	return mergeNode, true, nil
}

type flagger map[string]interface{}

func compile(fluxText string, now time.Time) (context.Context, *flux.Spec, error) {
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	spec, err := spec.FromScript(ctx, runtime.Default, now, fluxText)
	return ctx, spec, err
}

func TestConcurrencyQuota(t *testing.T) {
	now := parser.MustParseTime("2022-01-01T10:00:00Z").Value

	testcases := []struct {
		name                     string
		fromPlan                 int
		flux                     string
		flagger                  flagger
		parallelizeFactor        int
		queryConcurrencyIncrease int
		wantConcurrencyQuota     int
	}{
		// The concurrency quota is not computed if it is already specified in
		// the plan. This first test case exercises that path.
		{
			name:                 "from-plan",
			flux:                 `from(bucket: "bucket", host: "host") |> range( start: 0 )`,
			fromPlan:             9,
			wantConcurrencyQuota: 9,
		},
		// Various number of result sets.
		{
			name:                 "one-result",
			flux:                 `from(bucket: "bucket", host: "host") |> range( start: 0 ) |> filter( fn: (r) => r.key == "value" )`,
			wantConcurrencyQuota: 1,
		},
		{
			name: "two-results",
			flux: `
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> filter( fn: (r) => r.key == "value" ) 
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> filter( fn: (r) => r.key == "value" )
			`,
			wantConcurrencyQuota: 2,
		},
		{
			name: "five-results",
			flux: `
				from(bucket: "bucket", host: "host") |> range( start: 0 )
				from(bucket: "bucket", host: "host") |> range( start: 0 )
				from(bucket: "bucket", host: "host") |> range( start: 0 )
				from(bucket: "bucket", host: "host") |> range( start: 0 )
				from(bucket: "bucket", host: "host") |> range( start: 0 )
			`,
			wantConcurrencyQuota: 5,
		},
		{
			name: "five-yields",
			flux: `
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n1" )
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n2" )
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n3" )
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n4" )
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n5" )
			`,
			wantConcurrencyQuota: 5,
		},
		// Yields should be recursed through, but there should not be any
		// duplicates if the yields are adjacent.
		{
			name: "chained-yields",
			flux: `
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n1" ) |> yield( name: "n2" ) |> yield( name: "n3" )
				from(bucket: "bucket", host: "host") |> range( start: 0 ) |> yield( name: "n4" ) |> yield( name: "n5" ) 
			`,
			wantConcurrencyQuota: 2,
		},
		// A result with multiple predecessors. There should be a goroutine
		// allocated for each predecessor.
		{
			name: "one-union",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
			`,
			wantConcurrencyQuota: 2,
		},
		{
			name: "two-union",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
				union( tables: [ s1, s2, s3, s4 ] )
			`,
			wantConcurrencyQuota: 6,
		},
		// Results with multiple predecessors should be found even if behind
		// yields.
		{
			name: "two-union-behind-yield-1",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] ) |> yield(name: "n1")
				union( tables: [ s1, s2, s3, s4 ] )
			`,
			wantConcurrencyQuota: 6,
		},
		{
			name: "two-union-behind-yield-2",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] ) |> yield(name: "n1") |> yield( name: "n2" )
				union( tables: [ s1, s2, s3, s4 ] ) |> yield( name: "n3" )
			`,
			wantConcurrencyQuota: 6,
		},
		// If a yield results in a distinct result set, it should add to the
		// number of goroutines.
		{
			name: "inline-yield-1",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range( start: 0 )
					|> yield( name: "n1" )
					|> filter( fn: (r) => r.t == "tv" )
			`,
			wantConcurrencyQuota: 2,
		},
		{
			name: "inline-yield-2",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range( start: 0 )
					|> yield( name: "n1" )
					|> filter( fn: (r) => r.t == "tv" )
					|> yield( name: "n2" )
			`,
			wantConcurrencyQuota: 2,
		},
		{
			name: "inline-yield-3",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range( start: 0 )
					|> yield( name: "n1" )
					|> yield( name: "n2" )
					|> filter( fn: (r) => r.t == "tv" )
					|> yield( name: "n3" )
			`,
			wantConcurrencyQuota: 2,
		},
		{
			name: "inline-yield-4",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range( start: 0 )
					|> yield( name: "n1" )
					|> filter( fn: (r) => r.t == "tv" )
					|> yield( name: "n2" )
					|> filter( fn: (r) => r.t == "tv" )
					|> yield( name: "n3" )
			`,
			wantConcurrencyQuota: 3,
		},
		{
			name: "inline-yield-5",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
					|> yield(name: "n1")
					|> yield( name: "n2" ) 
					|> filter( fn: (r) => r.t == "tv" )
				union( tables: [ s1, s2, s3, s4 ] )
			`,
			wantConcurrencyQuota: 7,
		},
		// Test using QueryConcurrencyIncrease to increase the concurrency
		// quota.
		{
			name: "increase-1",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range( start: 0 )
					|> filter( fn: (r) => r.key == "value" )`,
			queryConcurrencyIncrease: 1,
			wantConcurrencyQuota:     2,
		},
		{
			name: "increase-2",
			flux: `
				from(bucket: "bucket", host: "host") |> range( start: 0 )
				from(bucket: "bucket", host: "host") |> range( start: 0 )
			`,
			queryConcurrencyIncrease: 2,
			wantConcurrencyQuota:     4,
		},
		{
			name: "increase-3",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
					|> yield(name: "n1")
					|> yield( name: "n2" ) 
					|> filter( fn: (r) => r.t == "tv" )
				union( tables: [ s1, s2, s3, s4 ] )
			`,
			queryConcurrencyIncrease: 3,
			wantConcurrencyQuota:     10,
		},
		// Parallelization causes an increase in concurrency quota. The
		// parallel merge was not itself a result set, so we didn't increase
		// concurrency due to the multiple predecessors of the merge. A full 8
		// goroutines are added (2x parallelization factor).
		{
			name: "parallelize-none-accounted-for",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
					|> filter( fn: (r) => r.t == "tv" )
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 9,
		},
		// Parallelization causes an increase in concurrency quota. The
		// parallel merge is a result set, so 3 goroutines are added because of
		// that, a remaining 5 are added to keep the total added the same, as
		// if the merge was not a result set.
		{
			name: "parallelize-some-accounted-for",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 9,
		},
		// Parallelization causes an increase, however enough goroutines are
		// added due to parallel merge result sets, so no additional goroutines
		// are needed for parallelization.
		{
			name: "parallelize-behind-yield-all-accounted-for",
			flux: `
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
					|> yield(name: "n1")
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
					|> yield(name: "n2")
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
					|> yield(name: "n3")
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 12,
		},
		// Here 6 for the two unions + 1 for the distinct yield + 8 for
		// parallelization. The "n1" yield causes an increase so some are
		// already accounted for.
		{
			name: "two-union-parallel-some-accounted-for",
			flux: `
				s1 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s2 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
				union( tables: [ s1, s2, s3, s4 ] )
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
					|> yield(name: "n1")
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 15,
		},
		// 6 for the two unions + 8 for the parallelization behind the unions.
		// No goroutines are added due to the parallel merge result sets, so
		// the full 8 are added for parallelization.
		{
			name: "parallel-behind-union-none-accounted-for",
			flux: `
				s1 = from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				s2 = from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
				union( tables: [ s1, s2, s3, s4 ] )
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 14,
		},
		// 6 for the two unions + 12 for the three parallelized results. No
		// additional goroutines need to be added to account for the
		// parallelization because 9 are added due to parallel merge result
		// sets.
		{
			name: "parallel-behind-union-all-accounted-for",
			flux: `
				s1 = from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				s2 = from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				s3 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				s4 = from(bucket: "bucket", host: "host") |> range( start: 0 )
				union( tables: [ s1, s2 ] )
				union( tables: [ s1, s2, s3, s4 ] )
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
				from(bucket: "bucket", host: "host")
					|> range(start: 0)
					|> to(bucket:"other-bucket")
			`,
			parallelizeFactor:    4,
			wantConcurrencyQuota: 18,
		},
		// Test interplay with tableFind. The input to tableFind has a yield,
		// which gets scrubbed from the tableFind, and thus does not increase
		// the concurrency quota of the sub-plan. It stays at 1. We can't test
		// that computation though. The yield does emerge in the top-level
		// plan, so the concurrency quota will be 2.
		{
			name: "tablefind-1",
			flux: `
				import "sampledata"

				x = sampledata.int()
				  |> toFloat() 
				  |> yield(name: "n1")

				t = x
				    |> map(fn: (r) => ({r with left: 0.0, right: 1.0 }))
				    |> tableFind(fn: (key) => true) |> getRecord(idx: 0)

				xbar = t.left / t.right

				x |> map(fn: (r) => ({r with xbar: xbar}))
			`,
			wantConcurrencyQuota: 2,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// We are compiling so we can specify large plans with just some
			// flux code.
			ctx, fluxSpec, err := compile(tc.flux, now)

			flagger := executetest.TestFlagger{}
			flagger[fluxfeature.QueryConcurrencyIncrease().Key()] = tc.queryConcurrencyIncrease
			ctx = feature.Inject(ctx, flagger)

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

			// We need a physical from in order to complete planning, and we
			// also need ranges merged in so the physical from is satisfied.
			rules := []plan.Rule{&influxdb.FromRemoteRule{}, &influxdb.MergeRemoteRangeRule{}}

			if tc.parallelizeFactor > 0 {
				rules = append(rules,
					&influxdb.MergeRemoteRangeRule{}, &parallelizeFromTo{Factor: tc.parallelizeFactor})
			}

			physicalPlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(rules...))

			physicalPlan, err := physicalPlanner.Plan(context.Background(), logicalPlan)
			if err != nil {
				t.Fatal(err)
			}

			if tc.fromPlan > 0 {
				physicalPlan.Resources.ConcurrencyQuota = tc.fromPlan
			}

			// This is a test helper that constructs a basic execution state
			// and chooses the default resources. It then returns the
			// concurrency quota.
			concurrencyQuota := execute.ConcurrencyQuotaFromPlan(ctx, physicalPlan, zaptest.NewLogger(t))

			if concurrencyQuota != tc.wantConcurrencyQuota {
				t.Errorf("Expected concurrency quota of %v, but execution state has %v",
					tc.wantConcurrencyQuota, concurrencyQuota)
			}
		})
	}
}
