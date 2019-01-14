package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestGroupOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"group","kind":"group","spec":{"mode":"by","columns":["t1","t2"]}}`)
	op := &flux.Operation{
		ID: "group",
		Spec: &universe.GroupOpSpec{
			Mode:    "by",
			Columns: []string{"t1", "t2"},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestGroup_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "group with no arguments",
			// group() defaults to group(columns: [], mode: "by")
			Raw: `from(bucket: "telegraf") |> range(start: -1m) |> group()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "from0",
						Spec: &influxdb.FromOpSpec{Bucket: "telegraf"},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Minute,
								IsRelative: true,
							},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID:   "group2",
						Spec: &universe.GroupOpSpec{Mode: "by", Columns: []string{}},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "group2"},
				},
			},
		},
		{
			Name: "group all",
			Raw:  `from(bucket: "telegraf") |> range(start: -1m) |> group(columns:[], mode: "except")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "from0",
						Spec: &influxdb.FromOpSpec{Bucket: "telegraf"},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Minute,
								IsRelative: true,
							},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID:   "group2",
						Spec: &universe.GroupOpSpec{Mode: "except", Columns: []string{}},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "group2"},
				},
			},
		},
		{
			Name: "group none",
			Raw:  `from(bucket: "telegraf") |> range(start: -1m) |> group(columns: [], mode: "by")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "from0",
						Spec: &influxdb.FromOpSpec{Bucket: "telegraf"},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Minute,
								IsRelative: true,
							},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID:   "group2",
						Spec: &universe.GroupOpSpec{Mode: "by", Columns: []string{}},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "group2"},
				},
			},
		},
		{
			Name: "group by",
			Raw:  `from(bucket: "telegraf") |> range(start: -1m) |> group(columns: ["host"], mode: "by")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "from0",
						Spec: &influxdb.FromOpSpec{Bucket: "telegraf"},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Minute,
								IsRelative: true,
							},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "group2",
						Spec: &universe.GroupOpSpec{
							Columns: []string{"host"},
							Mode:    "by",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "group2"},
				},
			},
		},
		{
			Name: "group except",
			Raw:  `from(bucket: "telegraf") |> range(start: -1m) |> group(columns: ["host"], mode: "except")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "from0",
						Spec: &influxdb.FromOpSpec{Bucket: "telegraf"},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Minute,
								IsRelative: true,
							},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "group2",
						Spec: &universe.GroupOpSpec{
							Columns: []string{"host"},
							Mode:    "except",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "group2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestGroup_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.GroupProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "fan in",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeBy,
				GroupKeys: []string{"t1"},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "y"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 4.0, "b", "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 7.0, "b", "y"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "x"},
						{execute.Time(2), 1.0, "a", "y"},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 4.0, "b", "x"},
						{execute.Time(2), 7.0, "b", "y"},
					},
				},
			},
		},
		{
			name: "fan in ignoring",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeExcept,
				GroupKeys: []string{"_time", "_value", "t2"},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1", "t2", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "m", "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "n", "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 4.0, "b", "m", "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 7.0, "b", "n", "x"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "m", "x"},
						{execute.Time(2), 1.0, "a", "n", "x"},
					},
				},
				{
					KeyCols: []string{"t1", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 4.0, "b", "m", "x"},
						{execute.Time(2), 7.0, "b", "n", "x"},
					},
				},
			},
		},
		{
			name: "fan in missing columns",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeBy,
				GroupKeys: []string{"t1"},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "y"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t3", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", int64(5), "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 7.0, "b", "y"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t3", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 4.0, "b", int64(7), "x"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "y", nil},
						{execute.Time(1), 2.0, "a", "x", int64(5)},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(2), 7.0, "b", "y", nil},
						{execute.Time(1), 4.0, "b", "x", int64(7)},
					},
				},
			},
		},
		{
			name: "fan out",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeBy,
				GroupKeys: []string{"t1"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, "a"},
					{execute.Time(2), 1.0, "b"},
				},
			}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a"},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "b"},
					},
				},
			},
		},
		{
			name: "fan out ignoring",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeExcept,
				GroupKeys: []string{"_time", "_value", "t2"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
					{Label: "t3", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, "a", "m", "x"},
					{execute.Time(2), 1.0, "a", "n", "y"},
				},
			}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "m", "x"},
					},
				},
				{
					KeyCols: []string{"t1", "t3"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
						{Label: "t3", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "n", "y"},
					},
				},
			},
		},
		{
			name: "heterogeneous typed columns",
			spec: &universe.GroupProcedureSpec{
				GroupMode: flux.GroupModeBy,
				GroupKeys: []string{"t1"},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 1.0, "a", "y"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t3", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "t1", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), "a", int64(5), "x"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 7.0, "b", "y"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1", "t3", "t2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "t1", Type: flux.TString},
						{Label: "t3", Type: flux.TInt},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(4), "b", int64(7), "x"},
					},
				},
			},
			want: []*executetest.Table{}, // TODO What do we want?
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "heterogeneous typed columns" {
				t.Skip("should pass once we decide the expected behavior: https://github.com/influxdata/flux/issues/439")
			}

			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewGroupTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestMergeGroupRule(t *testing.T) {
	var (
		from      = &influxdb.FromProcedureSpec{}
		groupNone = &universe.GroupProcedureSpec{
			GroupMode: flux.GroupModeBy,
			GroupKeys: []string{},
		}
		groupBy = &universe.GroupProcedureSpec{
			GroupMode: flux.GroupModeBy,
			GroupKeys: []string{"foo", "bar", "buz"},
		}
		groupExcept = &universe.GroupProcedureSpec{
			GroupMode: flux.GroupModeExcept,
			GroupKeys: []string{"foo", "bar", "buz"},
		}
		groupNotByNorExcept = &universe.GroupProcedureSpec{
			GroupMode: flux.GroupModeNone,
			GroupKeys: []string{},
		}
		filter = &universe.FilterProcedureSpec{}
	)

	tests := []plantest.RuleTestCase{
		{
			Name:  "single group",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group", groupBy),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			NoChange: true,
		},
		{
			Name:  "double group",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupNone),
					plan.CreateLogicalNode("group1", groupBy),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("merged_group0_group1", groupBy),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			Name:  "triple group",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupNone),
					plan.CreateLogicalNode("group1", groupBy),
					plan.CreateLogicalNode("group2", groupExcept),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("merged_group0_group1_group2", groupExcept),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			Name:  "double group not by nor except",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupNone),
					plan.CreateLogicalNode("group1", groupNotByNorExcept),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			NoChange: true,
		},
		{
			// the last group by/except always overrides the group key
			Name:  "triple group not by nor except",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupNone),
					plan.CreateLogicalNode("group1", groupNotByNorExcept),
					plan.CreateLogicalNode("group2", groupExcept),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("merged_group0_group1_group2", groupExcept),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			Name:  "quad group not by nor except",
			Rules: []plan.Rule{&universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupNone),
					plan.CreateLogicalNode("group1", groupNotByNorExcept),
					plan.CreateLogicalNode("group2", groupExcept),
					plan.CreateLogicalNode("group3", groupNotByNorExcept),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("merged_group0_group1_group2", groupExcept),
					plan.CreateLogicalNode("group3", groupNotByNorExcept),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
		{
			Name:  "from group group filter",
			Rules: []plan.Rule{universe.MergeGroupRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("group0", groupExcept),
					plan.CreateLogicalNode("group1", groupBy),
					plan.CreateLogicalNode("filter", filter),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from", from),
					plan.CreateLogicalNode("merged_group0_group1", groupBy),
					plan.CreateLogicalNode("filter", filter),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			plantest.RuleTestHelper(t, &tc)
		})
	}
}
