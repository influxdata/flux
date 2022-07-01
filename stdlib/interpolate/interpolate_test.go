package interpolate_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/interpolate"
	"github.com/influxdata/flux/values"
)

func TestLinearInterpolate(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *interpolate.LinearInterpolateProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "basic0",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(9), 9.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(5), 5.0},
					{execute.Time(9), 9.0},
				},
			}},
		},
		{
			name: "basic1",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(6), 6.0},
					{execute.Time(7), 7.0},
					{execute.Time(8), 8.0},
					{execute.Time(9), 9.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(5), 5.0},
					{execute.Time(6), 6.0},
					{execute.Time(7), 7.0},
					{execute.Time(8), 8.0},
					{execute.Time(9), 9.0},
				},
			}},
		},
		{
			name: "basic2",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_field"},
				ColMeta: []flux.ColMeta{
					{Label: "_field", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"a", execute.Time(1), 1.0},
					{"a", execute.Time(9), 9.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_field"},
				ColMeta: []flux.ColMeta{
					{Label: "_field", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"a", execute.Time(1), 1.0},
					{"a", execute.Time(5), 5.0},
					{"a", execute.Time(9), 9.0},
				},
			}},
		},
		{
			name: "group key error",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_field", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"a", execute.Time(1), 1.0},
					{"b", execute.Time(9), 9.0},
				},
			}},
			wantErr: fmt.Errorf("interpolate.linear requires column \"_field\" to be in group key"),
		},
		{
			name: "ints",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(1)},
					{execute.Time(9), int64(2)},
				},
			}},
			wantErr: fmt.Errorf("cannot interpolate int values; expected float values"),
		},
		{
			name: "nulls",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(5), nil},
					{execute.Time(9), 9.0},
				},
			}},
			wantErr: fmt.Errorf("null _value found during linear interpolation"),
		},
		{
			name: "no extrapolation",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), 0.0},
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(6), 6.0},
					{execute.Time(7), 7.0},
					{execute.Time(8), 8.0},
					{execute.Time(9), 9.0},
					{execute.Time(10), 10.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), 0.0},
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(5), 5.0},
					{execute.Time(6), 6.0},
					{execute.Time(7), 7.0},
					{execute.Time(8), 8.0},
					{execute.Time(9), 9.0},
					{execute.Time(10), 10.0},
				},
			}},
		},
		{
			name: "empty periods",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(10 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), 11.0},
					{execute.Time(99), 99.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), 11.0},
					{execute.Time(20), 20.0},
					{execute.Time(30), 30.0},
					{execute.Time(40), 40.0},
					{execute.Time(50), 50.0},
					{execute.Time(60), 60.0},
					{execute.Time(70), 70.0},
					{execute.Time(80), 80.0},
					{execute.Time(90), 90.0},
					{execute.Time(99), 99.0},
				},
			}},
		},
		{
			name: "no points",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
			}},
		},
		{
			name: "one point",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(5 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 3.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 3.0},
				},
			}},
		},
		{
			name: "identity",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: flux.ConvertDuration(10 * time.Nanosecond),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(20), 20.0},
					{execute.Time(25), 25.0},
					{execute.Time(30), 30.0},
					{execute.Time(35), 35.0},
					{execute.Time(40), 40.0},
					{execute.Time(45), 45.0},
					{execute.Time(50), 50.0},
					{execute.Time(55), 55.0},
					{execute.Time(60), 60.0},
					{execute.Time(65), 65.0},
					{execute.Time(70), 70.0},
					{execute.Time(75), 75.0},
					{execute.Time(80), 80.0},
					{execute.Time(85), 85.0},
					{execute.Time(90), 90.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(20), 20.0},
					{execute.Time(25), 25.0},
					{execute.Time(30), 30.0},
					{execute.Time(35), 35.0},
					{execute.Time(40), 40.0},
					{execute.Time(45), 45.0},
					{execute.Time(50), 50.0},
					{execute.Time(55), 55.0},
					{execute.Time(60), 60.0},
					{execute.Time(65), 65.0},
					{execute.Time(70), 70.0},
					{execute.Time(75), 75.0},
					{execute.Time(80), 80.0},
					{execute.Time(85), 85.0},
					{execute.Time(90), 90.0},
				},
			}},
		},
		{
			name: "calendar duration",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: func() values.Duration {
					d, _ := values.ParseDuration("3mo")
					return d
				}(),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{func() execute.Time {
						t, _ := values.ParseTime("2014-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2015-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{func() execute.Time {
						t, _ := values.ParseTime("2014-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-04-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-07-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-10-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2015-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
				},
			}},
		},
		{
			name: "calendar duration",
			spec: &interpolate.LinearInterpolateProcedureSpec{
				Every: func() values.Duration {
					d, _ := values.ParseDuration("1mo")
					return d
				}(),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{func() execute.Time {
						t, _ := values.ParseTime("2014-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-04-01T00:00:00.000000000Z")
						return t
					}(), 91.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{func() execute.Time {
						t, _ := values.ParseTime("2014-01-01T00:00:00.000000000Z")
						return t
					}(), 1.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-02-01T00:00:00.000000000Z")
						return t
					}(), 32.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-03-01T00:00:00.000000000Z")
						return t
					}(), 60.0},
					{func() execute.Time {
						t, _ := values.ParseTime("2014-04-01T00:00:00.000000000Z")
						return t
					}(), 91.0},
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
					return interpolate.NewInterpolateTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
