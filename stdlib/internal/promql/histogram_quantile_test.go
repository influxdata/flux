package promql_test

import (
	"math"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/internal/promql"
	"github.com/pkg/errors"
)

var linearHist = func() []flux.Table {
	return []flux.Table{&executetest.Table{
		KeyCols: []string{"_start", "_stop"},
		ColMeta: []flux.ColMeta{
			{Label: "_start", Type: flux.TTime},
			{Label: "_stop", Type: flux.TTime},
			{Label: "_time", Type: flux.TTime},
			{Label: "le", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
		Data: [][]interface{}{
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.1", 1.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.3", 3.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.5", 5.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.7", 7.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 8.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.9", 9.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 10.0},
		},
	}}
}

var linearHistNoMax = func() []flux.Table {
	return []flux.Table{&executetest.Table{
		KeyCols: []string{"_start", "_stop"},
		ColMeta: []flux.ColMeta{
			{Label: "_start", Type: flux.TTime},
			{Label: "_stop", Type: flux.TTime},
			{Label: "_time", Type: flux.TTime},
			{Label: "le", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
		Data: [][]interface{}{
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 8.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
		},
	}}
}

var unsortedOddHist = func() []flux.Table {
	return []flux.Table{&executetest.Table{
		KeyCols: []string{"_start", "_stop"},
		ColMeta: []flux.ColMeta{
			{Label: "_start", Type: flux.TTime},
			{Label: "_stop", Type: flux.TTime},
			{Label: "_time", Type: flux.TTime},
			{Label: "le", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
		Data: [][]interface{}{
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 10.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 10.0},
		},
	}}
}

var nonLinearHist = func() []flux.Table {
	return []flux.Table{&executetest.Table{
		KeyCols: []string{"_start", "_stop"},
		ColMeta: []flux.ColMeta{
			{Label: "_start", Type: flux.TTime},
			{Label: "_stop", Type: flux.TTime},
			{Label: "_time", Type: flux.TTime},
			{Label: "le", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
		Data: [][]interface{}{
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.1", 1.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "0.5", 5.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
			{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 11.0},
		},
	}}
}

func TestHistogramQuantile_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *promql.HistogramQuantileProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "90th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.9,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.9},
				},
			}},
		},
		{
			name: "0th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.0,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.0},
				},
			}},
		},
		{
			name: "5th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.05,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.05},
				},
			}},
		},
		{
			name: "10th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.1,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.1},
				},
			}},
		},
		{
			name: "95th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.95,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.95},
				},
			}},
		},
		{
			name: "99.999th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.99999,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.99999},
				},
			}},
		},
		{
			name: "100th linear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         1.0,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 1.0},
				},
			}},
		},
		{
			name: "100th linear no max",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         1.0,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: linearHistNoMax(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), math.NaN()},
				},
			}},
		},
		{
			name: "90th linear unsorted odd",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.9,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: unsortedOddHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.75},
				},
			}},
		},
		{
			name: "100th linear unsorted odd",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         1.0,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: unsortedOddHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.8},
				},
			}},
		},
		{
			name: "90th nonlinear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.90,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: nonLinearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.99},
				},
			}},
		},
		{
			name: "highest finite upper bound nonlinear",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.99,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: nonLinearHist(),
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 1.0},
				},
			}},
		},
		{
			name: "null in count column",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.9,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "le", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.1", 1.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.3", 3.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.5", 5.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.7", nil},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 8.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.9", 9.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 10.0},
				},
			}},
			wantErr: errors.New("unexpected null in the countColumn"),
		},
		{
			name: "null in upperBound column",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.9,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "le", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.1", 1.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.3", 3.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.5", 5.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), nil, 7.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 8.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.9", 9.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 10.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.9},
				},
			}},
		},
		{
			name: "unparseable value in upperBound column",
			spec: &promql.HistogramQuantileProcedureSpec{
				Quantile:         0.9,
				CountColumn:      "_value",
				UpperBoundColumn: "le",
				ValueColumn:      "_value",
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "le", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.1", 1.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.2", 2.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.3", 3.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.4", 4.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.5", 5.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.6", 6.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "invalid", 7.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.8", 8.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "0.9", 9.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "1.0", 10.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), "+Inf", 10.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), 0.9},
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
					return promql.NewHistogramQuantileTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
