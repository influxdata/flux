package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
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
	data := []byte(`{"id":"hw","kind":"holtWinters","spec":{"n":84,"s":4,"interval":"42m","time_column":"t","column":"v","with_fit":true}}`)
	op := &flux.Operation{
		ID: "hw",
		Spec: &universe.HoltWintersOpSpec{
			WithFit:    true,
			Column:     "v",
			TimeColumn: "t",
			N:          84,
			S:          4,
			Interval:   flux.ConvertDuration(42 * time.Minute),
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestHoltWinters_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewHoltWintersTransformation(
			d,
			c,
			&memory.Allocator{},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.9914180362073646},
					{execute.Time(1440759060000000000), 0.8307286750677001},
					{execute.Time(1440781800000000000), 3.6959256712424695},
					{execute.Time(1440804540000000000), 2.945758601382681},
					{execute.Time(1440827280000000000), 3.398909006745872},
					{execute.Time(1440850020000000000), 0.6985261817067938},
					{execute.Time(1440872760000000000), 3.0650009664543356},
					{execute.Time(1440895500000000000), 2.4176070926199644},
					{execute.Time(1440918240000000000), 2.7815982631565688},
					{execute.Time(1440940980000000000), 0.5664788122687602},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440281520000000000), 4.948},
					{execute.Time(1440304260000000000), 0.7576983482641251},
					{execute.Time(1440327000000000000), 3.1381437247133945},
					{execute.Time(1440349740000000000), 2.5875124253677586},
					{execute.Time(1440372480000000000), 3.0898163409477224},
					{execute.Time(1440395220000000000), 0.7162614080885075},
					{execute.Time(1440417960000000000), 3.699916116400646},
					{execute.Time(1440440700000000000), 3.2863447391194933},
					{execute.Time(1440463440000000000), 3.9634098428094706},
					{execute.Time(1440486180000000000), 0.8997067190366455},
					{execute.Time(1440508920000000000), 4.430830461312907},
					{execute.Time(1440531660000000000), 3.785850242591291},
					{execute.Time(1440554400000000000), 4.479264854422211},
					{execute.Time(1440577140000000000), 0.9761816083195014},
					{execute.Time(1440599880000000000), 4.577665330603308},
					{execute.Time(1440622620000000000), 3.7849630817187445},
					{execute.Time(1440645360000000000), 4.421113012219597},
					{execute.Time(1440668100000000000), 0.9371785323971387},
					{execute.Time(1440690840000000000), 4.2567258021603624},
					{execute.Time(1440713580000000000), 3.443334270453622},
					{execute.Time(1440736320000000000), 3.9914180362073646},
					{execute.Time(1440759060000000000), 0.8307286750677001},
					{execute.Time(1440781800000000000), 3.6959256712424695},
					{execute.Time(1440804540000000000), 2.945758601382681},
					{execute.Time(1440827280000000000), 3.398909006745872},
					{execute.Time(1440850020000000000), 0.6985261817067938},
					{execute.Time(1440872760000000000), 3.0650009664543356},
					{execute.Time(1440895500000000000), 2.4176070926199644},
					{execute.Time(1440918240000000000), 2.7815982631565688},
					{execute.Time(1440940980000000000), 0.5664788122687602},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.9914180362073646},
					{execute.Time(1440759060000000000), 0.8307286750677001},
					{execute.Time(1440781800000000000), 3.6959256712424695},
					{execute.Time(1440804540000000000), 2.945758601382681},
					{execute.Time(1440827280000000000), 3.398909006745872},
					{execute.Time(1440850020000000000), 0.6985261817067938},
					{execute.Time(1440872760000000000), 3.0650009664543356},
					{execute.Time(1440895500000000000), 2.4176070926199644},
					{execute.Time(1440918240000000000), 2.7815982631565688},
					{execute.Time(1440940980000000000), 0.5664788122687602},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.9914180362073646},
					{execute.Time(1440759060000000000), 0.8307286750677001},
					{execute.Time(1440781800000000000), 3.6959256712424695},
					{execute.Time(1440804540000000000), 2.945758601382681},
					{execute.Time(1440827280000000000), 3.398909006745872},
					{execute.Time(1440850020000000000), 0.6985261817067938},
					{execute.Time(1440872760000000000), 3.0650009664543356},
					{execute.Time(1440895500000000000), 2.4176070926199644},
					{execute.Time(1440918240000000000), 2.7815982631565688},
					{execute.Time(1440940980000000000), 0.5664788122687602},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.9914180362073646},
					{execute.Time(1440759060000000000), 0.8307286750677001},
					{execute.Time(1440781800000000000), 3.6959256712424695},
					{execute.Time(1440804540000000000), 2.945758601382681},
					{execute.Time(1440827280000000000), 3.398909006745872},
					{execute.Time(1440850020000000000), 0.6985261817067938},
					{execute.Time(1440872760000000000), 3.0650009664543356},
					{execute.Time(1440895500000000000), 2.4176070926199644},
					{execute.Time(1440918240000000000), 2.7815982631565688},
					{execute.Time(1440940980000000000), 0.5664788122687602},
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
					},
					Data: [][]interface{}{
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440736320000000000), 3.9914180362073646},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440759060000000000), 0.8307286750677001},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440781800000000000), 3.6959256712424695},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440804540000000000), 2.945758601382681},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440827280000000000), 3.398909006745872},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440850020000000000), 0.6985261817067938},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440872760000000000), 3.0650009664543356},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440895500000000000), 2.4176070926199644},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440918240000000000), 2.7815982631565688},
						{"t", int64(0), uint64(0), 0.0, true, execute.Time(0), execute.Time(1440940980000000000), 0.5664788122687602},
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
					},
					Data: [][]interface{}{
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440736320000000000), 3.9914180362073646},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440759060000000000), 0.8307286750677001},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440781800000000000), 3.6959256712424695},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440804540000000000), 2.945758601382681},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440827280000000000), 3.398909006745872},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440850020000000000), 0.6985261817067938},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440872760000000000), 3.0650009664543356},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440895500000000000), 2.4176070926199644},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440918240000000000), 2.7815982631565688},
						{"t", int64(0), uint64(0), 0.0, false, execute.Time(0), execute.Time(1440940980000000000), 0.5664788122687602},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440713580000000000), 3.5876419909685495},
					{execute.Time(1440736320000000000), 3.49153693928191},
					{execute.Time(1440759060000000000), 3.2698815098513583},
					{execute.Time(1440781800000000000), 3.19382138573678},
					{execute.Time(1440804540000000000), 3.6184993955426985},
					{execute.Time(1440827280000000000), 3.5263702582818666},
					{execute.Time(1440850020000000000), 3.2973336160921813},
					{execute.Time(1440872760000000000), 3.2217179076885594},
					{execute.Time(1440895500000000000), 3.6392336870810817},
					{execute.Time(1440918240000000000), 3.5473339404590325},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440736320000000000), 3.254191388070825},
					{execute.Time(1440759060000000000), 3.2541742264816786},
					{execute.Time(1440781800000000000), 3.254163964980623},
					{execute.Time(1440804540000000000), 3.2541578295028817},
					{execute.Time(1440827280000000000), 3.2541541611055496},
					{execute.Time(1440850020000000000), 3.254151967802365},
					{execute.Time(1440872760000000000), 3.254150656455532},
					{execute.Time(1440895500000000000), 3.2541498724223517},
					{execute.Time(1440918240000000000), 3.2541494036628342},
					{execute.Time(1440940980000000000), 3.2541491234003104},
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
				},
				Data: [][]interface{}{
					{execute.Time(1440713580000000000), 6.514215038632055},
					{execute.Time(1440736320000000000), 8.497982411207028},
					{execute.Time(1440759060000000000), 11.634451346818643},
					{execute.Time(1440781800000000000), 16.593649776280504},
					{execute.Time(1440804540000000000), 24.43508329162218},
					{execute.Time(1440827280000000000), 36.834133612921335},
					{execute.Time(1440850020000000000), 56.440059666607354},
					{execute.Time(1440872760000000000), 87.44210023894752},
					{execute.Time(1440895500000000000), 136.46464506566383},
					{execute.Time(1440918240000000000), 213.98275775223067},
				},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			alloc := &memory.Allocator{}
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewHoltWintersTransformation(d, c, alloc, tc.spec)
				},
			)

			if m := alloc.Allocated(); m != 0 {
				t.Errorf("HoltWinters is using memory after finishing: %d", m)
			}
		})
	}
}
