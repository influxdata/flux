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

func TestKeys_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from range keys",
			Raw:  `from(bucket: "mydb") |> range(start:-1h) |> keys()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mydb",
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
						ID: "keys2",
						Spec: &universe.KeysOpSpec{
							Column: "_value",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "keys2"},
				},
			},
		},
		{
			Name: "from keys custom label",
			Raw:  `from(bucket: "mydb") |> keys(column: "keys")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mydb",
						},
					},
					{
						ID: "keys1",
						Spec: &universe.KeysOpSpec{
							Column: "keys",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "keys1"},
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

func TestKeysOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"keys","kind":"keys","spec":{"column":"keys"}}`)
	op := &flux.Operation{
		ID: "keys",
		Spec: &universe.KeysOpSpec{
			Column: "keys",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestKeys_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.KeysProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "one table no keys",
			spec: &universe.KeysProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
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
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: nil,
			}},
		},
		{
			name: "one table with keys",
			spec: &universe.KeysProcedureSpec{
				Column: "keys",
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
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0", "tag1"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
					{Label: "keys", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "b", "tag0"},
					{"a", "b", "tag1"},
				},
			}},
		},
		{
			name: "two tables with keys",
			spec: &universe.KeysProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", "b", execute.Time(1), 2.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", "c", execute.Time(1), 2.0},
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
						{"a", "b", "tag0"},
						{"a", "b", "tag1"},
					},
				},
				{
					KeyCols: []string{"tag0", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"a", "c", "tag0"},
						{"a", "c", "tag2"},
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
					return universe.NewKeysTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
