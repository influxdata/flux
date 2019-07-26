package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestExponentialMovingAverageOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"exponentialMovingAverage","kind":"exponentialMovingAverage","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "exponentialMovingAverage",
		Spec: &universe.ExponentialMovingAverageOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestExponentialMovingAverage_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewExponentialMovingAverageTransformation(
			d,
			c,
			&universe.ExponentialMovingAverageProcedureSpec{},
		)
		return s
	})
}

func TestExponentialMovingAverage_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.ExponentialMovingAverageProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "float",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 4.0},
					{execute.Time(3), 5.0},
					{execute.Time(4), 9.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "float with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 4.0},
						{execute.Time(3), 5.0},
						{execute.Time(4), 9.0},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "float with 3",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 4.0},
						{execute.Time(3), 5.0},
						{execute.Time(4), 9.0},
						{execute.Time(5), 8.0},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "float with 3 with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 4.0},
					{execute.Time(3), 5.0},
					{execute.Time(4), 9.0},
					{execute.Time(5), 8.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "int",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2)},
					{execute.Time(2), int64(4)},
					{execute.Time(3), int64(5)},
					{execute.Time(4), int64(9)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "int with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2)},
						{execute.Time(2), int64(4)},
						{execute.Time(3), int64(5)},
						{execute.Time(4), int64(9)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "int with 3",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2)},
					{execute.Time(2), int64(4)},
					{execute.Time(3), int64(5)},
					{execute.Time(4), int64(9)},
					{execute.Time(5), int64(8)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "int with 3 with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2)},
						{execute.Time(2), int64(4)},
						{execute.Time(3), int64(5)},
						{execute.Time(4), int64(9)},
						{execute.Time(5), int64(8)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "uint",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2)},
					{execute.Time(2), uint64(4)},
					{execute.Time(3), uint64(5)},
					{execute.Time(4), uint64(9)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "uint with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       2,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(2)},
						{execute.Time(2), uint64(4)},
						{execute.Time(3), uint64(5)},
						{execute.Time(4), uint64(9)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0},
					{execute.Time(3), (5.0 * 2.0 / 3.0) + 3.0*(1-2.0/3.0)},
					{execute.Time(4), (9.0 * 2.0 / 3.0) + ((5.0*2.0/3.0)+3.0*(1-2.0/3.0))*(1-2.0/3.0)},
				},
			}},
		},
		{
			name: "uint with 3",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2)},
					{execute.Time(2), uint64(4)},
					{execute.Time(3), uint64(5)},
					{execute.Time(4), uint64(9)},
					{execute.Time(5), uint64(8)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "uint with 3 with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(2)},
						{execute.Time(2), uint64(4)},
						{execute.Time(3), uint64(5)},
						{execute.Time(4), uint64(9)},
						{execute.Time(5), uint64(8)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5)},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5},
				},
			}},
		},
		{
			name: "float with tags",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, "a"},
					{execute.Time(2), 4.0, "b"},
					{execute.Time(3), 5.0, "c"},
					{execute.Time(4), 9.0, "d"},
					{execute.Time(5), 8.0, "e"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3, "c"},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5), "d"},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5, "e"},
				},
			}},
		},
		{
			name: "float with tags and chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       3,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a"},
						{execute.Time(2), 4.0, "b"},
						{execute.Time(3), 5.0, "c"},
						{execute.Time(4), 9.0, "d"},
						{execute.Time(5), 8.0, "e"},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 11.0 / 3, "c"},
					{execute.Time(4), 4.5 + (11.0/3)*(1.0-0.5), "d"},
					{execute.Time(5), 4.0 + (4.5+(11.0/3)*(1.0-0.5))*0.5, "e"},
				},
			}},
		},
		{
			name: "ints with null values",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{"x", "y", "z"},
				N:       2,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
					{Label: "z", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), nil, nil},
					{execute.Time(2), nil, int64(10), nil},
					{execute.Time(3), int64(8), int64(20), int64(4)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
					{Label: "z", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0, 10.0, nil},
					{execute.Time(3), 6.0, 50.0 / 3, 4.0},
				},
			}},
		},
		{
			name: "ints with null values and chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{"x", "y", "z"},
				N:       2,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "x", Type: flux.TInt},
						{Label: "y", Type: flux.TInt},
						{Label: "z", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), nil, nil},
						{execute.Time(2), nil, int64(10), nil},
						{execute.Time(3), int64(8), int64(20), int64(4)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
					{Label: "z", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0, 10.0, nil},
					{execute.Time(3), 6.0, 50.0 / 3, 4.0},
				},
			}},
		},
		{
			name: "pass over column",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{"x", "y"},
				N:       2,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
					{Label: "z", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), nil, nil},
					{execute.Time(2), nil, int64(10), nil},
					{execute.Time(3), int64(8), int64(20), int64(4)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
					{Label: "z", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0, 10.0, nil},
					{execute.Time(3), 6.0, 50.0 / 3, int64(4)},
				},
			}},
		},
		{
			name: "pass over column with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{"x", "y"},
				N:       2,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "x", Type: flux.TInt},
						{Label: "y", Type: flux.TInt},
						{Label: "z", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), nil, nil},
						{execute.Time(2), nil, int64(10), nil},
						{execute.Time(3), int64(8), int64(20), int64(4)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
					{Label: "z", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 2.0, 10.0, nil},
					{execute.Time(3), 6.0, 50.0 / 3, int64(4)},
				},
			}},
		},
		{
			name: "ints with less rows than period",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "_value2", Type: flux.TInt},
					{Label: "pass", Type: flux.TInt},
					{Label: "passNull", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), nil, int64(1), nil},
					{execute.Time(2), int64(4), nil, int64(2), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TInt},
					{Label: "passNull", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, int64(2), nil},
				},
			}},
		},
		{
			name: "ints with less rows than period with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "_value2", Type: flux.TInt},
						{Label: "pass", Type: flux.TInt},
						{Label: "passNull", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), nil, int64(1), nil},
						{execute.Time(2), int64(4), nil, int64(2), nil},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TInt},
					{Label: "passNull", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, int64(2), nil},
				},
			}},
		},
		{
			name: "uints with less rows than period",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
					{Label: "_value2", Type: flux.TUInt},
					{Label: "pass", Type: flux.TUInt},
					{Label: "passNull", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), nil, uint64(1), nil},
					{execute.Time(2), uint64(4), nil, uint64(2), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TUInt},
					{Label: "passNull", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, uint64(2), nil},
				},
			}},
		},
		{
			name: "uints with less rows than period with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
						{Label: "_value2", Type: flux.TUInt},
						{Label: "pass", Type: flux.TUInt},
						{Label: "passNull", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(2), nil, uint64(1), nil},
						{execute.Time(2), uint64(4), nil, uint64(2), nil},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TUInt},
					{Label: "passNull", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, uint64(2), nil},
				},
			}},
		},
		{
			name: "floats with less rows than period",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TFloat},
					{Label: "passNull", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), float64(2), nil, float64(1), nil},
					{execute.Time(2), float64(4), nil, float64(2), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TFloat},
					{Label: "passNull", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, float64(2), nil},
				},
			}},
		},
		{
			name: "floats with less rows than period with chunking",
			spec: &universe.ExponentialMovingAverageProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel, "_value2"},
				N:       5,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_value2", Type: flux.TFloat},
						{Label: "pass", Type: flux.TFloat},
						{Label: "passNull", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), float64(2), nil, float64(1), nil},
						{execute.Time(2), float64(4), nil, float64(2), nil},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_value2", Type: flux.TFloat},
					{Label: "pass", Type: flux.TFloat},
					{Label: "passNull", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 3.0, nil, float64(2), nil},
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
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewExponentialMovingAverageTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
