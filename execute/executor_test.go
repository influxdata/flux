package execute_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/influxdata/flux/values"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/plan"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap/zaptest"
)

var epoch = time.Unix(0, 0)

func TestExecutor_Execute(t *testing.T) {
	testCases := []struct {
		name string
		plan *plan.PlanSpec
		want map[string][]*executetest.Table
	}{
		{
			name: "simple aggregate",
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: newTestFromProcedureSource(
							[]*executetest.Table{&executetest.Table{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
									{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
									{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
									{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
									{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
								},
							}},
						),
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID: plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("sum")},
				},
			},
			want: map[string][]*executetest.Table{
				plan.DefaultYieldName: []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 15.0},
					},
				}},
			},
		},
		{
			name: "simple join",
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: newTestFromProcedureSource(
							[]*executetest.Table{&executetest.Table{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TInt},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), int64(1)},
									{execute.Time(0), execute.Time(5), execute.Time(1), int64(2)},
									{execute.Time(0), execute.Time(5), execute.Time(2), int64(3)},
									{execute.Time(0), execute.Time(5), execute.Time(3), int64(4)},
									{execute.Time(0), execute.Time(5), execute.Time(4), int64(5)},
								},
							}},
						),
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID: plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID: plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID: plan.ProcedureIDFromOperationID("join"),
						Spec: &functions.MergeJoinProcedureSpec{
							TableNames: map[plan.ProcedureID]string{
								plan.ProcedureIDFromOperationID("sum"):   "sum",
								plan.ProcedureIDFromOperationID("count"): "count",
							},
							On: []string{"_start", "_stop"},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
							plan.ProcedureIDFromOperationID("count"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("join")},
				},
			},
			want: map[string][]*executetest.Table{
				plan.DefaultYieldName: []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "count__value", Type: flux.TInt},
						{Label: "sum__value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), int64(5), int64(15)},
					},
				}},
			},
		},
		{
			name: "join with multiple tables",
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: newTestFromProcedureSource(
							[]*executetest.Table{
								&executetest.Table{
									KeyCols: []string{"_start", "_stop", "_key"},
									ColMeta: []flux.ColMeta{
										{Label: "_start", Type: flux.TTime},
										{Label: "_stop", Type: flux.TTime},
										{Label: "_time", Type: flux.TTime},
										{Label: "_key", Type: flux.TString},
										{Label: "_value", Type: flux.TInt},
									},
									Data: [][]interface{}{
										{execute.Time(0), execute.Time(5), execute.Time(0), "a", int64(1)},
									},
								},
								&executetest.Table{
									KeyCols: []string{"_start", "_stop", "_key"},
									ColMeta: []flux.ColMeta{
										{Label: "_start", Type: flux.TTime},
										{Label: "_stop", Type: flux.TTime},
										{Label: "_time", Type: flux.TTime},
										{Label: "_key", Type: flux.TString},
										{Label: "_value", Type: flux.TInt},
									},
									Data: [][]interface{}{
										{execute.Time(0), execute.Time(5), execute.Time(1), "b", int64(2)},
									},
								},
								&executetest.Table{
									KeyCols: []string{"_start", "_stop", "_key"},
									ColMeta: []flux.ColMeta{
										{Label: "_start", Type: flux.TTime},
										{Label: "_stop", Type: flux.TTime},
										{Label: "_time", Type: flux.TTime},
										{Label: "_key", Type: flux.TString},
										{Label: "_value", Type: flux.TInt},
									},
									Data: [][]interface{}{
										{execute.Time(0), execute.Time(5), execute.Time(2), "c", int64(3)},
									},
								},
								&executetest.Table{
									KeyCols: []string{"_start", "_stop", "_key"},
									ColMeta: []flux.ColMeta{
										{Label: "_start", Type: flux.TTime},
										{Label: "_stop", Type: flux.TTime},
										{Label: "_time", Type: flux.TTime},
										{Label: "_key", Type: flux.TString},
										{Label: "_value", Type: flux.TInt},
									},
									Data: [][]interface{}{
										{execute.Time(0), execute.Time(5), execute.Time(3), "d", int64(4)},
									},
								},
								&executetest.Table{
									KeyCols: []string{"_start", "_stop", "_key"},
									ColMeta: []flux.ColMeta{
										{Label: "_start", Type: flux.TTime},
										{Label: "_stop", Type: flux.TTime},
										{Label: "_time", Type: flux.TTime},
										{Label: "_key", Type: flux.TString},
										{Label: "_value", Type: flux.TInt},
									},
									Data: [][]interface{}{
										{execute.Time(0), execute.Time(5), execute.Time(4), "e", int64(5)},
									},
								},
							},
						),
						Parents: nil,
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID: plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID: plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID: plan.ProcedureIDFromOperationID("join"),
						Spec: &functions.MergeJoinProcedureSpec{
							TableNames: map[plan.ProcedureID]string{
								plan.ProcedureIDFromOperationID("sum"):   "sum",
								plan.ProcedureIDFromOperationID("count"): "count",
							},
							On: []string{"_start", "_stop", "_key"},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
							plan.ProcedureIDFromOperationID("count"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("join")},
				},
			},
			want: map[string][]*executetest.Table{
				plan.DefaultYieldName: []*executetest.Table{
					{
						KeyCols: []string{"_key", "_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_key", Type: flux.TString},
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "count__value", Type: flux.TInt},
							{Label: "sum__value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{"a", execute.Time(0), execute.Time(5), int64(1), int64(1)},
						},
					},
					{
						KeyCols: []string{"_key", "_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_key", Type: flux.TString},
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "count__value", Type: flux.TInt},
							{Label: "sum__value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{"b", execute.Time(0), execute.Time(5), int64(1), int64(2)},
						},
					},
					{
						KeyCols: []string{"_key", "_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_key", Type: flux.TString},
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "count__value", Type: flux.TInt},
							{Label: "sum__value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{"c", execute.Time(0), execute.Time(5), int64(1), int64(3)},
						},
					},
					{
						KeyCols: []string{"_key", "_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_key", Type: flux.TString},
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "count__value", Type: flux.TInt},
							{Label: "sum__value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{"d", execute.Time(0), execute.Time(5), int64(1), int64(4)},
						},
					},
					{
						KeyCols: []string{"_key", "_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_key", Type: flux.TString},
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "count__value", Type: flux.TInt},
							{Label: "sum__value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{"e", execute.Time(0), execute.Time(5), int64(1), int64(5)},
						},
					},
				},
			},
		},
		{
			name: "multiple aggregates",
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: newTestFromProcedureSource(
							[]*executetest.Table{&executetest.Table{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
									{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
									{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
									{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
									{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
								},
							}},
						),
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
							plan.ProcedureIDFromOperationID("mean"),
						},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID: plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: nil,
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID: plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(time.Unix(0, 1)),
							Stop:  values.ConvertTime(time.Unix(0, 5)),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					"sum":  {ID: plan.ProcedureIDFromOperationID("sum")},
					"mean": {ID: plan.ProcedureIDFromOperationID("mean")},
				},
			},
			want: map[string][]*executetest.Table{
				"sum": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 15.0},
					},
				}},
				"mean": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 3.0},
					},
				}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			exe := execute.NewExecutor(nil, zaptest.NewLogger(t))
			results, err := exe.Execute(context.Background(), tc.plan, executetest.UnlimitedAllocator)
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string][]*executetest.Table, len(results))
			for name, r := range results {
				if err := r.Tables().Do(func(tbl flux.Table) error {
					cb, err := executetest.ConvertTable(tbl)
					if err != nil {
						return err
					}
					got[name] = append(got[name], cb)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}

			for _, g := range got {
				executetest.NormalizeTables(g)
			}
			for _, w := range tc.want {
				executetest.NormalizeTables(w)
			}

			if !cmp.Equal(got, tc.want) {
				t.Error("unexpected results -want/+got", cmp.Diff(tc.want, got))
			}
		})
	}
}

type testFromProcedureSource struct {
	data []*executetest.Table
	ts   []execute.Transformation
}

func newTestFromProcedureSource(data []*executetest.Table) *testFromProcedureSource {
	p := &testFromProcedureSource{
		data: data,
	}
	// Normalize the data before anything can read it
	for _, tbl := range p.data {
		tbl.Normalize()
	}
	return p
}

func (p *testFromProcedureSource) Kind() plan.ProcedureKind {
	return "from-test"
}

func (p *testFromProcedureSource) Copy() plan.ProcedureSpec {
	return p
}

func (p *testFromProcedureSource) AddTransformation(t execute.Transformation) {
	p.ts = append(p.ts, t)
}

func (p *testFromProcedureSource) Run(ctx context.Context) {
	id := execute.DatasetID(uuid.NewV4())
	for _, t := range p.ts {
		var max execute.Time
		for _, tbl := range p.data {
			t.Process(id, tbl)
			stopIdx := execute.ColIdx(execute.DefaultStopColLabel, tbl.Cols())
			if stopIdx >= 0 {
				if s := tbl.Key().ValueTime(stopIdx); s > max {
					max = s
				}
			}
		}
		t.UpdateWatermark(id, max)
		t.Finish(id, nil)
	}
}

func init() {
	execute.RegisterSource("from-test", createTestFromSource)
}

func createTestFromSource(prSpec plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	return prSpec.(*testFromProcedureSource), nil
}
