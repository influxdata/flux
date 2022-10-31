package universe_test

import (
	"context"
	"math"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

const epsilon float64 = 1e-5

var floatOptions = cmp.Options{
	cmpopts.EquateApprox(0, epsilon),
}

func TestHoltWinters_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "holt winters defaults",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> holtWinters(n: 84, interval: 42d)`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
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
						ID: "holtWinters2",
						Spec: &universe.HoltWintersOpSpec{
							WithFit:    false,
							Column:     execute.DefaultValueColLabel,
							TimeColumn: execute.DefaultTimeColLabel,
							N:          84,
							S:          0,
							Interval:   flux.ConvertDuration(42 * 24 * time.Hour),
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "holtWinters2"},
				},
			},
		},
		{
			Name: "holt winters no defaults",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> holtWinters(n: 84, seasonality: 4, interval: 42d, timeColumn: "t", column: "v", withFit: true)`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
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
						ID: "holtWinters2",
						Spec: &universe.HoltWintersOpSpec{
							WithFit:    true,
							Column:     "v",
							TimeColumn: "t",
							N:          84,
							S:          4,
							Interval:   flux.ConvertDuration(42 * 24 * time.Hour),
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "holtWinters2"},
				},
			},
		},
		{
			Name:    "holt winters blank",
			Raw:     `from(bucket:"mydb") |> range(start:-1h) |> holtWinters()`,
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

func TestHoltWinters_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewHoltWintersTransformation(
			d,
			c,
			&memory.ResourceAllocator{},
			&universe.HoltWintersProcedureSpec{},
		)
		return s
	})
}

// The data for these tests is taken from the InfluxQL example:
// http://docs.influxdata.com/influxdb/v1.7/query_language/functions/#holt-winters.
// To obtain the data employed in tests, we took the NOAA waterhouse database,
// imported it in an instance of InfluxDB 1.7, and ran the queries described in the example.
// The initial data used in tests is obtained from the original (large) dataset with this query:
// ```
// SELECT FIRST("water_level") into "first"."autogen"."data"
//
//	FROM "water"."autogen"."h2o_feet"
//	WHERE "location"='santa_monica' and time >= '2015-08-22 22:12:00' and time <= '2015-08-28 03:00:00'
//	GROUP BY time(379m,348m)
//
// ```
// HoltWinters is then calculated on the database "first":
// ```
// SELECT holt_winters(max("first"), 10, 4)
//
//	from "first"."autogen"."data"
//	GROUP BY time(379m,348m)
//
// ```
// We followed a similar procedure for other tests with missing values.
func TestHoltWinters_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.HoltWintersProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "NOAA water - no fit",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000)},
						{2.192, execute.Time(1440304260000000000)},
						{3.035, execute.Time(1440327000000000000)},
						{2.93, execute.Time(1440349740000000000)},
						{5.121, execute.Time(1440372480000000000)},
						{1.722, execute.Time(1440395220000000000)},
						{3.209, execute.Time(1440417960000000000)},
						{2.877, execute.Time(1440440700000000000)},
						{5.449, execute.Time(1440463440000000000)},
						{0.896, execute.Time(1440486180000000000)},
						{3.655, execute.Time(1440508920000000000)},
						{2.71, execute.Time(1440531660000000000)},
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{4.357, execute.Time(1440599880000000000)},
						{2.618, execute.Time(1440622620000000000)},
						{6.102, execute.Time(1440645360000000000)},
						{0.072, execute.Time(1440668100000000000)},
						{4.816, execute.Time(1440690840000000000)},
						{2.612, execute.Time(1440713580000000000)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
					{execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
					{execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
					{execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
					{execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
					{execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
					{execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
					{execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
					{execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
					{execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
				},
			}},
		},
		{
			name: "NOAA water - with fit",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    true,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000)},
						{2.192, execute.Time(1440304260000000000)},
						{3.035, execute.Time(1440327000000000000)},
						{2.93, execute.Time(1440349740000000000)},
						{5.121, execute.Time(1440372480000000000)},
						{1.722, execute.Time(1440395220000000000)},
						{3.209, execute.Time(1440417960000000000)},
						{2.877, execute.Time(1440440700000000000)},
						{5.449, execute.Time(1440463440000000000)},
						{0.896, execute.Time(1440486180000000000)},
						{3.655, execute.Time(1440508920000000000)},
						{2.71, execute.Time(1440531660000000000)},
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{4.357, execute.Time(1440599880000000000)},
						{2.618, execute.Time(1440622620000000000)},
						{6.102, execute.Time(1440645360000000000)},
						{0.072, execute.Time(1440668100000000000)},
						{4.816, execute.Time(1440690840000000000)},
						{2.612, execute.Time(1440713580000000000)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440281520000000000), 4.948, 3.9637130993181717},
					{execute.Time(1440304260000000000), 1.9838639270131349, 3.9637130993181717},
					{execute.Time(1440327000000000000), 3.1642193518149218, 3.9637130993181717},
					{execute.Time(1440349740000000000), 2.245682698032771, 3.9637130993181717},
					{execute.Time(1440372480000000000), 5.1920419019025905, 3.9637130993181717},
					{execute.Time(1440395220000000000), 0.8468035647406211, 3.9637130993181717},
					{execute.Time(1440417960000000000), 3.9260207208633666, 3.9637130993181717},
					{execute.Time(1440440700000000000), 2.8208710933980967, 3.9637130993181717},
					{execute.Time(1440463440000000000), 5.803055089581454, 3.9637130993181717},
					{execute.Time(1440486180000000000), 0.888718952477037, 3.9637130993181717},
					{execute.Time(1440508920000000000), 4.006733423268878, 3.9637130993181717},
					{execute.Time(1440531660000000000), 2.8449068276505463, 3.9637130993181717},
					{execute.Time(1440554400000000000), 5.823540425832358, 3.9637130993181717},
					{execute.Time(1440577140000000000), 0.8900185986232847, 3.9637130993181717},
					{execute.Time(1440599880000000000), 4.009160677724419, 3.9637130993181717},
					{execute.Time(1440622620000000000), 2.8456207547297403, 3.9637130993181717},
					{execute.Time(1440645360000000000), 5.82414581225715, 3.9637130993181717},
					{execute.Time(1440668100000000000), 0.8900569255488864, 3.9637130993181717},
					{execute.Time(1440690840000000000), 4.00923219607614, 3.9637130993181717},
					{execute.Time(1440713580000000000), 2.8456417828335208, 3.9637130993181717},
					{execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
					{execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
					{execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
					{execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
					{execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
					{execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
					{execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
					{execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
					{execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
					{execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
				},
			}},
		},
		{
			name: "null times get skipped",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{0.00000000042, nil}, // should be skipped
						{4.948, execute.Time(1440281520000000000)},
						{0.00000000042, nil}, // should be skipped
						{2.192, execute.Time(1440304260000000000)},
						{0.00000000042, nil}, // should be skipped
						{3.035, execute.Time(1440327000000000000)},
						{0.00000000042, nil}, // should be skipped
						{2.93, execute.Time(1440349740000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{5.121, execute.Time(1440372480000000000)},
						{0.00000000042, nil}, // should be skipped
						{1.722, execute.Time(1440395220000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{3.209, execute.Time(1440417960000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{2.877, execute.Time(1440440700000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{5.449, execute.Time(1440463440000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.896, execute.Time(1440486180000000000)},
						{0.00000000042, nil}, // should be skipped
						{3.655, execute.Time(1440508920000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{2.71, execute.Time(1440531660000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{0.00000000042, nil}, // should be skipped
						{4.357, execute.Time(1440599880000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{2.618, execute.Time(1440622620000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{6.102, execute.Time(1440645360000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.072, execute.Time(1440668100000000000)},
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{0.00000000042, nil}, // should be skipped
						{4.816, execute.Time(1440690840000000000)},
						{0.00000000042, nil}, // should be skipped
						{2.612, execute.Time(1440713580000000000)},
						{0.00000000042, nil}, // should be skipped
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
					{execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
					{execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
					{execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
					{execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
					{execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
					{execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
					{execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
					{execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
					{execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
				},
			}},
		},
		{
			name: "first invalid values gets ignored",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{nil, execute.Time(1440281520000000000)},
						{nil, execute.Time(1440281520000000000)},
						{nil, execute.Time(1440281520000000000)},
						{4.948, execute.Time(1440281520000000000)},
						{2.192, execute.Time(1440304260000000000)},
						{3.035, execute.Time(1440327000000000000)},
						{2.93, execute.Time(1440349740000000000)},
						{5.121, execute.Time(1440372480000000000)},
						{1.722, execute.Time(1440395220000000000)},
						{3.209, execute.Time(1440417960000000000)},
						{2.877, execute.Time(1440440700000000000)},
						{5.449, execute.Time(1440463440000000000)},
						{0.896, execute.Time(1440486180000000000)},
						{3.655, execute.Time(1440508920000000000)},
						{2.71, execute.Time(1440531660000000000)},
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{4.357, execute.Time(1440599880000000000)},
						{2.618, execute.Time(1440622620000000000)},
						{6.102, execute.Time(1440645360000000000)},
						{0.072, execute.Time(1440668100000000000)},
						{4.816, execute.Time(1440690840000000000)},
						{2.612, execute.Time(1440713580000000000)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
					{execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
					{execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
					{execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
					{execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
					{execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
					{execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
					{execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
					{execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
					{execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
				},
			}},
		},
		{
			name: "only the first value per bucket is considered",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000)},
						{4.948, execute.Time(1440281520000000000 + 1)},
						{4.948, execute.Time(1440281520000000000 + 2)},
						{2.192, execute.Time(1440304260000000000)},
						{2.192, execute.Time(1440304260000000000 + 0)},
						{3.035, execute.Time(1440327000000000000)},
						{2.93, execute.Time(1440349740000000000)},
						{5.121, execute.Time(1440372480000000000)},
						{5.121, execute.Time(1440372480000000000 + 0)},
						{5.121, execute.Time(1440372480000000000 + 1)},
						{5.121, execute.Time(1440372480000000000 + 2)},
						{5.121, execute.Time(1440372480000000000 + 3)},
						{1.722, execute.Time(1440395220000000000)},
						{3.209, execute.Time(1440417960000000000)},
						{2.877, execute.Time(1440440700000000000)},
						{5.449, execute.Time(1440463440000000000)},
						{0.896, execute.Time(1440486180000000000 - 3)},
						{0.896, execute.Time(1440486180000000000 - 2)},
						{0.896, execute.Time(1440486180000000000 - 1)},
						{0.896, execute.Time(1440486180000000000)},
						{3.655, execute.Time(1440508920000000000)},
						{2.71, execute.Time(1440531660000000000)},
						{2.71, execute.Time(1440531660000000000 + 0)},
						{2.71, execute.Time(1440531660000000000 + 0)},
						{2.71, execute.Time(1440531660000000000 + 0)},
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{4.357, execute.Time(1440599880000000000)},
						{2.618, execute.Time(1440622620000000000)},
						{6.102, execute.Time(1440645360000000000)},
						{0.072, execute.Time(1440668100000000000)},
						{4.816, execute.Time(1440690840000000000)},
						{2.612, execute.Time(1440713580000000000)},
						{2.612, execute.Time(1440713580000000000 + 1)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
					{execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
					{execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
					{execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
					{execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
					{execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
					{execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
					{execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
					{execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
					{execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
				},
			}},
		},
		{
			name: "NOAA water - with tag keys",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag_string", "tag_int", "tag_uint", "tag_float", "tag_bool", "tag_time"},
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
						{Label: "tag_string", Type: flux.TString},
						{Label: "tag_int", Type: flux.TInt},
						{Label: "tag_uint", Type: flux.TUInt},
						{Label: "tag_float", Type: flux.TFloat},
						{Label: "tag_bool", Type: flux.TBool},
						{Label: "tag_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.192, execute.Time(1440304260000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{3.035, execute.Time(1440327000000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.93, execute.Time(1440349740000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{5.121, execute.Time(1440372480000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{1.722, execute.Time(1440395220000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{3.209, execute.Time(1440417960000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.877, execute.Time(1440440700000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{5.449, execute.Time(1440463440000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{0.896, execute.Time(1440486180000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{3.655, execute.Time(1440508920000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.71, execute.Time(1440531660000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{5.961, execute.Time(1440554400000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{0.404, execute.Time(1440577140000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{4.357, execute.Time(1440599880000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.618, execute.Time(1440622620000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{6.102, execute.Time(1440645360000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{0.072, execute.Time(1440668100000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{4.816, execute.Time(1440690840000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
						{2.612, execute.Time(1440713580000000000), "t", int64(0), uint64(0), 0.0, true, execute.Time(0)},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag_string", "tag_int", "tag_uint", "tag_float", "tag_bool", "tag_time"},
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
						{Label: "tag_string", Type: flux.TString},
						{Label: "tag_int", Type: flux.TInt},
						{Label: "tag_uint", Type: flux.TUInt},
						{Label: "tag_float", Type: flux.TFloat},
						{Label: "tag_bool", Type: flux.TBool},
						{Label: "tag_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.192, execute.Time(1440304260000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{3.035, execute.Time(1440327000000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.93, execute.Time(1440349740000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{5.121, execute.Time(1440372480000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{1.722, execute.Time(1440395220000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{3.209, execute.Time(1440417960000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.877, execute.Time(1440440700000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{5.449, execute.Time(1440463440000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{0.896, execute.Time(1440486180000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{3.655, execute.Time(1440508920000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.71, execute.Time(1440531660000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{5.961, execute.Time(1440554400000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{0.404, execute.Time(1440577140000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{4.357, execute.Time(1440599880000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.618, execute.Time(1440622620000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{6.102, execute.Time(1440645360000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{0.072, execute.Time(1440668100000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{4.816, execute.Time(1440690840000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
						{2.612, execute.Time(1440713580000000000), "t", int64(0), uint64(0), 0.0, false, execute.Time(0)},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag_string", "tag_int", "tag_uint", "tag_float", "tag_bool", "tag_time"},
					ColMeta: []flux.ColMeta{
						{Label: "tag_string", Type: flux.TString},
						{Label: "tag_int", Type: flux.TInt},
						{Label: "tag_uint", Type: flux.TUInt},
						{Label: "tag_float", Type: flux.TFloat},
						{Label: "tag_bool", Type: flux.TBool},
						{Label: "tag_time", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "minSSE", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
					},
				},
				{
					KeyCols: []string{"tag_string", "tag_int", "tag_uint", "tag_float", "tag_bool", "tag_time"},
					ColMeta: []flux.ColMeta{
						{Label: "tag_string", Type: flux.TString},
						{Label: "tag_int", Type: flux.TInt},
						{Label: "tag_uint", Type: flux.TUInt},
						{Label: "tag_float", Type: flux.TFloat},
						{Label: "tag_bool", Type: flux.TBool},
						{Label: "tag_time", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "minSSE", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440736320000000000), 5.824163640738761, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440759060000000000), 0.8900580541979224, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440781800000000000), 4.009234302090598, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440804540000000000), 2.845642402045534, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440827280000000000), 5.8241641657295915, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440850020000000000), 0.8900580874329075, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440872760000000000), 4.009234364105719, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440895500000000000), 2.8456424202792623, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440918240000000000), 5.824164181188819, 3.9637130993181717},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440940980000000000), 0.8900580884115666, 3.9637130993181717},
					},
				},
			},
		},
		// This test was crafted by removing some points and adding null ones from the original dataset.
		// The expected result is calculated by performing the query in InfluxQL.
		{
			name: "NOAA water - with nulls and missing",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1440281520000000000), 4.948},
						{execute.Time(1440281520000000001), nil},
						{execute.Time(1440281520000000002), nil},
						{execute.Time(1440304260000000000), 2.192},
						// missing point for 1440327000000000000
						{execute.Time(1440349740000000000), 2.93},
						{execute.Time(1440372480000000000), 5.121},
						// missing point for 1440395220000000000
						{execute.Time(1440395220000000001), nil},
						// missing point for 1440417960000000000
						{execute.Time(1440440700000000000), 2.877},
						{execute.Time(1440463440000000000), 5.449},
						{execute.Time(1440486180000000000), 0.896},
						{execute.Time(1440486180000000001), nil},
						{execute.Time(1440486180000000002), nil},
						{execute.Time(1440486180000000003), nil},
						// missing point for 1440508920000000000
						// missing point for 1440531660000000000
						// missing point for 1440554400000000000
						{execute.Time(1440577140000000000), 0.404},
						{execute.Time(1440599880000000000), 4.357},
						{execute.Time(1440599880000000001), nil},
						{execute.Time(1440622620000000000), 2.618},
						{execute.Time(1440645360000000000), 6.102},
						// missing point for 1440668100000000000
						{execute.Time(1440690840000000000), 4.816},
						// missing point for 1440713580000000000
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440713580000000000), 2.856048785717742, 0.4082862899942894},
					{execute.Time(1440736320000000000), 5.916469115638421, 0.4082862899942894},
					{execute.Time(1440759060000000000), 0.6319256010521764, 0.4082862899942894},
					{execute.Time(1440781800000000000), 4.594962598842862, 0.4082862899942894},
					{execute.Time(1440804540000000000), 2.856791807223551, 0.4082862899942894},
					{execute.Time(1440827280000000000), 5.917429252436063, 0.4082862899942894},
					{execute.Time(1440850020000000000), 0.6319895674111194, 0.4082862899942894},
					{execute.Time(1440872760000000000), 4.595252714139058, 0.4082862899942894},
					{execute.Time(1440895500000000000), 2.8569043098337357, 0.4082862899942894},
					{execute.Time(1440918240000000000), 5.917574599980711, 0.4082862899942894},
				},
			}},
		},
		// The expected result is calculated by executing the query in InfluxQL with no seasonality.
		{
			name: "NOAA water - no seasonal",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          0,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "_stop", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{4.948, execute.Time(1440281520000000000)},
						{2.192, execute.Time(1440304260000000000)},
						{3.035, execute.Time(1440327000000000000)},
						{2.93, execute.Time(1440349740000000000)},
						{5.121, execute.Time(1440372480000000000)},
						{1.722, execute.Time(1440395220000000000)},
						{3.209, execute.Time(1440417960000000000)},
						{2.877, execute.Time(1440440700000000000)},
						{5.449, execute.Time(1440463440000000000)},
						{0.896, execute.Time(1440486180000000000)},
						{3.655, execute.Time(1440508920000000000)},
						{2.71, execute.Time(1440531660000000000)},
						{5.961, execute.Time(1440554400000000000)},
						{0.404, execute.Time(1440577140000000000)},
						{4.357, execute.Time(1440599880000000000)},
						{2.618, execute.Time(1440622620000000000)},
						{6.102, execute.Time(1440645360000000000)},
						{0.072, execute.Time(1440668100000000000)},
						{4.816, execute.Time(1440690840000000000)},
						{2.612, execute.Time(1440713580000000000)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.243467293744046, 55.00975192379251},
					{execute.Time(1440759060000000000), 3.2434611181992774, 55.00975192379251},
					{execute.Time(1440781800000000000), 3.243457659915101, 55.00975192379251},
					{execute.Time(1440804540000000000), 3.243455723309489, 55.00975192379251},
					{execute.Time(1440827280000000000), 3.243454638835965, 55.00975192379251},
					{execute.Time(1440850020000000000), 3.2434540315472833, 55.00975192379251},
					{execute.Time(1440872760000000000), 3.2434536914755303, 55.00975192379251},
					{execute.Time(1440895500000000000), 3.2434535010411083, 55.00975192379251},
					{execute.Time(1440918240000000000), 3.2434533944011235, 55.00975192379251},
					{execute.Time(1440940980000000000), 3.2434533346845957, 55.00975192379251},
				},
			}},
		},
		// The expected result is calculated by executing the query in InfluxQL with no seasonality.
		{
			name: "NOAA water - with nulls and missing - no seasonal",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          0,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1440281520000000000), 4.948},
						{execute.Time(1440281520000000001), nil},
						{execute.Time(1440281520000000002), nil},
						{execute.Time(1440304260000000000), 2.192},
						// missing point for 1440327000000000000
						{execute.Time(1440349740000000000), 2.93},
						{execute.Time(1440372480000000000), 5.121},
						// missing point for 1440395220000000000
						{execute.Time(1440395220000000001), nil},
						// missing point for 1440417960000000000
						{execute.Time(1440440700000000000), 2.877},
						{execute.Time(1440463440000000000), 5.449},
						{execute.Time(1440486180000000000), 0.896},
						{execute.Time(1440486180000000001), nil},
						{execute.Time(1440486180000000002), nil},
						{execute.Time(1440486180000000003), nil},
						// missing point for 1440508920000000000
						// missing point for 1440531660000000000
						// missing point for 1440554400000000000
						{execute.Time(1440577140000000000), 0.404},
						{execute.Time(1440599880000000000), 4.357},
						{execute.Time(1440599880000000001), nil},
						{execute.Time(1440622620000000000), 2.618},
						{execute.Time(1440645360000000000), 6.102},
						// missing point for 1440668100000000000
						{execute.Time(1440690840000000000), 4.816},
						// missing point for 1440713580000000000
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1440713580000000000), 6.515949253684278, 30.53516643428765},
					{execute.Time(1440736320000000000), 8.457963501700332, 30.53516643428765},
					{execute.Time(1440759060000000000), 11.484286385118807, 30.53516643428765},
					{execute.Time(1440781800000000000), 16.201176132268735, 30.53516643428765},
					{execute.Time(1440804540000000000), 23.553968786025475, 30.53516643428765},
					{execute.Time(1440827280000000000), 35.016736978634626, 30.53516643428765},
					{execute.Time(1440850020000000000), 52.88803362131611, 30.53516643428765},
					{execute.Time(1440872760000000000), 80.7520589326412, 30.53516643428765},
					{execute.Time(1440895500000000000), 124.1977801083343, 30.53516643428765},
					{execute.Time(1440918240000000000), 191.9402890971891, 30.53516643428765},
				},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			alloc := &memory.ResourceAllocator{}
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewHoltWintersTransformation(d, c, alloc, tc.spec)
				},
				floatOptions,
			)

			for i := 0; i < 30; i++ {
				runtime.GC()
				if alloc.Allocated() <= 0 {
					break
				}
			}

			if m := alloc.Allocated(); m != 0 {
				t.Errorf("HoltWinters is using memory after finishing: %d", m)
			}
		})
	}
}

func TestHoltWinters_Error_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.HoltWintersProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr string
	}{
		{
			name: "NaN in the input",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1440281520000000000), math.NaN()},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{},
			}},
		},
		{
			name: "Inf in the input",
			spec: &universe.HoltWintersProcedureSpec{
				Column:     "_value",
				TimeColumn: "_stop",
				WithFit:    false,
				N:          10,
				S:          4,
				Interval:   flux.ConvertDuration(379 * time.Minute),
				WithMinSSE: true,
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1440281520000000000), math.Inf(1)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "minSSE", Type: flux.TFloat},
				},
				Data: [][]interface{}{},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			wantErr := errors.New(codes.Invalid, "NaN/Inf in input")
			alloc := &memory.ResourceAllocator{}
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewHoltWintersTransformation(d, c, alloc, tc.spec)
				},
				floatOptions,
			)

			for i := 0; i < 30; i++ {
				runtime.GC()
				if alloc.Allocated() <= 0 {
					break
				}
			}

			if m := alloc.Allocated(); m != 0 {
				t.Errorf("HoltWinters is using memory after finishing: %d", m)
			}
		})
	}
}

func BenchmarkHoltWintersWithoutFit(b *testing.B) {
	benchmarkHoltWinters(b, 1000, 0, false)
}

func BenchmarkHoltWintersWithFit(b *testing.B) {
	benchmarkHoltWinters(b, 1000, 0, true)
}

func BenchmarkHoltWintersWithoutFitSeasonality(b *testing.B) {
	benchmarkHoltWinters(b, 1000, 4, false)
}

func BenchmarkHoltWintersWithFitSeasonality(b *testing.B) {
	benchmarkHoltWinters(b, 1000, 4, true)
}

func benchmarkHoltWinters(b *testing.B, n, seasonality int, withFit bool) {
	b.ReportAllocs()
	seed := int64(1234)
	spec := &universe.HoltWintersProcedureSpec{
		Column:     "_value",
		TimeColumn: "_time",
		WithFit:    withFit,
		N:          10,
		S:          int64(seasonality),
		Interval:   flux.ConvertDuration(379 * time.Minute),
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: n,
				Alloc:     alloc,
				Seed:      &seed,
			}
			return gen.Input(context.Background(), schema)
		},
		func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
			cache := execute.NewTableBuilderCache(alloc)
			d := execute.NewDataset(id, execute.DiscardingMode, cache)
			t := universe.NewHoltWintersTransformation(d, cache, alloc, spec)
			return t, d
		},
	)
}
