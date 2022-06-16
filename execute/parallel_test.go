package execute_test

import (
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	_ "github.com/influxdata/flux/fluxinit/static"
)

func init() {
	// We depend on the registrations that happen in executor_test.go
	execute.RegisterSource(executetest.ParallelFromTestKind, executetest.CreateParallelFromSource)
}

//func TestParallel_Execute(t *testing.T) {
//
//	testcases := []struct {
//		name              string
//		spec              *plantest.PlanSpec
//		want              map[string][]*executetest.Table
//		allocator         memory.Allocator
//		wantErr           error
//		wantValidationErr error
//	}{
//		{
//			// The from node is executed in parallel, then the data is merged,
//			// and finally filtered after the merge.
//			name: `parallel-from-merge-filter`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec(
//							[]*executetest.ParallelTable{
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, -1},
//										},
//									},
//									ResidesOnPartition: 0,
//								},
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(8), 8.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(9), 9.0, -1},
//										},
//									},
//									ResidesOnPartition: 1,
//								},
//							}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("filter", &universe.FilterProcedureSpec{
//						Fn: interpreter.ResolvedFunction{
//							Scope: runtime.Prelude(),
//							Fn:    executetest.FunctionExpression(t, "(r) => r._value < 7.5"),
//						},
//					}),
//					plantest.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
//				},
//				Edges: [][2]int{
//					{0, 1},
//					{1, 2},
//					{2, 3},
//				},
//			},
//			want: map[string][]*executetest.Table{
//				"_result": []*executetest.Table{
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, int64(0)},
//						},
//					},
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, int64(1)},
//						},
//					},
//				},
//			},
//		},
//		{
//			// The from and filter nodes are both executed in parallel, then
//			// the data is merged.
//			name: `parallel-from-filter-merge`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec(
//							[]*executetest.ParallelTable{
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, -1},
//										},
//									},
//									ResidesOnPartition: 0,
//								},
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(8), 8.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(9), 9.0, -1},
//										},
//									},
//									ResidesOnPartition: 1,
//								},
//							}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("filter",
//						&universe.FilterProcedureSpec{
//							Fn: interpreter.ResolvedFunction{
//								Scope: runtime.Prelude(),
//								Fn:    executetest.FunctionExpression(t, "(r) => r._value < 7.5"),
//							},
//						},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
//				},
//				Edges: [][2]int{
//					{0, 1},
//					{1, 2},
//					{2, 3},
//				},
//			},
//			want: map[string][]*executetest.Table{
//				"_result": []*executetest.Table{
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, int64(0)},
//						},
//					},
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, int64(1)},
//						},
//					},
//				},
//			},
//		},
//		{
//			name: `parallel-from-merge-no-successor`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec(
//							[]*executetest.ParallelTable{
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, -1},
//											{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, -1},
//										},
//									},
//									ResidesOnPartition: 0,
//								},
//								{
//									Table: &executetest.Table{
//										KeyCols: []string{"_start", "_stop"},
//										ColMeta: []flux.ColMeta{
//											{Label: "_start", Type: flux.TTime},
//											{Label: "_stop", Type: flux.TTime},
//											{Label: "_time", Type: flux.TTime},
//											{Label: "_value", Type: flux.TFloat},
//											{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//										},
//										Data: [][]interface{}{
//											{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(8), 8.0, -1},
//											{execute.Time(5), execute.Time(10), execute.Time(9), 9.0, -1},
//										},
//									},
//									ResidesOnPartition: 1,
//								},
//							}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 2})),
//				},
//				Edges: [][2]int{
//					{0, 1},
//				},
//			},
//			want: map[string][]*executetest.Table{
//				"_result": []*executetest.Table{
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, int64(0)},
//							{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, int64(0)},
//						},
//					},
//					{
//						KeyCols: []string{"_start", "_stop"},
//						ColMeta: []flux.ColMeta{
//							{Label: "_start", Type: flux.TTime},
//							{Label: "_stop", Type: flux.TTime},
//							{Label: "_time", Type: flux.TTime},
//							{Label: "_value", Type: flux.TFloat},
//							{Label: executetest.ParallelGroupColName, Type: flux.TInt},
//						},
//						Data: [][]interface{}{
//							{execute.Time(5), execute.Time(10), execute.Time(5), 5.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(6), 6.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(7), 7.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(8), 8.0, int64(1)},
//							{execute.Time(5), execute.Time(10), execute.Time(9), 9.0, int64(1)},
//						},
//					},
//				},
//			},
//		},
//		{
//			// Error: the from node does not specify the parallel-run
//			// attribute. It is required its successor, filter.
//			name: `from-missing-output`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec([]*executetest.ParallelTable{})),
//					plantest.CreatePhysicalNode("filter",
//						&universe.FilterProcedureSpec{
//							Fn: interpreter.ResolvedFunction{
//								Scope: runtime.Prelude(),
//								Fn:    executetest.FunctionExpression(t, "(r) => r._value < 7.5"),
//							},
//						},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
//				},
//				Edges: [][2]int{
//					{0, 1},
//					{1, 2},
//					{2, 3},
//				},
//			},
//			wantValidationErr: &flux.Error{
//				Code: codes.Internal,
//				Msg: fmt.Sprintf("invalid physical query plan; attribute \"parallel-run\" " +
//					"required by \"filter\" is missing from predecessor \"parallel-from-test\""),
//			},
//		},
//		{
//			// Error: the filter node does not require the parallel-run
//			// attribute. The paralle-run attribute dictates that all
//			// successors must require it.
//			name: `from-missing-required`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec([]*executetest.ParallelTable{}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("filter",
//						&universe.FilterProcedureSpec{
//							Fn: interpreter.ResolvedFunction{
//								Scope: runtime.Prelude(),
//								Fn:    executetest.FunctionExpression(t, "(r) => r._value < 7.5"),
//							},
//						},
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
//				},
//				Edges: [][2]int{
//					{0, 1},
//					{1, 2},
//					{2, 3},
//				},
//			},
//			wantValidationErr: &flux.Error{
//				Code: codes.Internal,
//				Msg: fmt.Sprintf("invalid physical query plan; attribute \"parallel-run\" " +
//					"on \"parallel-from-test\" must be required by all successors, but isn't on \"filter\""),
//			},
//		},
//		{
//			// Error: The value of a required attribute does not match value of
//			// the output attribute in the successor.
//			name: `from-factor-mismatch`,
//			spec: &plantest.PlanSpec{
//				Nodes: []plan.Node{
//					plantest.CreatePhysicalNode("parallel-from-test",
//						executetest.NewParallelFromProcedureSpec([]*executetest.ParallelTable{}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 2})),
//					plantest.CreatePhysicalNode("filter",
//						&universe.FilterProcedureSpec{
//							Fn: interpreter.ResolvedFunction{
//								Scope: runtime.Prelude(),
//								Fn:    executetest.FunctionExpression(t, "(r) => r._value < 7.5"),
//							},
//						},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 1}),
//						plantest.WithOutputAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 1})),
//					plantest.CreatePhysicalNode("merge", &universe.PartitionMergeProcedureSpec{},
//						plantest.WithRequiredAttr(plan.ParallelRunKey, plan.ParallelRunAttribute{Factor: 1}),
//						plantest.WithOutputAttr(plan.ParallelMergeKey, plan.ParallelMergeAttribute{Factor: 1})),
//					plantest.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
//				},
//				Edges: [][2]int{
//					{0, 1},
//					{1, 2},
//					{2, 3},
//				},
//			},
//			wantValidationErr: &flux.Error{
//				Code: codes.Internal,
//				Msg: fmt.Sprintf("invalid physical query plan; attribute \"parallel-run\" " +
//					"required by \"filter\" does not match attribute in predecessor \"parallel-from-test\""),
//			},
//		},
//	}
//
//	for _, tc := range testcases {
//		tc := tc
//		t.Run(tc.name, func(t *testing.T) {
//
//			tc.spec.Resources = flux.ResourceManagement{
//				ConcurrencyQuota: 1,
//				MemoryBytesQuota: math.MaxInt64,
//			}
//
//			tc.spec.Now = time.Now()
//
//			// Construct physical query plan
//			ps := plantest.CreatePlanSpec(tc.spec)
//
//			if err := ps.TopDownWalk(plan.SetTriggerSpec); err != nil {
//				return
//			}
//
//			err := plan.ValidatePhysicalPlan(ps)
//			if tc.wantValidationErr == nil && err != nil {
//				t.Fatal(err)
//			}
//
//			if tc.wantValidationErr != nil {
//				if err == nil {
//					t.Fatalf(`expected an error "%v" but got none`, tc.wantValidationErr)
//				}
//
//				if diff := cmp.Diff(tc.wantValidationErr, err); diff != "" {
//					t.Fatalf("unexpected error: -want/+got: %v", diff)
//				}
//				return
//			}
//
//			exe := execute.NewExecutor(zaptest.NewLogger(t))
//
//			alloc := tc.allocator
//			if alloc == nil {
//				alloc = executetest.UnlimitedAllocator
//			}
//
//			// Execute the query and preserve any error returned
//			ctx, deps := dependency.Inject(context.Background(), executetest.NewTestExecuteDependencies())
//			defer deps.Finish()
//
//			results, _, err := exe.Execute(ctx, ps, alloc)
//			var got map[string][]*executetest.Table
//			if err == nil {
//				got = make(map[string][]*executetest.Table, len(results))
//				for name, r := range results {
//					if err = r.Tables().Do(func(tbl flux.Table) error {
//						cb, err := executetest.ConvertTable(tbl)
//						if err != nil {
//							return err
//						}
//						got[name] = append(got[name], cb)
//						return nil
//					}); err != nil {
//						break
//					}
//				}
//			}
//
//			if tc.wantErr == nil && err != nil {
//				t.Fatal(err)
//			}
//
//			if tc.wantErr != nil {
//				if err == nil {
//					t.Fatalf(`expected an error "%v" but got none`, tc.wantErr)
//				}
//
//				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
//					t.Fatalf("unexpected error: -want/+got: %v", diff)
//				}
//				return
//			}
//
//			for _, g := range got {
//				executetest.NormalizeTables(g)
//			}
//			for _, w := range tc.want {
//				executetest.NormalizeTables(w)
//			}
//
//			if !cmp.Equal(got, tc.want) {
//				t.Error("unexpected results -want/+got", cmp.Diff(tc.want, got))
//			}
//		})
//	}
//}
