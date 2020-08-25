package universe_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestIntegralOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"integral","kind":"integral","spec":{"unit":"1m"}}`)
	op := &flux.Operation{
		ID: "integral",
		Spec: &universe.IntegralOpSpec{
			Unit: flux.ConvertDuration(time.Minute),
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestIntegral_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewIntegralTransformation(
			d,
			c,
			&universe.IntegralProcedureSpec{},
		)
		return s
	})
}

func TestIntegral_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.IntegralProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "float",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), 2.0},
					{execute.Time(1), execute.Time(3), execute.Time(2), 1.0},
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
					{execute.Time(1), execute.Time(3), 1.5},
				},
			}},
		},
		{
			name: "int",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), int64(2)},
					{execute.Time(1), execute.Time(3), execute.Time(2), int64(1)},
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
					{execute.Time(1), execute.Time(3), 1.5},
				},
			}},
		},
		{
			name: "uint",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), uint64(2)},
					{execute.Time(1), execute.Time(3), execute.Time(2), uint64(1)},
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
					{execute.Time(1), execute.Time(3), 1.5},
				},
			}},
		},
		{
			name: "float with units",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(time.Second),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), execute.Time(4 * time.Second), execute.Time(1 * time.Second), 2.0},
					{execute.Time(1 * time.Second), execute.Time(4 * time.Second), execute.Time(3 * time.Second), 1.0},
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
					{execute.Time(1 * time.Second), execute.Time(4 * time.Second), 3.0},
				},
			}},
		},
		{
			name: "float with tags",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), 2.0, "a"},
					{execute.Time(1), execute.Time(3), execute.Time(2), 1.0, "b"},
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
					{execute.Time(1), execute.Time(3), 1.5},
				},
			}},
		},
		{
			name: "float with multiple values",
			spec: &universe.IntegralProcedureSpec{
				Unit:       flux.ConvertDuration(1),
				TimeColumn: execute.DefaultTimeColLabel,
				AggregateConfig: execute.AggregateConfig{
					Columns: []string{"x", "y"},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(5), execute.Time(1), 2.0, 20.0},
					{execute.Time(1), execute.Time(5), execute.Time(2), 1.0, 10.0},
					{execute.Time(1), execute.Time(5), execute.Time(3), 2.0, 20.0},
					{execute.Time(1), execute.Time(5), execute.Time(4), 1.0, 10.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(5), 4.5, 45.0},
				},
			}},
		},
		{
			name: "float with null timestamps",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(4), execute.Time(1), 2.0},
					{execute.Time(1), execute.Time(4), nil, 3.0},
					{execute.Time(1), execute.Time(4), nil, 1.0},
					{execute.Time(1), execute.Time(4), execute.Time(3), nil},
				},
			}},
			wantErr: fmt.Errorf("integral found null time in time column"),
		},
		{
			name: "float with null values",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(6), execute.Time(1), 2.0},
					{execute.Time(1), execute.Time(6), execute.Time(2), 3.0},
					{execute.Time(1), execute.Time(6), execute.Time(3), 1.0},
					{execute.Time(1), execute.Time(6), execute.Time(4), nil},
					{execute.Time(1), execute.Time(6), execute.Time(5), 4.0},
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
					{execute.Time(1), execute.Time(6), 9.5},
				},
			}},
		},
		{
			name: "float with out-of-order timestamps",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(5), execute.Time(1), 2.0},
					{execute.Time(1), execute.Time(5), execute.Time(4), 5.0},
					{execute.Time(1), execute.Time(5), execute.Time(3), 1.0},
					{execute.Time(1), execute.Time(5), execute.Time(2), 3.0},
				},
			}},
			wantErr: fmt.Errorf("integral found out-of-order times in time column"),
		},
		{
			name: "integral over string",
			spec: &universe.IntegralProcedureSpec{
				Unit:       flux.ConvertDuration(1),
				TimeColumn: execute.DefaultTimeColLabel,
				AggregateConfig: execute.AggregateConfig{
					Columns: []string{"t"},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), 2.0, "a"},
					{execute.Time(1), execute.Time(3), execute.Time(2), 1.0, "b"},
				},
			}},
			wantErr: fmt.Errorf("cannot perform integral over string"),
		},
		{
			name: "float repeated times",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(3), execute.Time(1), 2.0},
					{execute.Time(1), execute.Time(3), execute.Time(1), 32.0},
					{execute.Time(1), execute.Time(3), execute.Time(2), 1.0},
					{execute.Time(1), execute.Time(3), execute.Time(2), 42.0},
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
					{execute.Time(1), execute.Time(3), 1.5},
				},
			}},
		},
		{
			name: "interpolate0",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(1), 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 2.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 3.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 4.0},
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
					{execute.Time(0), execute.Time(5), 12.5},
				},
			}},
		},
		{
			name: "interpolate1",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(10), execute.Time(2), 4.0},
					{execute.Time(0), execute.Time(10), execute.Time(4), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(6), 2.0},
					{execute.Time(0), execute.Time(10), execute.Time(8), 1.0},
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
					{execute.Time(0), execute.Time(10), 25.0},
				},
			}},
		},
		{
			name: "interpolate2",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(6), execute.Time(0), 1.0},
					{execute.Time(0), execute.Time(6), execute.Time(1), 2.0},
					{execute.Time(0), execute.Time(6), execute.Time(2), 4.0},
					{execute.Time(0), execute.Time(6), execute.Time(3), 6.0},
					{execute.Time(0), execute.Time(6), execute.Time(4), 8.0},
					{execute.Time(0), execute.Time(6), execute.Time(5), 10.0},
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
					{execute.Time(0), execute.Time(6), 36.5},
				},
			}},
		},
		{
			name: "interpolate3",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(6), execute.Time(0), 1.0},
					{execute.Time(0), execute.Time(6), execute.Time(1), 1.0},
					{execute.Time(0), execute.Time(6), execute.Time(2), 1.0},
					{execute.Time(0), execute.Time(6), execute.Time(3), 0.0},
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
					{execute.Time(0), execute.Time(6), -2.0},
				},
			}},
		},
		{
			name: "interpolate skip start",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(10), execute.Time(0), 2.0},
					{execute.Time(0), execute.Time(10), execute.Time(2), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(4), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(6), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(8), 3.0},
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
					{execute.Time(0), execute.Time(10), 29.0},
				},
			}},
		},
		{
			name: "start and stop not in group key",
			spec: &universe.IntegralProcedureSpec{
				Unit:            flux.ConvertDuration(1),
				TimeColumn:      execute.DefaultTimeColLabel,
				Interpolate:     true,
				AggregateConfig: execute.DefaultAggregateConfig,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(10), execute.Time(0), 2.0},
					{execute.Time(0), execute.Time(10), execute.Time(2), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(4), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(6), 3.0},
					{execute.Time(0), execute.Time(10), execute.Time(8), 3.0},
				},
			}},
			wantErr: fmt.Errorf("integral function needs _start column to be part of group key"),
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
					return universe.NewIntegralTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
