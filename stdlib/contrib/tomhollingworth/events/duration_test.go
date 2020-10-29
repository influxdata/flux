package events_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/contrib/tomhollingworth/events"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestDuration_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "duration missing stop column",
			Raw:     `import "contrib/tomhollingworth/events" from(bucket:"mydb") |> range(start:-1h) |> drop(columns: ["_stop"]  |> events.duration()`,
			WantErr: true,
		},
		{
			Name:    "duration missing time column",
			Raw:     `import "contrib/tomhollingworth/events" from(bucket:"mydb") |> range(start:-1h) |> drop(columns: ["_time"]  |> events.duration()`,
			WantErr: true,
		},
		{
			Name:    "duration default",
			Raw:     `import "contrib/tomhollingworth/events" from(bucket:"mydb") |> range(start:-1h)  |> events.duration()`,
			WantErr: false,
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
						ID: "duration2",
						Spec: &events.DurationOpSpec{
							Unit:       flux.ConvertDuration(time.Second),
							TimeColumn: "_time",
							ColumnName: "duration",
							StopColumn: "_stop",
							Stop:       flux.Now,
							IsStop:     false,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "duration2"},
				},
			},
		},
		{
			Name:    "duration different unit and columns",
			Raw:     `import "contrib/tomhollingworth/events" from(bucket:"mydb") |> range(start:-1h)  |> events.duration(unit: 1ms, timeColumn: "start", stopColumn: "end", columnName: "result")`,
			WantErr: false,
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
						ID: "duration2",
						Spec: &events.DurationOpSpec{
							Unit:       flux.ConvertDuration(time.Millisecond),
							TimeColumn: "start",
							ColumnName: "result",
							StopColumn: "end",
							Stop:       flux.Now,
							IsStop:     false,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "duration2"},
				},
			},
		},
		{
			Name:    "duration with stop",
			Raw:     `import "contrib/tomhollingworth/events" from(bucket:"mydb") |> range(start:-1h)  |> events.duration(stop: 2020-10-20T08:30:00Z)`,
			WantErr: false,
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
						ID: "duration2",
						Spec: &events.DurationOpSpec{
							Unit:       flux.ConvertDuration(time.Second),
							TimeColumn: "_time",
							ColumnName: "duration",
							StopColumn: "_stop",
							Stop: flux.Time{
								Absolute: time.Date(2020, 10, 20, 8, 30, 0, 0, time.UTC),
							},
							IsStop: true,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "duration2"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestDurationOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"duration","kind":"duration","spec":{"timeColumn": "_time"}}`)
	op := &flux.Operation{
		ID: "duration",
		Spec: &events.DurationOpSpec{
			TimeColumn: "_time",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestDuration_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := events.NewDurationTransformation(
			d,
			c,
			&events.DurationProcedureSpec{},
		)
		return s
	})
}

func TestDuration_DurationProcedureSpec(t *testing.T) {
	goTime, _ := time.Parse(time.RFC3339, "2020-10-10T08:00:00Z")

	s := events.DurationProcedureSpec{
		Unit:       flux.ConvertDuration(time.Nanosecond),
		TimeColumn: execute.DefaultTimeColLabel,
		ColumnName: "duration",
		StopColumn: execute.DefaultStopColLabel,
		Stop: flux.Time{
			IsRelative: false,
			Relative:   time.Duration(0),
			Absolute:   goTime,
		},
		IsStop: true,
	}

	if s.Kind() != "duration" {
		t.Errorf("s.Kind() != %s; want duration", s.Kind())
	}

	sCopy := s.Copy()

	if sCopy.Kind() != s.Kind() {
		t.Errorf("sCopy.Kind() != %s; want %s", sCopy.Kind(), s.Kind())
	}
}

func TestDuration_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *events.DurationProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "basic output",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(10), execute.Time(1)},
					{execute.Time(1), execute.Time(10), execute.Time(3)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(10), execute.Time(1), int64(execute.Time(3) - execute.Time(1))},
					{execute.Time(1), execute.Time(10), execute.Time(3), int64(execute.Time(10) - execute.Time(3))},
				},
			}},
		},
		{
			name: "basic output. test columnName",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration_label",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(10), execute.Time(1)},
					{execute.Time(1), execute.Time(10), execute.Time(3)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration_label", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), execute.Time(10), execute.Time(1), int64(execute.Time(3) - execute.Time(1))},
					{execute.Time(1), execute.Time(10), execute.Time(3), int64(execute.Time(10) - execute.Time(3))},
				},
			}},
		},
		{
			name: "basic output. test timeColumn",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Second),
				TimeColumn: "timeStamp",
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "timeStamp", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(1 * time.Second)},
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(3 * time.Second)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "timeStamp", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(1 * time.Second), int64(2)},
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(3 * time.Second), int64(7)},
				},
			}},
		},
		{
			name: "basic output. test stopColumn",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Second),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: "end",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "end", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(1 * time.Second)},
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(3 * time.Second)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "end", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(1 * time.Second), int64(2)},
					{execute.Time(1 * time.Second), execute.Time(10 * time.Second), execute.Time(3 * time.Second), int64(7)},
				},
			}},
		},
		{
			name: "basic output. test unit",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Second),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(10 * time.Second), execute.Time(1 * time.Second)},
					{execute.Time(10 * time.Second), execute.Time(5 * time.Second)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(10 * time.Second), execute.Time(1 * time.Second), int64(4)},
					{execute.Time(10 * time.Second), execute.Time(5 * time.Second), int64(5)},
				},
			}},
		},
		{
			name: "basic output. test stop",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Second),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
				IsStop:     true,
				Stop: flux.Time{
					IsRelative: false,
					Relative:   0,
					Absolute:   time.Unix(10, 0),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second)},
					{execute.Time(3 * time.Second)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1 * time.Second), int64(2)},
					{execute.Time(3 * time.Second), int64(7)},
				},
			}},
		},
		{
			name: "a little less basic output, but still simple",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(20), execute.Time(1)},
					{execute.Time(20), execute.Time(2)},
					{execute.Time(20), execute.Time(3)},
					{execute.Time(20), execute.Time(4)},
					{execute.Time(20), execute.Time(5)},
					{execute.Time(20), execute.Time(6)},
					{execute.Time(20), execute.Time(7)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(20), execute.Time(1), int64(execute.Time(2) - execute.Time(1))},
					{execute.Time(20), execute.Time(2), int64(execute.Time(3) - execute.Time(2))},
					{execute.Time(20), execute.Time(3), int64(execute.Time(4) - execute.Time(3))},
					{execute.Time(20), execute.Time(4), int64(execute.Time(5) - execute.Time(4))},
					{execute.Time(20), execute.Time(5), int64(execute.Time(6) - execute.Time(5))},
					{execute.Time(20), execute.Time(6), int64(execute.Time(7) - execute.Time(6))},
					{execute.Time(20), execute.Time(7), int64(execute.Time(20) - execute.Time(7))},
				},
			}},
		},
		{
			name: "three columns: _stop, _time, _value",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(30), execute.Time(1), int64(2)},
					{execute.Time(30), execute.Time(2), int64(2)},
					{execute.Time(30), execute.Time(3), int64(2)},
					{execute.Time(30), execute.Time(4), int64(2)},
					{execute.Time(30), execute.Time(5), int64(7)},
					{execute.Time(30), execute.Time(6), int64(2)},
					{execute.Time(30), execute.Time(7), int64(2)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(30), execute.Time(1), int64(2), int64(execute.Time(2) - execute.Time(1))},
					{execute.Time(30), execute.Time(2), int64(2), int64(execute.Time(3) - execute.Time(2))},
					{execute.Time(30), execute.Time(3), int64(2), int64(execute.Time(4) - execute.Time(3))},
					{execute.Time(30), execute.Time(4), int64(2), int64(execute.Time(5) - execute.Time(4))},
					{execute.Time(30), execute.Time(5), int64(7), int64(execute.Time(6) - execute.Time(5))},
					{execute.Time(30), execute.Time(6), int64(2), int64(execute.Time(7) - execute.Time(6))},
					{execute.Time(30), execute.Time(7), int64(2), int64(execute.Time(30) - execute.Time(7))},
				},
			}},
		},
		{
			name: "four columns: stop, time, _value, path",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "path", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(10), execute.Time(1), 2.0, "/"},
					{execute.Time(10), execute.Time(2), 1.0, "/"},
					{execute.Time(10), execute.Time(3), 3.6, "/"},
					{execute.Time(10), execute.Time(4), 9.7, "/"},
					{execute.Time(10), execute.Time(5), 13.1, "/"},
					{execute.Time(10), execute.Time(6), 10.2, "/"},
					{execute.Time(10), execute.Time(7), 5.4, "/"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "path", Type: flux.TString},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(10), execute.Time(1), 2.0, "/", int64(execute.Time(2) - execute.Time(1))},
					{execute.Time(10), execute.Time(2), 1.0, "/", int64(execute.Time(3) - execute.Time(2))},
					{execute.Time(10), execute.Time(3), 3.6, "/", int64(execute.Time(4) - execute.Time(3))},
					{execute.Time(10), execute.Time(4), 9.7, "/", int64(execute.Time(5) - execute.Time(4))},
					{execute.Time(10), execute.Time(5), 13.1, "/", int64(execute.Time(6) - execute.Time(5))},
					{execute.Time(10), execute.Time(6), 10.2, "/", int64(execute.Time(7) - execute.Time(6))},
					{execute.Time(10), execute.Time(7), 5.4, "/", int64(execute.Time(10) - execute.Time(7))},
				},
			}},
		},
		{
			name: "multiple time columns",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: "start",
				ColumnName: "duration",
				StopColumn: "finish",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "finish", Type: flux.TTime},
					{Label: "start", Type: flux.TTime},
					{Label: "end", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(10), execute.Time(1), execute.Time(2)},
					{execute.Time(10), execute.Time(2), execute.Time(3)},
					{execute.Time(10), execute.Time(3), execute.Time(4)},
					{execute.Time(10), execute.Time(4), execute.Time(5)},
					{execute.Time(10), execute.Time(5), execute.Time(6)},
					{execute.Time(10), execute.Time(6), execute.Time(7)},
					{execute.Time(10), execute.Time(7), execute.Time(8)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "finish", Type: flux.TTime},
					{Label: "start", Type: flux.TTime},
					{Label: "end", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(10), execute.Time(1), execute.Time(2), int64(execute.Time(2) - execute.Time(1))},
					{execute.Time(10), execute.Time(2), execute.Time(3), int64(execute.Time(3) - execute.Time(2))},
					{execute.Time(10), execute.Time(3), execute.Time(4), int64(execute.Time(4) - execute.Time(3))},
					{execute.Time(10), execute.Time(4), execute.Time(5), int64(execute.Time(5) - execute.Time(4))},
					{execute.Time(10), execute.Time(5), execute.Time(6), int64(execute.Time(6) - execute.Time(5))},
					{execute.Time(10), execute.Time(6), execute.Time(7), int64(execute.Time(7) - execute.Time(6))},
					{execute.Time(10), execute.Time(7), execute.Time(8), int64(execute.Time(10) - execute.Time(7))},
				},
			}},
		},
		{
			name: "multiple buffers",
			spec: &events.DurationProcedureSpec{
				Unit:       flux.ConvertDuration(time.Nanosecond),
				TimeColumn: execute.DefaultTimeColLabel,
				ColumnName: "duration",
				StopColumn: execute.DefaultStopColLabel,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(50), execute.Time(0)},
						{execute.Time(50), execute.Time(1)},
						{execute.Time(50), execute.Time(2)},
						{execute.Time(50), execute.Time(3)},
						{execute.Time(50), execute.Time(4)},
						{execute.Time(50), execute.Time(5)},
						{execute.Time(50), execute.Time(6)},
						{execute.Time(50), execute.Time(7)},
						{execute.Time(50), execute.Time(8)},
						{execute.Time(50), execute.Time(9)},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(50), execute.Time(0), int64(execute.Time(50) - execute.Time(0))},
					{execute.Time(50), execute.Time(1), int64(execute.Time(50) - execute.Time(1))},
					{execute.Time(50), execute.Time(2), int64(execute.Time(50) - execute.Time(2))},
					{execute.Time(50), execute.Time(3), int64(execute.Time(50) - execute.Time(3))},
					{execute.Time(50), execute.Time(4), int64(execute.Time(50) - execute.Time(4))},
					{execute.Time(50), execute.Time(5), int64(execute.Time(50) - execute.Time(5))},
					{execute.Time(50), execute.Time(6), int64(execute.Time(50) - execute.Time(6))},
					{execute.Time(50), execute.Time(7), int64(execute.Time(50) - execute.Time(7))},
					{execute.Time(50), execute.Time(8), int64(execute.Time(50) - execute.Time(8))},
					{execute.Time(50), execute.Time(9), int64(execute.Time(50) - execute.Time(9))},
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
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return events.NewDurationTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
