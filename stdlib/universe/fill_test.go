package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux/stdlib/influxdata/influxdb"

	"github.com/influxdata/flux/plan"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestFillOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"fill","kind":"fill","spec":{"column":"t1","type":"float","value":"5.0"}}`)
	op := &flux.Operation{
		ID: "fill",
		Spec: &universe.FillOpSpec{
			Column: "t1",
			Type:   "float",
			Value:  "5.0",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)

	data = []byte(`{"id":"fill","kind":"fill","spec":{"column":"t1","type":"bool","value":"true"}}`)
	op = &flux.Operation{
		ID: "fill",
		Spec: &universe.FillOpSpec{
			Column: "t1",
			Type:   "bool",
			Value:  "true",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)

	data = []byte(`{"id":"fill","kind":"fill","spec":{"column":"t1","type":"int","value":"1"}}`)
	op = &flux.Operation{
		ID: "fill",
		Spec: &universe.FillOpSpec{
			Column: "t1",
			Type:   "int",
			Value:  "1",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)

	data = []byte(`{"id":"fill","kind":"fill","spec":{"column":"t1","type":"uint","value":"-1"}}`)
	op = &flux.Operation{
		ID: "fill",
		Spec: &universe.FillOpSpec{
			Column: "t1",
			Type:   "uint",
			Value:  "-1",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFill_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with range and fill",
			Raw:  `from(bucket:"mydb") |> range(start:-4h, stop:-2h) |> fill(column: "c1", value: 1.0)`,
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
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "fill2",
						Spec: &universe.FillOpSpec{
							Column: "c1",
							Type:   "float",
							Value:  "1",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "fill2"},
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

func TestFill_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.FillProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "nothing to fill",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(0.0),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
		},
		{
			name: "null bool",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(false),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{execute.Time(1), true},
					{execute.Time(2), nil},
					{execute.Time(3), false},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{execute.Time(1), true},
					{execute.Time(2), false},
					{execute.Time(3), false},
					{execute.Time(4), false},
				},
			}},
		},
		{
			name: "null int",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(int64(-1)),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2)},
					{execute.Time(2), nil},
					{execute.Time(3), int64(4)},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2)},
					{execute.Time(2), int64(-1)},
					{execute.Time(3), int64(4)},
					{execute.Time(4), int64(-1)},
				},
			}},
		},
		{
			name: "null uint",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(uint64(0)),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2)},
					{execute.Time(2), nil},
					{execute.Time(3), uint64(4)},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2)},
					{execute.Time(2), uint64(0)},
					{execute.Time(3), uint64(4)},
					{execute.Time(4), uint64(0)},
				},
			}},
		},
		{
			name: "null float",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(0.0),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), nil},
					{execute.Time(3), 4.0},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 0.0},
					{execute.Time(3), 4.0},
					{execute.Time(4), 0.0},
				},
			}},
		},
		{
			name: "null string",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New("UNK"),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), nil},
					{execute.Time(3), "B"},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), "UNK"},
					{execute.Time(3), "B"},
					{execute.Time(4), "UNK"},
				},
			}},
		},
		{
			name: "null time",
			spec: &universe.FillProcedureSpec{
				Column: "_time",
				Value:  values.New(execute.Time(0)),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{nil, "B"},
					{execute.Time(3), "B"},
					{nil, "C"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(0), "B"},
					{execute.Time(3), "B"},
					{execute.Time(0), "C"},
				},
			}},
		},
		{
			name: "fill previous",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), nil},
					{execute.Time(3), "B"},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), "A"},
					{execute.Time(3), "B"},
					{execute.Time(4), "B"},
				},
			}},
		},
		{
			name: "fill previous first nil",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), "A"},
					{execute.Time(3), "B"},
					{execute.Time(4), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), "A"},
					{execute.Time(3), "B"},
					{execute.Time(4), "B"},
				},
			}},
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
					return universe.NewFillTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
