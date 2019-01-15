package universe_test

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestUnion_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "basic two-way union",
			Raw: `
				a = from(bucket:"dbA") |> range(start:-1h)
				b = from(bucket:"dbB") |> range(start:-1h)
				union(tables: [a, b])`,
			Want: &flux.Spec{Operations: []*flux.Operation{
				{
					ID:   "from0",
					Spec: &influxdb.FromOpSpec{Bucket: "dbA"},
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
					ID:   "from2",
					Spec: &influxdb.FromOpSpec{Bucket: "dbB"},
				},
				{
					ID: "range3",
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
					ID:   "union4",
					Spec: &universe.UnionOpSpec{},
				},
			},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "from2", Child: "range3"},
					{Parent: "range1", Child: "union4"},
					{Parent: "range3", Child: "union4"},
				},
			},
		},
		{
			Name: "basic three-way union",
			Raw: `
				a = from(bucket:"dbA") |> range(start:-1h)
				b = from(bucket:"dbB") |> range(start:-1h)
				c = from(bucket:"dbC") |> range(start:-1h)
				union(tables: [a, b, c])`,
			Want: &flux.Spec{Operations: []*flux.Operation{
				{
					ID:   "from0",
					Spec: &influxdb.FromOpSpec{Bucket: "dbA"},
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
					ID:   "from2",
					Spec: &influxdb.FromOpSpec{Bucket: "dbB"},
				},
				{
					ID: "range3",
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
					ID:   "from4",
					Spec: &influxdb.FromOpSpec{Bucket: "dbC"},
				},
				{
					ID: "range5",
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
					ID:   "union6",
					Spec: &universe.UnionOpSpec{},
				},
			},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "from2", Child: "range3"},
					{Parent: "from4", Child: "range5"},
					{Parent: "range1", Child: "union6"},
					{Parent: "range3", Child: "union6"},
					{Parent: "range5", Child: "union6"},
				},
			},
		},
		{
			Name: "union no argument",
			Raw: `
				union()`,
			WantErr: true,
		},
		{
			Name: "one-way union",
			Raw: `
				b = from(bucket:"dbB") |> range(start:-1h)
				union(tables: [b])`,
			WantErr: true,
		},
		{
			Name: "non-table union",
			Raw: `
				union(tables: [{a: "a"}, {a: "b"}])`,
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

func TestUnionOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"union",
		"kind":"union",
		"spec":{
		}
	}`)
	op := &flux.Operation{
		ID:   "union",
		Spec: &universe.UnionOpSpec{},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestUnion_Process(t *testing.T) {
	spec := &universe.UnionProcedureSpec{}

	testCases := []struct {
		name string
		data [][]flux.Table // data from parents
		want []*executetest.Table
	}{
		{
			name: "two streams union same schema",
			data: [][]flux.Table{
				// stream 1
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 70.0},
							{execute.Time(2), "temp", 75.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "humidity", 81.0},
							{execute.Time(2), "humidity", 82.0},
						},
					},
				},
				// stream 2
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 55.0},
							{execute.Time(2), "temp", 56.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", 29.82},
							{execute.Time(2), "pressure", 30.01},
						},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "temp", 70.0},
						{execute.Time(2), "temp", 75.0},
						{execute.Time(1), "temp", 55.0},
						{execute.Time(2), "temp", 56.0},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "humidity", 81.0},
						{execute.Time(2), "humidity", 82.0},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "pressure", 29.82},
						{execute.Time(2), "pressure", 30.01},
					},
				},
			},
		},
		{
			name: "two streams union heterogeneous schema",
			data: [][]flux.Table{
				// stream 1
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 70.0},
							{execute.Time(2), "temp", 75.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "humidity", 81.0},
							{execute.Time(2), "humidity", 82.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
							{Label: "room", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", 42.0, "r0"},
						},
					},
				},
				// stream 2
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
							{Label: "room", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 55.0, "r1"},
							{execute.Time(2), "temp", 56.0, "r0"},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", 29.82},
							{execute.Time(2), "pressure", 30.01},
						},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "room", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "temp", 70.0, nil},
						{execute.Time(2), "temp", 75.0, nil},
						{execute.Time(1), "temp", 55.0, "r1"},
						{execute.Time(2), "temp", 56.0, "r0"},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "humidity", 81.0},
						{execute.Time(2), "humidity", 82.0},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "room", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "pressure", 42.0, "r0"},
						{execute.Time(1), "pressure", 29.82, nil},
						{execute.Time(2), "pressure", 30.01, nil},
					},
				},
			},
		},
		{
			name: "two streams union heterogeneous schema group key",
			data: [][]flux.Table{
				// stream 1
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 70.0},
							{execute.Time(2), "temp", 75.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "humidity", 81.0},
							{execute.Time(2), "humidity", 82.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field", "room"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
							{Label: "room", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", 42.0, "r0"},
						},
					},
				},
				// stream 2
				{
					&executetest.Table{
						KeyCols: []string{"_field", "room"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
							{Label: "room", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", 55.0, "r1"},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field", "room"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
							{Label: "room", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(2), "temp", 56.0, "r0"},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", 29.82},
							{execute.Time(2), "pressure", 30.01},
						},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "temp", 70.0},
						{execute.Time(2), "temp", 75.0},
					},
				},
				{
					KeyCols: []string{"_field", "room"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "room", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "temp", 55.0, "r1"},
					},
				},
				{
					KeyCols: []string{"_field", "room"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "room", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(2), "temp", 56.0, "r0"},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "humidity", 81.0},
						{execute.Time(2), "humidity", 82.0},
					},
				},
				{
					KeyCols: []string{"_field", "room"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "room", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "pressure", 42.0, "r0"},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "pressure", 29.82},
						{execute.Time(2), "pressure", 30.01},
					},
				},
			},
		},
		{
			name: "three streams union with nulls",
			data: [][]flux.Table{
				// stream 1
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "temp", nil},
							{nil, "temp", 75.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), nil, 18.0},
							{execute.Time(2), nil, 23.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "humidity", nil},
							{execute.Time(2), "humidity", 82.0},
						},
					},
				},
				// stream 2
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{nil, "temp", 55.0},
							{execute.Time(2), "temp", 56.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "pressure", nil},
							{execute.Time(2), "pressure", 30.01},
						},
					},
				},
				// stream 3
				{
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{nil, "temp", 42.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{nil, "humidity", 55.82},
							{nil, "humidity", 32.01},
						},
					},
					&executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{nil, nil, nil},
							{nil, nil, nil},
						},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "temp", nil},
						{nil, "temp", 75.0},
						{nil, "temp", 55.0},
						{execute.Time(2), "temp", 56.0},
						{nil, "temp", 42.0},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), nil, 18.0},
						{execute.Time(2), nil, 23.0},
						{nil, nil, nil},
						{nil, nil, nil},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "humidity", nil},
						{execute.Time(2), "humidity", 82.0},
						{nil, "humidity", 55.82},
						{nil, "humidity", 32.01},
					},
				},
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "pressure", nil},
						{execute.Time(2), "pressure", 30.01},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parentIds := make([]execute.DatasetID, len(tc.data))
			for i := 0; i < len(parentIds); i++ {
				parentIds[i] = executetest.RandomDatasetID()
			}

			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(execute.DefaultTriggerSpec)
			ut := universe.NewUnionTransformation(d, c, spec, parentIds)

			for i, s := range tc.data {
				for _, tbl := range s {
					if err := ut.Process(parentIds[i], tbl); err != nil {
						t.Fatal(err)
					}
				}
			}

			got, err := executetest.TablesFromCache(c)
			if err != nil {
				t.Fatal(err)
			}

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(tc.want)

			sort.Sort(executetest.SortedTables(got))
			sort.Sort(executetest.SortedTables(tc.want))

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
