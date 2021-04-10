package universe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestGroup_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.GroupProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
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
			wantErr: errors.New(`schema collision detected: column "_value" is both of type int and float`),
		},
		{
			name: "null values",
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
					{execute.Time(1), nil, "a"},
					{execute.Time(2), nil, "b"},
					{execute.Time(3), 1.0, "a"},
					{execute.Time(4), 2.0, "b"},
					{execute.Time(5), 3.0, "a"},
					{nil, 4.0, "b"},
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
						{execute.Time(1), nil, "a"},
						{execute.Time(3), 1.0, "a"},
						{execute.Time(5), 3.0, "a"},
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
						{execute.Time(2), nil, "b"},
						{execute.Time(4), 2.0, "b"},
						{nil, 4.0, "b"},
					},
				},
			},
		},
		{
			name: "null values in group key",
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
					{execute.Time(1), 1.0, "a"},
					{execute.Time(2), 2.0, "b"},
					{execute.Time(3), 3.0, nil},
					{execute.Time(4), 4.0, "a"},
					{execute.Time(5), 5.0, "b"},
					{execute.Time(6), 6.0, nil},
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
						{execute.Time(1), 1.0, "a"},
						{execute.Time(4), 4.0, "a"},
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
						{execute.Time(2), 2.0, "b"},
						{execute.Time(5), 5.0, "b"},
					},
				},
				{
					KeyCols:   []string{"t1"},
					KeyValues: []interface{}{nil},
					GroupKey: execute.NewGroupKey(
						[]flux.ColMeta{{Label: "t1", Type: flux.TString}},
						[]values.Value{values.NewNull(semantic.BasicString)},
					),
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0, nil},
						{execute.Time(6), 6.0, nil},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset) {
					return universe.NewGroupTransformation(tc.spec, id, alloc)
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
				Nodes: []plan.Node{
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
			plantest.LogicalRuleTestHelper(t, &tc)
		})
	}
}

func BenchmarkGroup_ByKey_1000(b *testing.B) {
	benchmarkGroupByKey(b, 1000)
}

func benchmarkGroupByKey(b *testing.B, n int) {
	spec := &universe.GroupProcedureSpec{
		GroupMode: flux.GroupModeBy,
		GroupKeys: []string{"t0"},
	}
	benchmarkGroup(b, n, spec)
}

func benchmarkGroup(b *testing.B, n int, spec *universe.GroupProcedureSpec) {
	b.ReportAllocs()
	executetest.ProcessBenchmarkHelper(b,
		func(alloc *memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: n,
				Alloc:     alloc,
				Tags: []gen.Tag{
					{Name: "_measurement", Cardinality: 1},
					{Name: "_field", Cardinality: 6},
					{Name: "t0", Cardinality: 100},
					{Name: "t1", Cardinality: 50},
				},
			}
			return gen.Input(context.Background(), schema)
		},
		func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset) {
			return universe.NewGroupTransformation(spec, id, alloc)
		},
	)
}
