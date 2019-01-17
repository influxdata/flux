package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestPivot_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "pivot [_measurement, _field] around _time",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> pivot(rowKey: ["_time"], columnKey: ["_measurement", "_field"], valueColumn: "_value")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
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
				Edges: []flux.Edge{
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

func TestPivotOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"pivot",
		"kind":"pivot",
		"spec":{
			"rowKey":["_time"],
			"columnKey":["_measurement", "_field"], 
			"valueColumn":"_value"
		}
	}`)
	op := &flux.Operation{
		ID: "pivot",
		Spec: &universe.PivotOpSpec{
			RowKey:      []string{"_time"},
			ColumnKey:   []string{"_measurement", "_field"},
			ValueColumn: "_value",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestPivot_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.PivotProcedureSpec
		data []flux.Table
		want []*executetest.Table
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
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewPivotTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
