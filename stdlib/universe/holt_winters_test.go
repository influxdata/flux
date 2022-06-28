package universe_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestHoltWinters_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "holt winters defaults",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> holtWinters(n: 84, interval: 42d)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
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
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "holtWinters2"},
				},
			},
		},
		{
			Name: "holt winters no defaults",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> holtWinters(n: 84, seasonality: 4, interval: 42d, timeColumn: "t", column: "v", withFit: true)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
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
				Edges: []flux.Edge{
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

func TestHoltWinters_Marshaling(t *testing.T) {
	data := []byte(`{"id":"hw","kind":"holtWinters","spec":{"n":84,"s":4,"interval":"42m","time_column":"t","column":"v","with_fit":true,"with_minsse":true}}`)
	op := &flux.Operation{
		ID: "hw",
		Spec: &universe.HoltWintersOpSpec{
			WithFit:    true,
			Column:     "v",
			TimeColumn: "t",
			N:          84,
			S:          4,
			Interval:   flux.ConvertDuration(42 * time.Minute),
			WithMinSSE: true,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
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
// 	FROM "water"."autogen"."h2o_feet"
// 	WHERE "location"='santa_monica' and time >= '2015-08-22 22:12:00' and time <= '2015-08-28 03:00:00'
// 	GROUP BY time(379m,348m)
// ```
// HoltWinters is then calculated on the database "first":
// ```
// SELECT holt_winters(max("first"), 10, 4)
// 	from "first"."autogen"."data"
// 	GROUP BY time(379m,348m)
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
					{execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
					{execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
					{execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
					{execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
					{execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
					{execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
					{execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
					{execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
					{execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
					{execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
					{execute.Time(1440281520000000000), 4.948, 3.963712566228843},
					{execute.Time(1440304260000000000), 1.983675129748691, 3.963712566228843},
					{execute.Time(1440327000000000000), 3.1640361715606247, 3.963712566228843},
					{execute.Time(1440349740000000000), 2.2457174796999486, 3.963712566228843},
					{execute.Time(1440372480000000000), 5.1919956963843426, 3.963712566228843},
					{execute.Time(1440395220000000000), 0.8468568048222737, 3.963712566228843},
					{execute.Time(1440417960000000000), 3.9258243579739003, 3.963712566228843},
					{execute.Time(1440440700000000000), 2.820767836091717, 3.963712566228843},
					{execute.Time(1440463440000000000), 5.802807022732507, 3.963712566228843},
					{execute.Time(1440486180000000000), 0.8887593034239336, 3.963712566228843},
					{execute.Time(1440508920000000000), 4.006499161070992, 3.963712566228843},
					{execute.Time(1440531660000000000), 2.8447913944020473, 3.963712566228843},
					{execute.Time(1440554400000000000), 5.8232808086768015, 3.963712566228843},
					{execute.Time(1440577140000000000), 0.8900582644352708, 3.963712566228843},
					{execute.Time(1440599880000000000), 4.008924758785957, 3.963712566228843},
					{execute.Time(1440622620000000000), 2.8455048100875557, 3.963712566228843},
					{execute.Time(1440645360000000000), 5.823885727761851, 3.963712566228843},
					{execute.Time(1440668100000000000), 0.8900965632076626, 3.963712566228843},
					{execute.Time(1440690840000000000), 4.008996213518486, 3.963712566228843},
					{execute.Time(1440713580000000000), 2.84552581877966, 3.963712566228843},
					{execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
					{execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
					{execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
					{execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
					{execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
					{execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
					{execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
					{execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
					{execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
					{execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
					{execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
					{execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
					{execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
					{execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
					{execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
					{execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
					{execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
					{execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
					{execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
					{execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
					{execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
					{execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
					{execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
					{execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
					{execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
					{execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
					{execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
					{execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
					{execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
					{execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
					{execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
					{execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
					{execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
					{execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
					{execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
					{execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
					{execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
					{execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
					{execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
					{execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440736320000000000), 5.823903538794723, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440759060000000000), 0.8900976907943485, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440781800000000000), 4.008998317224315, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440804540000000000), 2.8455264372921047, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440827280000000000), 5.823904063163267, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440850020000000000), 0.8900977239911837, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440872760000000000), 4.008998379158643, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440895500000000000), 2.8455264555014668, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440918240000000000), 5.823904078600977, 3.963712566228843},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440940980000000000), 0.8900977249685175, 3.963712566228843},
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
					{execute.Time(1440713580000000000), 8.060871776244943, 4.308085710844578},
					{execute.Time(1440736320000000000), 17.347287691732163, 4.308085710844578},
					{execute.Time(1440759060000000000), 11.601811194225883, 4.308085710844578},
					{execute.Time(1440781800000000000), 56.77062065062736, 4.308085710844578},
					{execute.Time(1440804540000000000), 291.3806861459567, 4.308085710844578},
					{execute.Time(1440827280000000000), 709.9025450745096, 4.308085710844578},
					{execute.Time(1440850020000000000), 586.9971224242656, 4.308085710844578},
					{execute.Time(1440872760000000000), 3212.9646577264193, 4.308085710844578},
					{execute.Time(1440895500000000000), 17562.424047669054, 4.308085710844578},
					{execute.Time(1440918240000000000), 42944.7111670221, 4.308085710844578},
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
					{execute.Time(1440713580000000000), 6.517746679116747, 30.52018686151099},
					{execute.Time(1440736320000000000), 8.44990936365862, 30.52018686151099},
					{execute.Time(1440759060000000000), 11.450103919073781, 30.52018686151099},
					{execute.Time(1440781800000000000), 16.108703181906527, 30.52018686151099},
					{execute.Time(1440804540000000000), 23.342418185768715, 30.52018686151099},
					{execute.Time(1440827280000000000), 34.57468700800914, 30.52018686151099},
					{execute.Time(1440850020000000000), 52.01577652114399, 30.52018686151099},
					{execute.Time(1440872760000000000), 79.09771500282905, 30.52018686151099},
					{execute.Time(1440895500000000000), 121.14964091593322, 30.52018686151099},
					{execute.Time(1440918240000000000), 186.4464618626079, 30.52018686151099},
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
