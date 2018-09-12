package plan_test

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/flux/values"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
)

func TestPhysicalPlanner_Plan(t *testing.T) {
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	testCases := []struct {
		name string
		lp   *plan.LogicalPlanSpec
		pp   *plan.PlanSpec
	}{
		{
			name: "single push down",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("count")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("count"),
				},
			},
		},
		{
			name: "single push down with match",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("last")},
					},
					plan.ProcedureIDFromOperationID("last"): {
						ID:   plan.ProcedureIDFromOperationID("last"),
						Spec: &functions.LastProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("last"),
				},
			},
			pp: &plan.PlanSpec{
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.MinTime,
								Stop:  flux.Now,
							},
							LimitSet:      true,
							PointsLimit:   1,
							DescendingSet: true,
							Descending:    true,
						},
						Bounds: &plan.BoundsSpec{
							Start: plan.MinTime,
							Stop:  values.ConvertTime(now),
						},
						Parents:  nil,
						Children: []plan.ProcedureID{},
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("from")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
			},
		},
		{
			name: "multiple push down",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					plan.ProcedureIDFromOperationID("limit"): {
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &functions.LimitProcedureSpec{
							N: 10,
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("limit"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 2,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							LimitSet:    true,
							PointsLimit: 10,
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("from")),
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("mean")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
		},
		{
			name: "multiple yield",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("stddev"),
							plan.ProcedureIDFromOperationID("skew"),
						},
					},
					plan.ProcedureIDFromOperationID("stddev"): {
						ID:       plan.ProcedureIDFromOperationID("stddev"),
						Spec:     &functions.StddevProcedureSpec{},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("yieldStddev")},
					},
					plan.ProcedureIDFromOperationID("yieldStddev"): {
						ID:       plan.ProcedureIDFromOperationID("yieldStddev"),
						Spec:     &functions.YieldProcedureSpec{Name: "stddev"},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("stddev")},
						Children: nil,
					},
					plan.ProcedureIDFromOperationID("skew"): {
						ID:       plan.ProcedureIDFromOperationID("skew"),
						Spec:     &functions.SkewProcedureSpec{},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("yieldSkew")},
					},
					plan.ProcedureIDFromOperationID("yieldSkew"): {
						ID:       plan.ProcedureIDFromOperationID("yieldSkew"),
						Spec:     &functions.YieldProcedureSpec{Name: "skew"},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("skew")},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("stddev"),
					plan.ProcedureIDFromOperationID("yieldStddev"),
					plan.ProcedureIDFromOperationID("skew"),
					plan.ProcedureIDFromOperationID("yieldSkew"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("stddev"),
							plan.ProcedureIDFromOperationID("skew"),
						},
					},
					plan.ProcedureIDFromOperationID("stddev"): {
						ID:   plan.ProcedureIDFromOperationID("stddev"),
						Spec: &functions.StddevProcedureSpec{},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{},
					},
					plan.ProcedureIDFromOperationID("skew"): {
						ID:   plan.ProcedureIDFromOperationID("skew"),
						Spec: &functions.SkewProcedureSpec{},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{},
					},
				},
				Results: map[string]plan.YieldSpec{
					"stddev": {ID: plan.ProcedureIDFromOperationID("stddev")},
					"skew":   {ID: plan.ProcedureIDFromOperationID("skew")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("stddev"),
					plan.ProcedureIDFromOperationID("skew"),
				},
			},
		},
		{
			name: "group with aggregate",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("group"),
						},
					},
					plan.ProcedureIDFromOperationID("group"): {
						ID: plan.ProcedureIDFromOperationID("group"),
						Spec: &functions.GroupProcedureSpec{
							GroupMode: functions.GroupModeBy,
							GroupKeys: []string{"host", "region"},
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:      plan.ProcedureIDFromOperationID("sum"),
						Spec:    &functions.SumProcedureSpec{},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("group")},
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("group"),
					plan.ProcedureIDFromOperationID("sum"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							GroupingSet:     true,
							GroupMode:       functions.GroupModeBy,
							GroupKeys:       []string{"host", "region"},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
						},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
					},
				},
				Results: map[string]plan.YieldSpec{
					"_result": {ID: plan.ProcedureIDFromOperationID("sum")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("sum"),
				},
			},
		},
		{
			name: "group with distinct on tag",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("group"),
						},
					},
					plan.ProcedureIDFromOperationID("group"): {
						ID: plan.ProcedureIDFromOperationID("group"),
						Spec: &functions.GroupProcedureSpec{
							GroupMode: functions.GroupModeBy,
							GroupKeys: []string{"host"},
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("distinct")},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID: plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{
							Column: "host",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("group")},
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("group"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							GroupingSet: true,
							GroupMode:   functions.GroupModeBy,
							GroupKeys:   []string{"host"},
							LimitSet:    true,
							PointsLimit: -1,
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("distinct"),
						},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID:   plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{Column: "host"},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
					},
				},
				Results: map[string]plan.YieldSpec{
					"_result": {ID: plan.ProcedureIDFromOperationID("distinct")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
		},
		{
			name: "group with distinct on _value does not optimize",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("group"),
						},
					},
					plan.ProcedureIDFromOperationID("group"): {
						ID: plan.ProcedureIDFromOperationID("group"),
						Spec: &functions.GroupProcedureSpec{
							GroupMode: functions.GroupModeBy,
							GroupKeys: []string{"host"},
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("distinct")},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID: plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{
							Column: "_value",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("group")},
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("group"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							GroupingSet: true,
							GroupMode:   functions.GroupModeBy,
							GroupKeys:   []string{"host"},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("distinct"),
						},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID:   plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{Column: "_value"},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
					},
				},
				Results: map[string]plan.YieldSpec{
					"_result": {ID: plan.ProcedureIDFromOperationID("distinct")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
		},
		{
			name: "group with distinct on non-grouped does not optimize",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
					},
					plan.ProcedureIDFromOperationID("range"): {
						ID: plan.ProcedureIDFromOperationID("range"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("group"),
						},
					},
					plan.ProcedureIDFromOperationID("group"): {
						ID: plan.ProcedureIDFromOperationID("group"),
						Spec: &functions.GroupProcedureSpec{
							GroupMode: functions.GroupModeBy,
							GroupKeys: []string{"host"},
						},
						Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("distinct")},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID: plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{
							Column: "region",
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("group")},
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("range"),
					plan.ProcedureIDFromOperationID("group"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
			pp: &plan.PlanSpec{
				Now: time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC),
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: 10000,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Bucket:    "mybucket",
							BoundsSet: true,
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							GroupingSet: true,
							GroupMode:   functions.GroupModeBy,
							GroupKeys:   []string{"host"},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: nil,
						Children: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("distinct"),
						},
					},
					plan.ProcedureIDFromOperationID("distinct"): {
						ID:   plan.ProcedureIDFromOperationID("distinct"),
						Spec: &functions.DistinctProcedureSpec{Column: "region"},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
					},
				},
				Results: map[string]plan.YieldSpec{
					"_result": {ID: plan.ProcedureIDFromOperationID("distinct")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
					plan.ProcedureIDFromOperationID("distinct"),
				},
			},
		},
		{
			name: "bounds context",
			lp: &plan.LogicalPlanSpec{
				Now: now,
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("fromCSV"): {
						ID: plan.ProcedureIDFromOperationID("fromCSV"),
						Spec: &functions.FromCSVProcedureSpec{
							File: "file",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range1")},
					},
					plan.ProcedureIDFromOperationID("range1"): {
						ID: plan.ProcedureIDFromOperationID("range1"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("fromCSV")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range2")},
					},
					plan.ProcedureIDFromOperationID("range2"): {
						ID: plan.ProcedureIDFromOperationID("range2"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -30 * time.Minute,
								},
								Stop: flux.Now,
							},
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range1")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					plan.ProcedureIDFromOperationID("limit"): {
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &functions.LimitProcedureSpec{
							N: 10,
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range2"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("fromCSV"),
					plan.ProcedureIDFromOperationID("range1"),
					plan.ProcedureIDFromOperationID("range2"),
					plan.ProcedureIDFromOperationID("limit"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
			pp: &plan.PlanSpec{
				Now: now,
				Resources: flux.ResourceManagement{
					ConcurrencyQuota: 5,
					MemoryBytesQuota: math.MaxInt64,
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("fromCSV"): {
						ID: plan.ProcedureIDFromOperationID("fromCSV"),
						Spec: &functions.FromCSVProcedureSpec{
							File: "file",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range1")},
					},
					plan.ProcedureIDFromOperationID("range1"): {
						ID: plan.ProcedureIDFromOperationID("range1"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -1 * time.Hour,
								},
								Stop: flux.Now,
							},
							TimeCol: "_time",
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-1 * time.Hour)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("fromCSV")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range2")},
					},
					plan.ProcedureIDFromOperationID("range2"): {
						ID: plan.ProcedureIDFromOperationID("range2"),
						Spec: &functions.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{
									IsRelative: true,
									Relative:   -30 * time.Minute,
								},
								Stop: flux.Now,
							},
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-30 * time.Minute)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("range1")),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("limit")},
					},
					plan.ProcedureIDFromOperationID("limit"): {
						ID: plan.ProcedureIDFromOperationID("limit"),
						Spec: &functions.LimitProcedureSpec{
							N: 10,
						},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-30 * time.Minute)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range2"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("mean")},
					},
					plan.ProcedureIDFromOperationID("mean"): {
						ID:   plan.ProcedureIDFromOperationID("mean"),
						Spec: &functions.MeanProcedureSpec{},
						Bounds: &plan.BoundsSpec{
							Start: values.ConvertTime(now.Add(-30 * time.Minute)),
							Stop:  values.ConvertTime(now),
						},
						Parents: []plan.ProcedureID{
							(plan.ProcedureIDFromOperationID("limit")),
						},
						Children: nil,
					},
				},
				Results: map[string]plan.YieldSpec{
					plan.DefaultYieldName: {ID: plan.ProcedureIDFromOperationID("mean")},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("fromCSV"),
					plan.ProcedureIDFromOperationID("range1"),
					plan.ProcedureIDFromOperationID("range2"),
					plan.ProcedureIDFromOperationID("limit"),
					plan.ProcedureIDFromOperationID("mean"),
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			PhysicalPlanTestHelper(t, tc.lp, tc.pp)
		})
	}
}

func TestPhysicalPlanner_Plan_PushDown_Branch(t *testing.T) {
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	lp := &plan.LogicalPlanSpec{
		Now: now,
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Bucket: "mybucket",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("first"),
					plan.ProcedureIDFromOperationID("last"),
				},
			},
			plan.ProcedureIDFromOperationID("first"): {
				ID:       plan.ProcedureIDFromOperationID("first"),
				Spec:     &functions.FirstProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("yieldFirst")},
			},
			plan.ProcedureIDFromOperationID("yieldFirst"): {
				ID:       plan.ProcedureIDFromOperationID("yieldFirst"),
				Spec:     &functions.YieldProcedureSpec{Name: "first"},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("first")},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("last"): {
				ID:       plan.ProcedureIDFromOperationID("last"),
				Spec:     &functions.LastProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("yieldLast")},
			},
			plan.ProcedureIDFromOperationID("yieldLast"): {
				ID:       plan.ProcedureIDFromOperationID("yieldLast"),
				Spec:     &functions.YieldProcedureSpec{Name: "last"},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("last")},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("first"),
			plan.ProcedureIDFromOperationID("yieldFirst"),
			plan.ProcedureIDFromOperationID("last"), // last is last so it will be duplicated
			plan.ProcedureIDFromOperationID("yieldLast"),
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Now: now,
		Resources: flux.ResourceManagement{
			ConcurrencyQuota: 2,
			MemoryBytesQuota: math.MaxInt64,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Bucket:    "mybucket",
					BoundsSet: true,
					Bounds: flux.Bounds{
						Start: flux.MinTime,
						Stop:  flux.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    true, // last
				},
				Bounds: &plan.BoundsSpec{
					Start: plan.MinTime,
					Stop:  values.ConvertTime(now),
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Bucket:    "mybucket",
					BoundsSet: true,
					Bounds: flux.Bounds{
						Start: flux.MinTime,
						Stop:  flux.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    false, // first
				},
				Bounds: &plan.BoundsSpec{
					Start: plan.MinTime,
					Stop:  values.ConvertTime(now),
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: map[string]plan.YieldSpec{
			"first": {ID: fromIDDup},
			"last":  {ID: fromID},
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
	}

	PhysicalPlanTestHelper(t, lp, want)
}

func PhysicalPlanTestHelper(t *testing.T, lp *plan.LogicalPlanSpec, want *plan.PlanSpec) {
	t.Helper()

	// Setup expected now time if it doesn't exist
	now := time.Now()
	if lp.Now.IsZero() {
		lp.Now = now
	}

	if want.Now.IsZero() {
		want.Now = now
	}

	planner := plan.NewPlanner()
	got, err := planner.Plan(lp, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(got, want, plantest.CmpOptions...) {
		t.Log("Logical:", plan.Formatted(lp))
		t.Log("Want Physical:", plan.Formatted(want))
		t.Log("Got  Physical:", plan.Formatted(got))
		t.Errorf("unexpected physical plan -want/+got:\n%s", cmp.Diff(want, got, plantest.CmpOptions...))
	}
}

var benchmarkPhysicalPlan *plan.PlanSpec

func BenchmarkPhysicalPlan(b *testing.B) {
	var err error
	lp, err := plan.NewLogicalPlanner().Plan(benchmarkQuery)
	if err != nil {
		b.Fatal(err)
	}
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	lp.Now = now
	planner := plan.NewPlanner()
	for n := 0; n < b.N; n++ {
		benchmarkPhysicalPlan, err = planner.Plan(lp, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var benchmarkQueryToPhysicalPlan *plan.PlanSpec

func BenchmarkQueryToPhysicalPlan(b *testing.B) {
	lp := plan.NewLogicalPlanner()
	pp := plan.NewPlanner()
	now := time.Date(2017, 8, 8, 0, 0, 0, 0, time.UTC)
	for n := 0; n < b.N; n++ {
		lp, err := lp.Plan(benchmarkQuery)
		if err != nil {
			b.Fatal(err)
		}
		lp.Now = now
		benchmarkQueryToPhysicalPlan, err = pp.Plan(lp, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
