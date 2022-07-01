package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestTripleExponentialDerivativeOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"tripleExponentialDerivative","kind":"tripleExponentialDerivative","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "tripleExponentialDerivative",
		Spec: &universe.TripleExponentialDerivativeOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestTripleExponentialDerivative_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewTripleExponentialDerivativeTransformation(
			d,
			c,
			nil,
			&universe.TripleExponentialDerivativeProcedureSpec{},
		)
		return s
	})
}

func TestTripleExponentialDerivative_Procedure(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.TripleExponentialDerivativeProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "ints",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "ints with chunks",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "floats",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "floats with chunks",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "uints",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "uints with chunks",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
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
					{execute.Time(11), float64(18.181818181818187)},
					{execute.Time(12), float64(15.384615384615374)},
					{execute.Time(13), float64(13.33333333333333)},
					{execute.Time(14), float64(11.764705882352944)},
					{execute.Time(15), float64(10.526315789473696)},
					{execute.Time(16), float64(8.304761904761904)},
					{execute.Time(17), float64(5.641927541329594)},
					{execute.Time(18), float64(3.0392222148231784)},
					{execute.Time(19), float64(0.716067574030288)},
					{execute.Time(20), float64(-1.2848911076603242)},
					{execute.Time(21), float64(-2.9999661985600445)},
					{execute.Time(22), float64(-4.493448741755913)},
					{execute.Time(23), float64(-5.836238000516913)},
					{execute.Time(24), float64(-7.099092024379772)},
					{execute.Time(25), float64(-8.352897627933453)},
					{execute.Time(26), float64(-9.673028502435233)},
					{execute.Time(27), float64(-11.147601363985949)},
					{execute.Time(28), float64(-12.891818138458877)},
					{execute.Time(29), float64(-15.074463280730022)},
				},
			}},
		},
		{
			name: "not enough values for first ema",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 10,
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
					{execute.Time(7), nil},
				},
			}},
		},
		{
			name: "not enough values for first ema with chunking",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 10,
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
					{execute.Time(7), nil},
				},
			}},
		},
		{
			name: "not enough values for second ema",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 5,
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
					{execute.Time(7), nil},
				},
			}},
		},
		{
			name: "not enough values for second ema with chunking",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 5,
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
					{execute.Time(7), nil},
				},
			}},
		},
		{
			name: "not enough values for third ema",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 5,
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
				},
			},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(10), nil},
				},
			}},
		},
		{
			name: "not enough values for second ema with chunking",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 5,
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
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(10), nil},
				},
			}},
		},
		{
			name: "pass through strings",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "string", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(1), "1"},
					{execute.Time(2), int64(2), "2"},
					{execute.Time(3), int64(3), "3"},
					{execute.Time(4), int64(4), "4"},
					{execute.Time(5), int64(5), "5"},
					{execute.Time(6), int64(6), "6"},
					{execute.Time(7), int64(7), "7"},
					{execute.Time(8), int64(8), "8"},
					{execute.Time(9), int64(9), "9"},
					{execute.Time(10), int64(10), "10"},
					{execute.Time(11), int64(11), "11"},
					{execute.Time(12), int64(12), "12"},
					{execute.Time(13), int64(13), "13"},
					{execute.Time(14), int64(14), "14"},
					{execute.Time(15), int64(15), "15"},
					{execute.Time(16), int64(14), "14"},
					{execute.Time(17), int64(13), "13"},
					{execute.Time(18), int64(12), "12"},
					{execute.Time(19), int64(11), "11"},
					{execute.Time(20), int64(10), "10"},
					{execute.Time(21), int64(9), "9"},
					{execute.Time(22), int64(8), "8"},
					{execute.Time(23), int64(7), "7"},
					{execute.Time(24), int64(6), "6"},
					{execute.Time(25), int64(5), "5"},
					{execute.Time(26), int64(4), "4"},
					{execute.Time(27), int64(3), "3"},
					{execute.Time(28), int64(2), "2"},
					{execute.Time(29), int64(1), "1"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "string", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), float64(18.181818181818187), "11"},
					{execute.Time(12), float64(15.384615384615374), "12"},
					{execute.Time(13), float64(13.33333333333333), "13"},
					{execute.Time(14), float64(11.764705882352944), "14"},
					{execute.Time(15), float64(10.526315789473696), "15"},
					{execute.Time(16), float64(8.304761904761904), "14"},
					{execute.Time(17), float64(5.641927541329594), "13"},
					{execute.Time(18), float64(3.0392222148231784), "12"},
					{execute.Time(19), float64(0.716067574030288), "11"},
					{execute.Time(20), float64(-1.2848911076603242), "10"},
					{execute.Time(21), float64(-2.9999661985600445), "9"},
					{execute.Time(22), float64(-4.493448741755913), "8"},
					{execute.Time(23), float64(-5.836238000516913), "7"},
					{execute.Time(24), float64(-7.099092024379772), "6"},
					{execute.Time(25), float64(-8.352897627933453), "5"},
					{execute.Time(26), float64(-9.673028502435233), "4"},
					{execute.Time(27), float64(-11.147601363985949), "3"},
					{execute.Time(28), float64(-12.891818138458877), "2"},
					{execute.Time(29), float64(-15.074463280730022), "1"},
				},
			}},
		},
		{
			name: "pass through strings with chunks",
			spec: &universe.TripleExponentialDerivativeProcedureSpec{
				N: 4,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "string", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(1), "1"},
						{execute.Time(2), int64(2), "2"},
						{execute.Time(3), int64(3), "3"},
						{execute.Time(4), int64(4), "4"},
						{execute.Time(5), int64(5), "5"},
						{execute.Time(6), int64(6), "6"},
						{execute.Time(7), int64(7), "7"},
						{execute.Time(8), int64(8), "8"},
						{execute.Time(9), int64(9), "9"},
						{execute.Time(10), int64(10), "10"},
						{execute.Time(11), int64(11), "11"},
						{execute.Time(12), int64(12), "12"},
						{execute.Time(13), int64(13), "13"},
						{execute.Time(14), int64(14), "14"},
						{execute.Time(15), int64(15), "15"},
						{execute.Time(16), int64(14), "14"},
						{execute.Time(17), int64(13), "13"},
						{execute.Time(18), int64(12), "12"},
						{execute.Time(19), int64(11), "11"},
						{execute.Time(20), int64(10), "10"},
						{execute.Time(21), int64(9), "9"},
						{execute.Time(22), int64(8), "8"},
						{execute.Time(23), int64(7), "7"},
						{execute.Time(24), int64(6), "6"},
						{execute.Time(25), int64(5), "5"},
						{execute.Time(26), int64(4), "4"},
						{execute.Time(27), int64(3), "3"},
						{execute.Time(28), int64(2), "2"},
						{execute.Time(29), int64(1), "1"},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "string", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), 18.181818181818187, "11"},
					{execute.Time(12), 15.384615384615374, "12"},
					{execute.Time(13), 13.33333333333333, "13"},
					{execute.Time(14), 11.764705882352944, "14"},
					{execute.Time(15), 10.526315789473696, "15"},
					{execute.Time(16), 8.304761904761904, "14"},
					{execute.Time(17), 5.641927541329594, "13"},
					{execute.Time(18), 3.0392222148231784, "12"},
					{execute.Time(19), 0.716067574030288, "11"},
					{execute.Time(20), -1.2848911076603242, "10"},
					{execute.Time(21), -2.9999661985600445, "9"},
					{execute.Time(22), -4.493448741755913, "8"},
					{execute.Time(23), -5.836238000516913, "7"},
					{execute.Time(24), -7.099092024379772, "6"},
					{execute.Time(25), -8.352897627933453, "5"},
					{execute.Time(26), -9.673028502435233, "4"},
					{execute.Time(27), -11.147601363985949, "3"},
					{execute.Time(28), -12.891818138458877, "2"},
					{execute.Time(29), -15.074463280730022, "1"},
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
					return universe.NewTripleExponentialDerivativeTransformation(d, c, nil, tc.spec)
				},
			)
		})
	}
}
