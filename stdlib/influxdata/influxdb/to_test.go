package influxdb_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	influxdb2 "github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal"
	"github.com/influxdata/flux/values/valuestest"
	protocol "github.com/influxdata/line-protocol"
)

type pointsWriter struct {
	writes []protocol.Metric
}

func (t *pointsWriter) Close() error {
	return nil
}

func (t *pointsWriter) Write(metric ...protocol.Metric) error {
	t.writes = append(t.writes, metric...)
	return nil
}

func rowMetric(m string, tags [][2]string, fields [][2]interface{}, ts time.Time) *internal.RowMetric {
	metric := &internal.RowMetric{
		NameStr: m,
		Tags:    make([]*protocol.Tag, 0, len(tags)),
		Fields:  make([]*protocol.Field, 0, len(fields)),
		TS:      ts.UTC(),
	}

	for _, tag := range tags {
		metric.Tags = append(metric.Tags, &protocol.Tag{Key: tag[0], Value: tag[1]})
	}

	for _, field := range fields {
		metric.Fields = append(metric.Fields, &protocol.Field{Key: field[0].(string), Value: field[1]})
	}

	return metric
}

func TestTo_Process(t *testing.T) {
	type wanted struct {
		tables []*executetest.Table
		result *pointsWriter
	}
	testCases := []struct {
		name    string
		spec    *influxdb.ToProcedureSpec
		data    []*executetest.Table
		want    wanted
		wantErr error
	}{
		{
			name: "default case",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_field", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(11), "a", 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), "a", 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), "b", 1.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(31), "a", 3.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(41), "c", 4.0, "_value"},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 21)),
						rowMetric("b", [][2]string{}, [][2]interface{}{{"_value", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"_value", 3.0}}, time.Unix(0, 31)),
						rowMetric("c", [][2]string{}, [][2]interface{}{{"_value", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "wrong measurement column",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_wrong",
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_field", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(11), "a", 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), "a", 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), "b", 1.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(31), "a", 3.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(41), "c", 4.0, "_value"},
				},
			}},
			wantErr: fmt.Errorf("no column with label _wrong exists"),
		},
		{
			name: "wrong measurement type",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TInt},
					{Label: "_value", Type: flux.TFloat},
					{Label: "_field", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(11), int64(1), 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), int64(1), 2.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(21), int64(2), 1.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(31), int64(1), 3.0, "_value"},
					{execute.Time(0), execute.Time(100), execute.Time(41), int64(3), 4.0, "_value"},
				},
			}},
			wantErr: fmt.Errorf("column _measurement of type int is not of type string"),
		},
		{
			name: "default with multiple tag columns",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
					{Label: "tag2", Type: flux.TString},
					{Label: "_field", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", "a", "aa", "_value", 2.0},
					{execute.Time(21), "a", "a", "bb", "_value", 2.0},
					{execute.Time(21), "a", "b", "cc", "_value", 1.0},
					{execute.Time(31), "a", "a", "dd", "_value", 3.0},
					{execute.Time(41), "a", "c", "ee", "_value", 4.0},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "aa"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "bb"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{{"tag1", "b"}, {"tag2", "cc"}}, [][2]interface{}{{"_value", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "dd"}}, [][2]interface{}{{"_value", 3.0}}, time.Unix(0, 31)),
						rowMetric("a", [][2]string{{"tag1", "c"}, {"tag2", "ee"}}, [][2]interface{}{{"_value", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "default with heterogeneous tag columns",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "tag1", "tag2", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", "a", "aa", "_value", 2.0},
						{execute.Time(21), "a", "a", "bb", "_value", 2.0},
						{execute.Time(21), "a", "b", "cc", "_value", 1.0},
						{execute.Time(31), "a", "a", "dd", "_value", 3.0},
						{execute.Time(41), "a", "c", "ee", "_value", 4.0},
					},
				},
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "tagA", Type: flux.TString},
						{Label: "tagB", Type: flux.TString},
						{Label: "tagC", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "b", "a", "aa", "ff", "_value", 2.0},
						{execute.Time(21), "b", "a", "bb", "gg", "_value", 2.0},
						{execute.Time(21), "b", "b", "cc", "hh", "_value", 1.0},
						{execute.Time(31), "b", "a", "dd", "ii", "_value", 3.0},
						{execute.Time(41), "b", "c", "ee", "jj", "_value", 4.0},
					},
				},
			},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "aa"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "bb"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{{"tag1", "b"}, {"tag2", "cc"}}, [][2]interface{}{{"_value", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{{"tag1", "a"}, {"tag2", "dd"}}, [][2]interface{}{{"_value", 3.0}}, time.Unix(0, 31)),
						rowMetric("a", [][2]string{{"tag1", "c"}, {"tag2", "ee"}}, [][2]interface{}{{"_value", 4.0}}, time.Unix(0, 41)),
						rowMetric("b", [][2]string{{"tagA", "a"}, {"tagB", "aa"}, {"tagC", "ff"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 11)),
						rowMetric("b", [][2]string{{"tagA", "a"}, {"tagB", "bb"}, {"tagC", "gg"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 21)),
						rowMetric("b", [][2]string{{"tagA", "b"}, {"tagB", "cc"}, {"tagC", "hh"}}, [][2]interface{}{{"_value", 1.0}}, time.Unix(0, 21)),
						rowMetric("b", [][2]string{{"tagA", "a"}, {"tagB", "dd"}, {"tagC", "ii"}}, [][2]interface{}{{"_value", 3.0}}, time.Unix(0, 31)),
						rowMetric("b", [][2]string{{"tagA", "c"}, {"tagB", "ee"}, {"tagC", "jj"}}, [][2]interface{}{{"_value", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "explicit tags",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					TagColumns:        []string{"tag1", "tag2"},
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_field", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag1", Type: flux.TString},
					{Label: "tag2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "m", "_value", 2.0, "a", "aa"},
					{execute.Time(21), "m", "_value", 2.0, "a", "bb"},
					{execute.Time(21), "m", "_value", 1.0, "b", "cc"},
					{execute.Time(31), "m", "_value", 3.0, "a", "dd"},
					{execute.Time(41), "m", "_value", 4.0, "c", "ee"},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("m", [][2]string{{"tag1", "a"}, {"tag2", "aa"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 11)),
						rowMetric("m", [][2]string{{"tag1", "a"}, {"tag2", "bb"}}, [][2]interface{}{{"_value", 2.0}}, time.Unix(0, 21)),
						rowMetric("m", [][2]string{{"tag1", "b"}, {"tag2", "cc"}}, [][2]interface{}{{"_value", 1.0}}, time.Unix(0, 21)),
						rowMetric("m", [][2]string{{"tag1", "a"}, {"tag2", "dd"}}, [][2]interface{}{{"_value", 3.0}}, time.Unix(0, 31)),
						rowMetric("m", [][2]string{{"tag1", "c"}, {"tag2", "ee"}}, [][2]interface{}{{"_value", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "explicit field function",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
					FieldFn: interpreter.ResolvedFunction{
						Fn:    executetest.FunctionExpression(t, `(r) => ({temperature: r.temperature})`),
						Scope: valuestest.Scope(),
					},
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "temperature", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", 2.0},
					{execute.Time(21), "a", 2.0},
					{execute.Time(21), "b", 1.0},
					{execute.Time(31), "a", 3.0},
					{execute.Time(41), "c", 4.0},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 21)),
						rowMetric("b", [][2]string{}, [][2]interface{}{{"temperature", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 3.0}}, time.Unix(0, 31)),
						rowMetric("c", [][2]string{}, [][2]interface{}{{"temperature", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "explicit field function with custom measurement",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "the_msmnt",
					FieldFn: interpreter.ResolvedFunction{
						Fn:    executetest.FunctionExpression(t, `(r) => ({temperature: r.temperature})`),
						Scope: valuestest.Scope(),
					},
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "the_msmnt", Type: flux.TString},
					{Label: "temperature", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", "m0", 2.0},
					{execute.Time(21), "a", "m0", 2.0},
					{execute.Time(21), "b", "m1", 1.0},
					{execute.Time(31), "a", "m0", 3.0},
					{execute.Time(41), "c", "m2", 4.0},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("m0", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 11)),
						rowMetric("m0", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 21)),
						rowMetric("m1", [][2]string{}, [][2]interface{}{{"temperature", 1.0}}, time.Unix(0, 21)),
						rowMetric("m0", [][2]string{}, [][2]interface{}{{"temperature", 3.0}}, time.Unix(0, 31)),
						rowMetric("m2", [][2]string{}, [][2]interface{}{{"temperature", 4.0}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "infer tags from complex field function",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
					FieldFn: interpreter.ResolvedFunction{
						Fn:    executetest.FunctionExpression(t, `(r) => ({day: r.day, temperature: r.temperature, humidity: r.humidity, ratio: r.temperature / r.humidity})`),
						Scope: valuestest.Scope(),
					},
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_field", Type: flux.TString},
					{Label: "day", Type: flux.TString},
					{Label: "tag", Type: flux.TString},
					{Label: "temperature", Type: flux.TFloat},
					{Label: "humidity", Type: flux.TFloat},
					{Label: "_value", Type: flux.TString},
				},
				KeyCols: []string{"_measurement", "_field"},
				Data: [][]interface{}{
					{execute.Time(11), "a", "f", "Monday", "a", 2.0, 1.0, "bogus"},
					{execute.Time(21), "a", "f", "Tuesday", "a", 2.0, 2.0, "bogus"},
					{execute.Time(21), "a", "f", "Wednesday", "b", 1.0, 4.0, "bogus"},
					{execute.Time(31), "a", "f", "Thursday", "a", 3.0, 3.0, "bogus"},
					{execute.Time(41), "a", "f", "Friday", "c", 4.0, 5.0, "bogus"},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{
							{"tag", "a"},
						}, [][2]interface{}{
							{"day", "Monday"},
							{"humidity", 1.0},
							{"ratio", 2.0},
							{"temperature", 2.0},
						}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{
							{"tag", "a"},
						}, [][2]interface{}{
							{"day", "Tuesday"},
							{"humidity", 2.0},
							{"ratio", 1.0},
							{"temperature", 2.0},
						},
							time.Unix(0, 21)),
						rowMetric("a", [][2]string{
							{"tag", "b"},
						}, [][2]interface{}{
							{"day", "Wednesday"},
							{"humidity", 4.0},
							{"ratio", 0.25},
							{"temperature", 1.0},
						},
							time.Unix(0, 21)),
						rowMetric("a", [][2]string{
							{"tag", "a"},
						}, [][2]interface{}{
							{"day", "Thursday"},
							{"humidity", 3.0},
							{"ratio", 1.0},
							{"temperature", 3.0},
						},
							time.Unix(0, 31)),
						rowMetric("a", [][2]string{
							{"tag", "c"},
						}, [][2]interface{}{
							{"day", "Friday"},
							{"humidity", 5.0},
							{"ratio", 0.8},
							{"temperature", 4.0},
						},
							time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "explicit tag columns, multiple values in field function, and extra columns",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					MeasurementColumn: "_measurement",
					TagColumns:        []string{"tag1", "tag2"},
					FieldFn: interpreter.ResolvedFunction{
						Fn:    executetest.FunctionExpression(t, `(r) => ({temperature: r.temperature, humidity: r.humidity})`),
						Scope: valuestest.Scope(),
					},
				},
			},
			data: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
					{Label: "tag2", Type: flux.TString},
					{Label: "other-string-column", Type: flux.TString},
					{Label: "temperature", Type: flux.TFloat},
					{Label: "humidity", Type: flux.TInt},
					{Label: "other-value-column", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(11), "a", "a", "d", "misc", 2.0, int64(50), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(21), "a", "a", "d", "misc", 2.0, int64(50), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(21), "a", "b", "d", "misc", 1.0, int64(50), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(31), "a", "a", "e", "misc", 3.0, int64(60), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(41), "a", "c", "e", "misc", 4.0, int64(65), 1.0},
				},
			}},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{
							{"tag1", "a"},
							{"tag2", "d"},
						}, [][2]interface{}{
							{"humidity", int64(50)},
							{"temperature", 2.0},
						}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{
							{"tag1", "a"},
							{"tag2", "d"},
						}, [][2]interface{}{
							{"humidity", int64(50)},
							{"temperature", 2.0},
						},
							time.Unix(0, 21)),
						rowMetric("a", [][2]string{
							{"tag1", "b"},
							{"tag2", "d"},
						}, [][2]interface{}{
							{"humidity", int64(50)},
							{"temperature", 1.0},
						},
							time.Unix(0, 21)),
						rowMetric("a", [][2]string{
							{"tag1", "a"},
							{"tag2", "e"},
						}, [][2]interface{}{
							{"humidity", int64(60)},
							{"temperature", 3.0},
						},
							time.Unix(0, 31)),
						rowMetric("a", [][2]string{
							{"tag1", "c"},
							{"tag2", "e"},
						}, [][2]interface{}{
							{"humidity", int64(65)},
							{"temperature", 4.0},
						},
							time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "null values",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        execute.DefaultTimeColLabel,
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", "temperature", 2.0},
						{execute.Time(21), "a", "temperature", 1.0},
						{execute.Time(31), "a", "temperature", 3.0},
						{execute.Time(41), "a", "temperature", 4.0},
					},
				},
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", "humidity", int64(50)},
						{execute.Time(21), "a", "humidity", int64(50)},
						{execute.Time(31), "a", "humidity", nil},
						{execute.Time(41), "a", "humidity", int64(65)},
					},
				},
			},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 3.0}}, time.Unix(0, 31)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 4.0}}, time.Unix(0, 41)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(50)}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(50)}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(65)}}, time.Unix(0, 41)),
					},
				},
			},
		},
		{
			name: "null timestamp",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        execute.DefaultTimeColLabel,
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", "temperature", 2.0},
						{execute.Time(21), "a", "temperature", 1.0},
						{execute.Time(31), "a", "temperature", 3.0},
						{nil, "a", "temperature", 4.0},
					},
				},
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", "humidity", int64(50)},
						{execute.Time(21), "a", "humidity", int64(50)},
						{execute.Time(31), "a", "humidity", nil},
						{execute.Time(41), "a", "humidity", int64(65)},
					},
				},
			},
			want: wanted{
				result: &pointsWriter{
					writes: []protocol.Metric{
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 2.0}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 1.0}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"temperature", 3.0}}, time.Unix(0, 31)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(50)}}, time.Unix(0, 11)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(50)}}, time.Unix(0, 21)),
						rowMetric("a", [][2]string{}, [][2]interface{}{{"humidity", int64(65)}}, time.Unix(0, 41)),
					},
				},
				tables: []*executetest.Table{
					{
						KeyCols: []string{"_measurement", "_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_measurement", Type: flux.TString},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(11), "a", "temperature", 2.0},
							{execute.Time(21), "a", "temperature", 1.0},
							{execute.Time(31), "a", "temperature", 3.0},
						},
					},
					{
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_measurement", Type: flux.TString},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{execute.Time(11), "a", "humidity", int64(50)},
							{execute.Time(21), "a", "humidity", int64(50)},
							{execute.Time(31), "a", "humidity", nil},
							{execute.Time(41), "a", "humidity", int64(65)},
						},
					},
				},
			},
		},
		{
			name: "input tables with duplicate group key",
			spec: &influxdb.ToProcedureSpec{
				Spec: &influxdb.ToOpSpec{
					Org:               "my-org",
					Bucket:            "my-bucket",
					TimeColumn:        "_time",
					TagColumns:        []string{"tag1", "tag2"},
					MeasurementColumn: "_measurement",
				},
			},
			data: []*executetest.Table{
				{
					KeyCols: []string{"_field", "_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "m", "f", 1.0},
					},
				},
				{
					KeyCols: []string{"_field", "_measurement"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(11), "m", "f", 1.0},
					},
				},
			},
			wantErr: errors.New("to() found duplicate table with group key: {_field=f,_measurement=m}"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				inTables   = make([]flux.Table, 0, len(tc.data))
				wantTables = make([]*executetest.Table, 0, len(tc.data))
			)
			writer := &pointsWriter{}
			provider := mock.InfluxDBProvider{
				WriterForFn: func(ctx context.Context, conf influxdb2.Config) (influxdb2.Writer, error) {
					return writer, nil
				},
			}

			if tc.want.tables != nil && len(tc.data) != len(tc.want.tables) {
				t.Errorf("tc.data has %d tables but tc.want.tables has %d tables.", len(tc.data), len(tc.want.tables))
			}
			for i, tbl := range tc.data {
				rwTable := &executetest.RowWiseTable{Table: tbl}
				inTables = append(inTables, rwTable)
				if tc.want.tables != nil {
					wantTables = append(wantTables, tc.want.tables[i])
				} else {
					wantTables = append(wantTables, tbl)
				}
			}

			executetest.ProcessTestHelper(
				t,
				inTables,
				wantTables,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					newT, _ := influxdb.NewToTransformation(context.TODO(), d, c, tc.spec, provider)
					return newT
				},
			)
			for _, m := range writer.writes {
				rm := m.(*internal.RowMetric)
				sort.Slice(rm.Fields, func(i, j int) bool {
					return rm.Fields[i].Key < rm.Fields[j].Key
				})
				sort.Slice(rm.Tags, func(i, j int) bool {
					return rm.Tags[i].Key < rm.Tags[j].Key
				})
			}

			if tc.wantErr == nil {
				if len(writer.writes) != len(tc.want.result.writes) {
					t.Errorf("Expected result values to have length of %d but got %d", len(tc.want.result.writes), len(writer.writes))
				}

				if !cmp.Equal(writer, tc.want.result, cmp.AllowUnexported(pointsWriter{})) {
					t.Errorf("got other than expected -want/+got %s", cmp.Diff(tc.want.result, writer, cmp.AllowUnexported(pointsWriter{})))
				}
			}
		})
	}
}
