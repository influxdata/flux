package universe_test

import (
	"math"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestDEMAOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"doubleExponentialMovingAverage","kind":"doubleExponentialMovingAverage","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "doubleExponentialMovingAverage",
		Spec: &universe.DEMAOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestDEMA_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewDEMATransformation(
			d,
			c,
			nil,
			&universe.DEMAProcedureSpec{},
		)
		return s
	})
}

func TestDEMA_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.DEMAProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "ints",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(1)},
					{execute.Time(2), int64(2)},
					{execute.Time(3), int64(3)},
					{execute.Time(4), int64(4)},
					{execute.Time(5), int64(5)},
					{execute.Time(6), int64(6)},
					{execute.Time(7), int64(7)},
					{execute.Time(8), int64(8)},
					{execute.Time(9), int64(9)},
					{execute.Time(10), int64(10)},
					{execute.Time(11), int64(11)},
					{execute.Time(12), int64(12)},
					{execute.Time(13), int64(13)},
					{execute.Time(14), int64(14)},
					{execute.Time(15), int64(15)},
					{execute.Time(16), int64(14)},
					{execute.Time(17), int64(13)},
					{execute.Time(18), int64(12)},
					{execute.Time(19), int64(11)},
					{execute.Time(20), int64(10)},
					{execute.Time(21), int64(9)},
					{execute.Time(22), int64(8)},
					{execute.Time(23), int64(7)},
					{execute.Time(24), int64(6)},
					{execute.Time(25), int64(5)},
					{execute.Time(26), int64(4)},
					{execute.Time(27), int64(3)},
					{execute.Time(28), int64(2)},
					{execute.Time(29), int64(1)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "ints with chunking",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(1)},
						{execute.Time(2), int64(2)},
						{execute.Time(3), int64(3)},
						{execute.Time(4), int64(4)},
						{execute.Time(5), int64(5)},
						{execute.Time(6), int64(6)},
						{execute.Time(7), int64(7)},
						{execute.Time(8), int64(8)},
						{execute.Time(9), int64(9)},
						{execute.Time(10), int64(10)},
						{execute.Time(11), int64(11)},
						{execute.Time(12), int64(12)},
						{execute.Time(13), int64(13)},
						{execute.Time(14), int64(14)},
						{execute.Time(15), int64(15)},
						{execute.Time(16), int64(14)},
						{execute.Time(17), int64(13)},
						{execute.Time(18), int64(12)},
						{execute.Time(19), int64(11)},
						{execute.Time(20), int64(10)},
						{execute.Time(21), int64(9)},
						{execute.Time(22), int64(8)},
						{execute.Time(23), int64(7)},
						{execute.Time(24), int64(6)},
						{execute.Time(25), int64(5)},
						{execute.Time(26), int64(4)},
						{execute.Time(27), int64(3)},
						{execute.Time(28), int64(2)},
						{execute.Time(29), int64(1)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "uints",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(1)},
					{execute.Time(2), uint64(2)},
					{execute.Time(3), uint64(3)},
					{execute.Time(4), uint64(4)},
					{execute.Time(5), uint64(5)},
					{execute.Time(6), uint64(6)},
					{execute.Time(7), uint64(7)},
					{execute.Time(8), uint64(8)},
					{execute.Time(9), uint64(9)},
					{execute.Time(10), uint64(10)},
					{execute.Time(11), uint64(11)},
					{execute.Time(12), uint64(12)},
					{execute.Time(13), uint64(13)},
					{execute.Time(14), uint64(14)},
					{execute.Time(15), uint64(15)},
					{execute.Time(16), uint64(14)},
					{execute.Time(17), uint64(13)},
					{execute.Time(18), uint64(12)},
					{execute.Time(19), uint64(11)},
					{execute.Time(20), uint64(10)},
					{execute.Time(21), uint64(9)},
					{execute.Time(22), uint64(8)},
					{execute.Time(23), uint64(7)},
					{execute.Time(24), uint64(6)},
					{execute.Time(25), uint64(5)},
					{execute.Time(26), uint64(4)},
					{execute.Time(27), uint64(3)},
					{execute.Time(28), uint64(2)},
					{execute.Time(29), uint64(1)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "uints with chunking",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(1)},
						{execute.Time(2), uint64(2)},
						{execute.Time(3), uint64(3)},
						{execute.Time(4), uint64(4)},
						{execute.Time(5), uint64(5)},
						{execute.Time(6), uint64(6)},
						{execute.Time(7), uint64(7)},
						{execute.Time(8), uint64(8)},
						{execute.Time(9), uint64(9)},
						{execute.Time(10), uint64(10)},
						{execute.Time(11), uint64(11)},
						{execute.Time(12), uint64(12)},
						{execute.Time(13), uint64(13)},
						{execute.Time(14), uint64(14)},
						{execute.Time(15), uint64(15)},
						{execute.Time(16), uint64(14)},
						{execute.Time(17), uint64(13)},
						{execute.Time(18), uint64(12)},
						{execute.Time(19), uint64(11)},
						{execute.Time(20), uint64(10)},
						{execute.Time(21), uint64(9)},
						{execute.Time(22), uint64(8)},
						{execute.Time(23), uint64(7)},
						{execute.Time(24), uint64(6)},
						{execute.Time(25), uint64(5)},
						{execute.Time(26), uint64(4)},
						{execute.Time(27), uint64(3)},
						{execute.Time(28), uint64(2)},
						{execute.Time(29), uint64(1)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "floats",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), float64(1)},
					{execute.Time(2), float64(2)},
					{execute.Time(3), float64(3)},
					{execute.Time(4), float64(4)},
					{execute.Time(5), float64(5)},
					{execute.Time(6), float64(6)},
					{execute.Time(7), float64(7)},
					{execute.Time(8), float64(8)},
					{execute.Time(9), float64(9)},
					{execute.Time(10), float64(10)},
					{execute.Time(11), float64(11)},
					{execute.Time(12), float64(12)},
					{execute.Time(13), float64(13)},
					{execute.Time(14), float64(14)},
					{execute.Time(15), float64(15)},
					{execute.Time(16), float64(14)},
					{execute.Time(17), float64(13)},
					{execute.Time(18), float64(12)},
					{execute.Time(19), float64(11)},
					{execute.Time(20), float64(10)},
					{execute.Time(21), float64(9)},
					{execute.Time(22), float64(8)},
					{execute.Time(23), float64(7)},
					{execute.Time(24), float64(6)},
					{execute.Time(25), float64(5)},
					{execute.Time(26), float64(4)},
					{execute.Time(27), float64(3)},
					{execute.Time(28), float64(2)},
					{execute.Time(29), float64(1)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "floats with chunking",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), float64(1)},
						{execute.Time(2), float64(2)},
						{execute.Time(3), float64(3)},
						{execute.Time(4), float64(4)},
						{execute.Time(5), float64(5)},
						{execute.Time(6), float64(6)},
						{execute.Time(7), float64(7)},
						{execute.Time(8), float64(8)},
						{execute.Time(9), float64(9)},
						{execute.Time(10), float64(10)},
						{execute.Time(11), float64(11)},
						{execute.Time(12), float64(12)},
						{execute.Time(13), float64(13)},
						{execute.Time(14), float64(14)},
						{execute.Time(15), float64(15)},
						{execute.Time(16), float64(14)},
						{execute.Time(17), float64(13)},
						{execute.Time(18), float64(12)},
						{execute.Time(19), float64(11)},
						{execute.Time(20), float64(10)},
						{execute.Time(21), float64(9)},
						{execute.Time(22), float64(8)},
						{execute.Time(23), float64(7)},
						{execute.Time(24), float64(6)},
						{execute.Time(25), float64(5)},
						{execute.Time(26), float64(4)},
						{execute.Time(27), float64(3)},
						{execute.Time(28), float64(2)},
						{execute.Time(29), float64(1)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239)},
					{execute.Time(20), float64(12.70174811931398)},
					{execute.Time(21), float64(11.701405062848782)},
					{execute.Time(22), float64(10.611872766773772)},
					{execute.Time(23), float64(9.465595022565747)},
					{execute.Time(24), float64(8.286166283961508)},
					{execute.Time(25), float64(7.0904770859219255)},
					{execute.Time(26), float64(5.890371851336026)},
					{execute.Time(27), float64(4.6939254760732005)},
					{execute.Time(28), float64(3.5064225149113675)},
					{execute.Time(29), float64(2.3311049123183603)},
				},
			}},
		},
		{
			name: "pass through",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), float64(1), "a", int64(1)},
						{execute.Time(2), float64(2), "b", int64(2)},
						{execute.Time(3), float64(3), "c", int64(3)},
						{execute.Time(4), float64(4), "d", int64(4)},
						{execute.Time(5), float64(5), "e", int64(5)},
						{execute.Time(6), float64(6), "f", int64(6)},
						{execute.Time(7), float64(7), "g", int64(7)},
						{execute.Time(8), float64(8), "h", int64(8)},
						{execute.Time(9), float64(9), "i", int64(9)},
						{execute.Time(10), float64(10), "j", int64(10)},
						{execute.Time(11), float64(11), "k", int64(11)},
						{execute.Time(12), float64(12), "l", int64(12)},
						{execute.Time(13), float64(13), "m", int64(13)},
						{execute.Time(14), float64(14), "n", int64(14)},
						{execute.Time(15), float64(15), "o", int64(15)},
						{execute.Time(16), float64(14), "p", int64(14)},
						{execute.Time(17), float64(13), "q", int64(13)},
						{execute.Time(18), float64(12), "r", int64(12)},
						{execute.Time(19), float64(11), "s", int64(11)},
						{execute.Time(20), float64(10), "t", int64(10)},
						{execute.Time(21), float64(9), "u", int64(9)},
						{execute.Time(22), float64(8), "v", int64(8)},
						{execute.Time(23), float64(7), "w", int64(7)},
						{execute.Time(24), float64(6), "x", int64(6)},
						{execute.Time(25), float64(5), "y", int64(5)},
						{execute.Time(26), float64(4), "z", int64(4)},
						{execute.Time(27), float64(3), "aa", int64(3)},
						{execute.Time(28), float64(2), "ab", int64(2)},
						{execute.Time(29), float64(1), "ac", int64(1)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(19), float64(13.568840926166239), "s", int64(11)},
					{execute.Time(20), float64(12.70174811931398), "t", int64(10)},
					{execute.Time(21), float64(11.701405062848782), "u", int64(9)},
					{execute.Time(22), float64(10.611872766773772), "v", int64(8)},
					{execute.Time(23), float64(9.465595022565747), "w", int64(7)},
					{execute.Time(24), float64(8.286166283961508), "x", int64(6)},
					{execute.Time(25), float64(7.0904770859219255), "y", int64(5)},
					{execute.Time(26), float64(5.890371851336026), "z", int64(4)},
					{execute.Time(27), float64(4.6939254760732005), "aa", int64(3)},
					{execute.Time(28), float64(3.5064225149113675), "ab", int64(2)},
					{execute.Time(29), float64(2.3311049123183603), "ac", int64(1)},
				},
			}},
		},
		{
			name: "not enough values for first ema",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), float64(1)},
					{execute.Time(2), float64(2)},
					{execute.Time(3), float64(3)},
					{execute.Time(4), float64(4)},
					{execute.Time(5), float64(5)},
					{execute.Time(6), float64(6)},
					{execute.Time(7), float64(7)},
				},
			},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(7), math.NaN()},
				},
			}},
		},
		{
			name: "not enough values for first ema with chunking",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       10,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), float64(1)},
						{execute.Time(2), float64(2)},
						{execute.Time(3), float64(3)},
						{execute.Time(4), float64(4)},
						{execute.Time(5), float64(5)},
						{execute.Time(6), float64(6)},
						{execute.Time(7), float64(7)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(7), math.NaN()},
				},
			}},
		},
		{
			name: "not enough values for second ema",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       5,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), float64(1)},
					{execute.Time(2), float64(2)},
					{execute.Time(3), float64(3)},
					{execute.Time(4), float64(4)},
					{execute.Time(5), float64(5)},
					{execute.Time(6), float64(6)},
					{execute.Time(7), float64(7)},
				},
			},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(7), math.NaN()},
				},
			}},
		},
		{
			name: "not enough values for second ema with chunking",
			spec: &universe.DEMAProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
				N:       5,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), float64(1)},
						{execute.Time(2), float64(2)},
						{execute.Time(3), float64(3)},
						{execute.Time(4), float64(4)},
						{execute.Time(5), float64(5)},
						{execute.Time(6), float64(6)},
						{execute.Time(7), float64(7)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(7), math.NaN()},
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
					return universe.NewDEMATransformation(d, c, nil, tc.spec)
				},
			)
		})
	}
}
