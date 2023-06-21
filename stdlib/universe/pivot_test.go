package universe_test

import (
	"context"
	"testing"
	"time"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/internal/gen"
	"github.com/InfluxCommunity/flux/internal/operation"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/querytest"
	"github.com/InfluxCommunity/flux/stdlib/influxdata/influxdb"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func TestPivot_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "pivot [_measurement, _field] around _time",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> pivot(rowKey: ["_time"], columnKey: ["_measurement", "_field"], valueColumn: "_value")`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "testdb"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "pivot2",
						Spec: &universe.PivotOpSpec{
							RowKey:      []string{"_time"},
							ColumnKey:   []string{"_measurement", "_field"},
							ValueColumn: "_value",
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "pivot2"},
				},
			},
		},
		{
			Name:    "overlapping rowKey and columnKey",
			Raw:     `from(bucket:"testdb") |> range(start: -1h) |> pivot(rowKey: ["_time", "a"], columnKey: ["_measurement", "_field", "a"], valueColumn: "_value")`,
			WantErr: true,
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

func TestPivot_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.PivotProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "_field flatten case one table",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "_field flatten case two tables",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m2", "f3"},
						{execute.Time(1), 2.0, "m2", "f4"},
						{execute.Time(2), 3.0, "m2", "f3"},
						{execute.Time(2), 4.0, "m2", "f4"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f3", Type: flux.TFloat},
						{Label: "f4", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m2", 1.0, 2.0},
						{execute.Time(2), "m2", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "_field flatten case two tables different value type",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(1), "m2", "f3"},
						{execute.Time(1), int64(2), "m2", "f4"},
						{execute.Time(2), int64(3), "m2", "f3"},
						{execute.Time(2), int64(4), "m2", "f4"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f3", Type: flux.TInt},
						{Label: "f4", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m2", int64(1), int64(2)},
						{execute.Time(2), "m2", int64(3), int64(4)},
					},
				},
			},
		},
		{
			name: "duplicate rowKey + columnKey",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
						{execute.Time(1), 5.0, "m1", "f1"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: nil,
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "m1_f1", Type: flux.TFloat},
						{Label: "m1_f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 5.0, 2.0},
						{execute.Time(2), 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "dropping a column not in rowKey or groupKey",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "droppedcol", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", int64(1)},
						{execute.Time(1), 2.0, "m1", "f2", int64(1)},
						{execute.Time(2), 3.0, "m1", "f1", int64(1)},
						{execute.Time(2), 4.0, "m1", "f2", int64(1)},
						{execute.Time(1), 5.0, "m1", "f1", int64(1)},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: nil,
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "m1_f1", Type: flux.TFloat},
						{Label: "m1_f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 5.0, 2.0},
						{execute.Time(2), 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "group key doesn't change",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(1), 2.0, "m1", "f2", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
						{execute.Time(2), 4.0, "m1", "f2", "A"},
						{execute.Time(1), 5.0, "m1", "f1", "A"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "grouper", Type: flux.TString},
						{Label: "m1_f1", Type: flux.TFloat},
						{Label: "m1_f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A", 5.0, 2.0},
						{execute.Time(2), "A", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "group key loses a member",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", "B"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 4.0, "m1", "f2", "A"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "grouper", Type: flux.TString},
						{Label: "m1_f1", Type: flux.TFloat},
						{Label: "m1_f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A", 1.0, nil},
						{execute.Time(2), "A", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "grouper", Type: flux.TString},
						{Label: "m1_f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "B", 2.0},
					},
				},
			},
		},
		{
			name: "group key loses all members. drops _value",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "grouper",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", "B"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 4.0, "m1", "f2", "A"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: nil,
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "m1_f1", Type: flux.TString},
						{Label: "m1_f2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A", "B"},
						{execute.Time(2), "A", "A"},
					},
				},
			},
		},
		{
			name: "_field flatten case one table with null ColumnKey",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
						{execute.Time(3), 5.0, "m1", nil},
						{execute.Time(3), 6.0, "m1", "f2"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
						{Label: "null", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0, nil},
						{execute.Time(2), "m1", 3.0, 4.0, nil},
						{execute.Time(3), "m1", nil, 6.0, 5.0},
					},
				},
			},
		},
		{
			name: "_field flatten case one table with null RowKey",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
						{nil, 5.0, "m1", "f1"},
						{nil, 6.0, "m1", "f2"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
						{nil, "m1", 5.0, 6.0},
					},
				},
			},
		},
		{
			name: "_field flatten case one table with nulls",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(1), nil, "m1", "f3"},
						{execute.Time(1), 3.0, "m1", nil},

						{execute.Time(2), 4.0, "m1", "f1"},
						{execute.Time(2), 5.0, "m1", "f2"},
						{nil, 6.0, "m1", "f2"},
						{execute.Time(2), nil, "m1", "f3"},

						{execute.Time(3), nil, "m1", "f1"},
						{execute.Time(3), 7.0, "m1", nil},

						{execute.Time(4), 8.0, "m1", "f3"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
						{Label: "f3", Type: flux.TFloat},
						{Label: "null", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0, nil, 3.0},
						{execute.Time(2), "m1", 4.0, 5.0, nil, nil},
						{nil, "m1", nil, 6.0, nil, nil},
						{execute.Time(3), "m1", nil, nil, nil, 7.0},
						{execute.Time(4), "m1", nil, nil, 8.0, nil},
					},
				},
			},
		},
		{
			name: "two ColumnKeys with nulls and duplicate value",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_measurement", "_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(1), 3.0, nil, "f3"},
						{execute.Time(1), 4.0, nil, nil},

						{execute.Time(2), 5.0, "m1", "f1"},
						{execute.Time(2), 6.0, "m1", "f2"},
						{execute.Time(2), 7.0, "m1", "f3"},
						{execute.Time(2), 8.0, nil, nil},
						{nil, 9.0, "m1", "f3"},

						{execute.Time(3), 10.0, "m1", nil},
						{execute.Time(3), 11.0, "m1", nil},
						{execute.Time(3), 12.0, "m1", "f3"},
						{execute.Time(3), 13.0, nil, nil},
						{nil, 14.0, "m1", nil},
						{nil, 15.0, "m1", nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: nil,
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "m1_f1", Type: flux.TFloat},
						{Label: "m1_f2", Type: flux.TFloat},
						{Label: "null_f3", Type: flux.TFloat},
						{Label: "null_null", Type: flux.TFloat},
						{Label: "m1_f3", Type: flux.TFloat},
						{Label: "m1_null", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, 2.0, 3.0, 4.0, nil, nil},
						{execute.Time(2), 5.0, 6.0, nil, 8.0, 7.0, nil},
						{nil, nil, nil, nil, nil, 9.0, 15.0},
						{execute.Time(3), nil, nil, nil, 13.0, 12.0, 11.0},
					},
				},
			},
		},
		{
			name: "missing value column",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_new_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 3.0, "m1", "f1"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
			},
			wantErr: errors.New(codes.Invalid, "specified value column does not exist in table: _value"),
		},
		{
			name: "column name conflict",
			spec: &universe.PivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "f1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "f1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "foo"},
						{execute.Time(1), 2.0, "m1", "f2", "foo"},
						{execute.Time(2), 3.0, "m1", "f1", "foo"},
						{execute.Time(2), 4.0, "m1", "f2", "foo"},
					},
				},
			},
			wantErr: errors.New(
				codes.Invalid,
				`value "f1" appears in a column key column, but a column named "f1" already exists; consider renaming "f1" to something else before pivoting`,
			),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewPivotTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestSortedPivot_ProcessWithTags(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.SortedPivotProcedureSpec
		data []flux.Table // test case data must be in groupKey sorted order
		want []*executetest.Table
	}{
		{
			name: "_field and tag with one measurement",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.83, "system", "load1", "host.local"},
						{execute.Time(2), 1.7, "system", "load1", "host.local"},
						{execute.Time(3), 1.74, "system", "load1", "host.local"},
						{execute.Time(4), 1.63, "system", "load1", "host.local"},
						{execute.Time(5), 1.91, "system", "load1", "host.local"},
						{execute.Time(6), 1.84, "system", "load1", "host.local"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.98, "system", "load15", "host.local"},
						{execute.Time(2), 1.97, "system", "load15", "host.local"},
						{execute.Time(3), 1.97, "system", "load15", "host.local"},
						{execute.Time(4), 1.96, "system", "load15", "host.local"},
						{execute.Time(5), 1.98, "system", "load15", "host.local"},
						{execute.Time(6), 1.97, "system", "load15", "host.local"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(95), "system", "load5", "host.local"},
						{execute.Time(2), int64(92), "system", "load5", "host.local"},
						{execute.Time(3), int64(92), "system", "load5", "host.local"},
						{execute.Time(4), int64(89), "system", "load5", "host.local"},
						{execute.Time(5), int64(94), "system", "load5", "host.local"},
						{execute.Time(6), int64(93), "system", "load5", "host.local"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(82), "swap", "used_p", "host.local"},
						{execute.Time(2), int64(83), "swap", "used_p", "host.local"},
						{execute.Time(3), int64(84), "swap", "used_p", "host.local"},
						{execute.Time(4), int64(85), "swap", "used_p", "host.local"},
						{execute.Time(5), int64(82), "swap", "used_p", "host.local"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "used_p", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), "swap", "host.local", int64(82)},
						{execute.Time(2), "swap", "host.local", int64(83)},
						{execute.Time(3), "swap", "host.local", int64(84)},
						{execute.Time(4), "swap", "host.local", int64(85)},
						{execute.Time(5), "swap", "host.local", int64(82)},
					},
				},
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "load1", Type: flux.TFloat},
						{Label: "load15", Type: flux.TFloat},
						{Label: "load5", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), "system", "host.local", 1.83, 1.98, int64(95)},
						{execute.Time(2), "system", "host.local", 1.7, 1.97, int64(92)},
						{execute.Time(3), "system", "host.local", 1.74, 1.97, int64(92)},
						{execute.Time(4), "system", "host.local", 1.63, 1.96, int64(89)},
						{execute.Time(5), "system", "host.local", 1.91, 1.98, int64(94)},
						{execute.Time(6), "system", "host.local", 1.84, 1.97, int64(93)},
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
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					spec := *tc.spec
					tr, d, err := universe.NewSortedPivotTransformation(context.Background(), spec, id, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tr, d
				})
		})
	}
}

func TestSortedPivot_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.SortedPivotProcedureSpec
		data []flux.Table // test case data must be in groupKey sorted order
		want []*executetest.Table
	}{
		{
			name: "_field flatten case one measurement",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(2), 3.0, "m1", "f1"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "_field flatten case two measurements",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(2), 3.0, "m1", "f1"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m2", "f3"},
						{execute.Time(2), 3.0, "m2", "f3"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m2", "f4"},
						{execute.Time(2), 4.0, "m2", "f4"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f3", Type: flux.TFloat},
						{Label: "f4", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m2", 1.0, 2.0},
						{execute.Time(2), "m2", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "_field flatten case two measurements different value type",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1"},
						{execute.Time(2), 3.0, "m1", "f1"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 4.0, "m1", "f2"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(1), "m2", "f3"},
						{execute.Time(2), int64(3), "m2", "f3"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), "m2", "f4"},
						{execute.Time(2), int64(4), "m2", "f4"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f3", Type: flux.TInt},
						{Label: "f4", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m2", int64(1), int64(2)},
						{execute.Time(2), "m2", int64(3), int64(4)},
					},
				},
			},
		},
		{
			name: "dropping a column not in rowKey or groupKey",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "droppedcol", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", int64(1)},
						{execute.Time(2), 3.0, "m1", "f1", int64(1)},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "droppedcol", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", int64(1)},
						{execute.Time(2), 4.0, "m1", "f2", int64(1)},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0},
						{execute.Time(2), "m1", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "group key doesn't change",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field", "grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", "A"},
						{execute.Time(2), 4.0, "m1", "f2", "A"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", "A", 1.0, 2.0},
						{execute.Time(2), "m1", "A", 3.0, 4.0},
					},
				},
			},
		},
		{
			name: "group key loses a member",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 4.0, "m1", "f2", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", "B"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", "A", 1.0, nil},
						{execute.Time(2), "m1", "A", 3.0, 4.0},
					},
				},
				{
					KeyCols: []string{"_measurement", "grouper"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
						{Label: "f2", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", "B", 2.0},
					},
				},
			},
		},
		{
			name: "group key loses all members. drops _value",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "grouper",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "A"},
						{execute.Time(2), 3.0, "m1", "f1", "A"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2", "B"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "grouper", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "grouper", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), 4.0, "m1", "f2", "A"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TString},
						{Label: "f2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", "A", "B"},
						{execute.Time(2), "m1", "A", "A"},
					},
				},
			},
		},
		{
			name: "_field flatten case one table with nulls",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "gg", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "m1", "f1", "gg1"},
						{execute.Time(2), 4.0, "m1", "f1", "gg1"},
						{execute.Time(3), nil, "m1", "f1", "gg1"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "m1", "f2"},
						{execute.Time(2), 5.0, "m1", "f2"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), nil, "m1", "f3"},
						{execute.Time(2), nil, "m1", "f3"},
						{execute.Time(4), 8.0, "m1", "f3"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "f1", Type: flux.TFloat},
						{Label: "f2", Type: flux.TFloat},
						{Label: "f3", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "m1", 1.0, 2.0, nil},
						{execute.Time(2), "m1", 4.0, 5.0, nil},
						{execute.Time(3), "m1", nil, nil, nil},
						{execute.Time(4), "m1", nil, nil, 8.0},
					},
				},
			},
		},
		{
			name: "pivot no data",
			spec: &universe.SortedPivotProcedureSpec{
				RowKey:      []string{"_time"},
				ColumnKey:   []string{"_field"},
				ValueColumn: "_value",
			},
			data: []flux.Table{},
			want: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data,
				tc.want,
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					spec := *tc.spec
					tr, d, err := universe.NewSortedPivotTransformation(context.Background(), spec, id, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tr, d
				})
		})
	}
}

func TestSortedPivot_Process_VariousSchemas(t *testing.T) {
	spec := universe.SortedPivotProcedureSpec{
		RowKey:      []string{"_time"},
		ColumnKey:   []string{"_field"},
		ValueColumn: "_value",
	}
	mem := &memory.ResourceAllocator{}
	id := executetest.RandomDatasetID()
	tr, d, err := universe.NewSortedPivotTransformation(context.Background(), spec, id, mem)
	if err != nil {
		t.Fatal(err)
	}

	store := executetest.NewDataStore()
	d.AddTransformation(store)

	tables, err := gen.Input(context.Background(), gen.Schema{
		Tags: []gen.Tag{
			{Name: "_measurement", Cardinality: 1},
			{Name: "_field", Cardinality: 10},
			{Name: "t0", Cardinality: 50},
			{Name: "t1", Cardinality: 5},
		},
		GroupBy: []string{"_measurement", "_field"},
		Nulls:   0.1,
		Types: map[flux.ColType]int{
			flux.TInt:    1,
			flux.TUInt:   1,
			flux.TFloat:  1,
			flux.TString: 1,
			flux.TBool:   1,
			flux.TTime:   1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	parent := executetest.RandomDatasetID()
	if err := tables.Do(func(table flux.Table) error {
		return tr.Process(parent, table)
	}); err != nil {
		t.Fatal(err)
	}

	d.Finish(nil)
	if err := store.Err(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkPivot(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		benchmarkPivot(b, 1000)
	})
}

func benchmarkPivot(b *testing.B, n int) {
	b.ReportAllocs()
	spec := &universe.SortedPivotProcedureSpec{
		RowKey:      []string{execute.DefaultTimeColLabel},
		ColumnKey:   []string{"_field"},
		ValueColumn: execute.DefaultValueColLabel,
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc memory.Allocator) (flux.TableIterator, error) {
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
		func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
			// cache := execute.NewTableBuilderCache(alloc)
			// d := execute.NewDataset(id, execute.DiscardingMode, cache)
			// t := NewPivotTransformation(d, cache, spec)
			t, d, err := universe.NewSortedPivotTransformation(context.Background(), *spec, id, alloc)
			if err != nil {
				b.Fatal(err)
			}
			return t, d
		},
	)
}
