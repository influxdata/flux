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

func TestColumns_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from range columns",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> columns()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop:        flux.Now,
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "columns2",
						Spec: &universe.ColumnsOpSpec{
							Column: execute.DefaultValueColLabel,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "columns2"},
				},
			},
		},
		{
			Name: "from columns custom label",
			Raw:  `from(bucket:"mydb") |> columns(column: "labels")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
						},
					},
					{
						ID: "columns1",
						Spec: &universe.ColumnsOpSpec{
							Column: "labels",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "columns1"},
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

func TestColumnsOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"columns","kind":"columns","spec":{"column":"new"}}`)
	op := &flux.Operation{
		ID: "columns",
		Spec: &universe.ColumnsOpSpec{
			Column: "new",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestColumns_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.ColumnsProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "one table",
			spec: &universe.ColumnsProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "valid", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b", true},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0", "tag1"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "b", "_time"},
					{"a", "b", "_value"},
					{"a", "b", "tag0"},
					{"a", "b", "tag1"},
					{"a", "b", "valid"},
				},
			}},
		},
		{
			name: "three tables",
			spec: &universe.ColumnsProcedureSpec{
				Column: "val",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "b", "b"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "b", "c"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "val", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"a", "b", "_time"},
						{"a", "b", "_value"},
						{"a", "b", "tag0"},
						{"a", "b", "tag1"},
					},
				},
				{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "val", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"b", "b", "_time"},
						{"b", "b", "_value"},
						{"b", "b", "tag0"},
						{"b", "b", "tag1"},
					},
				},
				{
					KeyCols: []string{"tag0", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "val", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"b", "c", "_time"},
						{"b", "c", "_value"},
						{"b", "c", "tag0"},
						{"b", "c", "tag2"},
					},
				},
			},
		},
		{
			name: "with nulls",
			spec: &universe.ColumnsProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a", "b"},
						{execute.Time(2), nil, "a", "b"},
						{nil, 3.0, "a", "b"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{nil, 1.0, "a", nil},
						{nil, 2.0, "a", nil},
						{execute.Time(3), 3.0, "a", nil},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, nil},
						{execute.Time(2), nil, nil},
						{execute.Time(3), nil, nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"a", "b", "_time"},
						{"a", "b", "_value"},
						{"a", "b", "tag0"},
						{"a", "b", "tag1"},
					},
				},
				{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"a", nil, "_time"},
						{"a", nil, "_value"},
						{"a", nil, "tag0"},
						{"a", nil, "tag1"},
					},
				},
				{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{nil, "_time"},
						{nil, "_value"},
						{nil, "tag0"},
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
					return universe.NewColumnsTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
