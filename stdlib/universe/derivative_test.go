package universe_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestDerivativeOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"derivative","kind":"derivative","spec":{"unit":"1m","nonNegative":true}}`)
	op := &flux.Operation{
		ID: "derivative",
		Spec: &universe.DerivativeOpSpec{
			Unit:        flux.ConvertDuration(time.Minute),
			NonNegative: true,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestDerivative_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewDerivativeTransformation(
			d,
			c,
			&universe.DerivativeProcedureSpec{},
		)
		return s
	})
}

func TestDerivative_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.DerivativeProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "float",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
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
					{execute.Time(2), -1.0},
				},
			}},
		},
		{
			name: "float with units",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(time.Second),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), 2.0},
					{execute.Time(3 * time.Second), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3 * time.Second), -0.5},
				},
			}},
		},
		{
			name: "float with tags",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, "a"},
					{execute.Time(2), 1.0, "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "float with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, 20.0},
					{execute.Time(2), 1.0, 10.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, -10.0},
				},
			}},
		},
		{
			name: "float non negative with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{"x", "y"},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, 20.0},
					{execute.Time(2), 1.0, 10.0},
					{execute.Time(3), 2.0, 0.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 1.0, nil},
				},
			}},
		},
		{
			name: "float with null values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, nil},
					{execute.Time(2), nil, 10.0},
					{execute.Time(3), 8.0, 20.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 3.0, 10.0},
				},
			}},
		},
		{
			name: "float rowwise",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0},
				},
			}},
		},
		{
			name: "float with passthrough",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, 20.0},
					{execute.Time(2), 1.0, 10.0},
					{execute.Time(3), 1.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, 10.0},
					{execute.Time(3), 0.0, nil},
				},
			}},
		},
		{
			name: "int",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(20)},
					{execute.Time(2), int64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -10.0},
				},
			}},
		},
		{
			name: "int with units",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(time.Second),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), int64(20)},
					{execute.Time(3 * time.Second), int64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3 * time.Second), -5.0},
				},
			}},
		},
		{
			name: "int non negative",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(20)},
					{execute.Time(2), int64(10)},
					{execute.Time(3), int64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), 10.0},
				},
			}},
		},
		{
			name: "int with tags",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), "a"},
					{execute.Time(2), int64(1), "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "int with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), int64(20)},
					{execute.Time(2), int64(1), int64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, -10.0},
				},
			}},
		},
		{
			name: "int non negative with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{"x", "y"},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), int64(20)},
					{execute.Time(2), int64(1), int64(10)},
					{execute.Time(3), int64(2), int64(0)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 1.0, nil},
				},
			}},
		},
		{
			name: "int with null values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), nil},
					{execute.Time(2), nil, int64(10)},
					{execute.Time(3), int64(8), int64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 3.0, 10.0},
				},
			}},
		},
		{
			name: "int rowwise",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(20)},
						{execute.Time(2), int64(10)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -10.0},
				},
			}},
		},
		{
			name: "int with passthrough",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), int64(20)},
					{execute.Time(2), int64(1), int64(10)},
					{execute.Time(3), int64(1), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, int64(10)},
					{execute.Time(3), 0.0, nil},
				},
			}},
		},
		{
			name: "uint",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(10)},
					{execute.Time(2), uint64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 10.0},
				},
			}},
		},
		{
			name: "uint with negative result",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(20)},
					{execute.Time(2), uint64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -10.0},
				},
			}},
		},
		{
			name: "uint with non negative",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(20)},
					{execute.Time(2), uint64(10)},
					{execute.Time(3), uint64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), 10.0},
				},
			}},
		},
		{
			name: "uint with units",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(time.Second),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), uint64(20)},
					{execute.Time(3 * time.Second), uint64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3 * time.Second), -5.0},
				},
			}},
		},
		{
			name: "uint with tags",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), "a"},
					{execute.Time(2), uint64(1), "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "uint with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TUInt},
					{Label: "y", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), uint64(20)},
					{execute.Time(2), uint64(1), uint64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, -10.0},
				},
			}},
		},
		{
			name: "uint non negative with multiple values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{"x", "y"},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TUInt},
					{Label: "y", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), uint64(20)},
					{execute.Time(2), uint64(1), uint64(10)},
					{execute.Time(3), uint64(2), uint64(0)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 1.0, nil},
				},
			}},
		},
		{
			name: "uint with null values",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TUInt},
					{Label: "y", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), nil},
					{execute.Time(2), nil, uint64(10)},
					{execute.Time(3), uint64(8), uint64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 3.0, 10.0},
				},
			}},
		},
		{
			name: "uint rowwise",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(10)},
						{execute.Time(2), uint64(20)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 10.0},
				},
			}},
		},
		{
			name: "uint with passthrough",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TUInt},
					{Label: "y", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), uint64(20)},
					{execute.Time(2), uint64(1), uint64(10)},
					{execute.Time(3), uint64(1), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, uint64(10)},
					{execute.Time(3), 0.0, nil},
				},
			}},
		},
		{
			name: "non negative one table",
			spec: &universe.DerivativeProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				TimeColumn:  execute.DefaultTimeColLabel,
				Unit:        flux.ConvertDuration(1),
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 2.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), 1.0},
				},
			}},
		},
		{
			name: "nulls in time column",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, nil},
					{execute.Time(2), nil, 10.0},
					{nil, 8.0, 20.0},
					{execute.Time(4), 8.0, 20.0},
					{nil, 8.0, 20.0},
					{execute.Time(6), 10.0, 25.0},
					{nil, 8.0, 20.0},
				},
			}},
			wantErr: fmt.Errorf("derivative found null time in time column"),
		},
		{
			name: "times out of order",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x", "y"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, 10.0},
					{execute.Time(4), 8.0, 20.0},
					{execute.Time(6), 10.0, 25.0},

					{execute.Time(3), nil, 10.0},
					{execute.Time(5), 8.0, 20.0},
					{execute.Time(7), 10.0, nil},
				},
			}},
			wantErr: fmt.Errorf("derivative found out-of-order times in time column"),
		},
		{
			name: "pass through with repeated times",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{"x"},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "b", Type: flux.TBool},
					{Label: "s", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, false, "bar"},
					{execute.Time(2), 1.0, false, "bar"},
					{execute.Time(4), 8.0, false, nil},
					{execute.Time(4), 9.0, true, "baz"},
					{execute.Time(6), 10.0, nil, "dog"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "b", Type: flux.TBool},
					{Label: "s", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(4), nil, false, nil},
					{execute.Time(6), 1.0, nil, "dog"},
				},
			}},
		},
		{
			name: "string",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a"},
					{execute.Time(2), "b"},
				},
			}},
			wantErr: errors.New(codes.FailedPrecondition, "unsupported derivative column type _value:string"),
		},
		{
			name: "bool",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{execute.Time(1), true},
					{execute.Time(2), false},
				},
			}},
			wantErr: errors.New(codes.FailedPrecondition, "unsupported derivative column type _value:bool"),
		},
		{
			name: "time",
			spec: &universe.DerivativeProcedureSpec{
				Columns:    []string{execute.DefaultValueColLabel},
				TimeColumn: execute.DefaultTimeColLabel,
				Unit:       flux.ConvertDuration(1),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(1)},
					{execute.Time(2), execute.Time(2)},
				},
			}},
			wantErr: errors.New(codes.FailedPrecondition, "unsupported derivative column type _value:time"),
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
					return universe.NewDerivativeTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
