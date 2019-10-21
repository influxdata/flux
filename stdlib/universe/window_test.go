package universe_test

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestWindow_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with window",
			Raw:  `from(bucket:"mybucket") |> window(every:1h, offset: -5m)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "window1",
						Spec: &universe.WindowOpSpec{
							Every:       flux.ConvertDuration(time.Hour),
							Period:      flux.ConvertDuration(time.Hour),
							Offset:      flux.ConvertDuration(time.Minute * -5),
							TimeColumn:  execute.DefaultTimeColLabel,
							StartColumn: execute.DefaultStartColLabel,
							StopColumn:  execute.DefaultStopColLabel,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "window1"},
				},
			},
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

func TestWindowOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"window","kind":"window","spec":{"every":"1m","period":"1h","offset":"30m"}}`)
	op := &flux.Operation{
		ID: "window",
		Spec: &universe.WindowOpSpec{
			Every:  flux.ConvertDuration(time.Minute),
			Period: flux.ConvertDuration(time.Hour),
			Offset: flux.ConvertDuration(30 * time.Minute),
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFixedWindow_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		fw := universe.NewFixedWindowTransformation(
			d,
			c,
			execute.Bounds{},
			execute.NewWindow(
				values.ConvertDuration(time.Minute),
				values.ConvertDuration(time.Minute),
				values.ConvertDuration(0)),
			execute.DefaultTimeColLabel,
			execute.DefaultStartColLabel,
			execute.DefaultStopColLabel,
			false,
		)
		return fw
	})
}

func newEmptyWindowTable(start execute.Time, stop execute.Time, cols []flux.ColMeta) *executetest.Table {
	return &executetest.Table{
		KeyCols:   []string{"_start", "_stop"},
		KeyValues: []interface{}{start, stop},
		ColMeta:   cols,
		Data:      [][]interface{}(nil),
		GroupKey: execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "_start", Type: flux.TTime},
				{Label: "_stop", Type: flux.TTime},
			},
			[]values.Value{
				values.NewTime(start),
				values.NewTime(stop),
			},
		),
	}
}

func TestFixedWindow_Process(t *testing.T) {
	alignedBounds := execute.Bounds{
		Start: execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
		Stop:  execute.Time(time.Date(2017, 10, 10, 0, 10, 0, 0, time.UTC).UnixNano()),
	}
	nonalignedBounds := execute.Bounds{
		Start: execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
		Stop:  execute.Time(time.Date(2017, 10, 10, 10, 20, 10, 10, time.UTC).UnixNano()),
	}

	// test columns which all expected data will use
	testCols := []flux.ColMeta{
		{Label: "_start", Type: flux.TTime},
		{Label: "_stop", Type: flux.TTime},
		{Label: "_time", Type: flux.TTime},
		{Label: "_value", Type: flux.TFloat},
	}
	testCases := []struct {
		name                  string
		valueCol              flux.ColMeta
		bounds                execute.Bounds
		every, period, offset execute.Duration
		createEmpty           bool
		num                   int
		want                  func(start execute.Time) []*executetest.Table
	}{
		{
			name:     "nonoverlapping_nonaligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use bounds and offset that is *not* aligned with the every/period durations of the window
			bounds:      nonalignedBounds,
			offset:      values.ConvertDuration(10*time.Second + 10*time.Nanosecond),
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(time.Minute),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, 0.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(3*time.Minute), start+execute.Time(4*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(4*time.Minute), start+execute.Time(5*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute), start+execute.Time(7*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute), start+execute.Time(9*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "nonoverlapping_aligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use bounds that are aligned with period and duration of window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(time.Minute),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, 0.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						KeyCols:   []string{"_start", "_stop"},
						KeyValues: []interface{}{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute)},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						GroupKey: execute.NewGroupKey(
							[]flux.ColMeta{
								{Label: "_start", Type: flux.TTime},
								{Label: "_stop", Type: flux.TTime},
							},
							[]values.Value{
								values.NewTime(start + execute.Time(3*time.Minute)),
								values.NewTime(start + execute.Time(4*time.Minute)),
							},
						),
					},
					newEmptyWindowTable(start+execute.Time(4*time.Minute), start+execute.Time(5*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute), start+execute.Time(7*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute), start+execute.Time(9*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "overlapping_nonaligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use a time that is *not* aligned with the every/period durations of the window
			bounds:      nonalignedBounds,
			offset:      values.ConvertDuration(time.Second*10 + time.Nanosecond*10),
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(2 * time.Minute),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, 0.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(2*time.Minute), start, 0.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(50*time.Second), 5.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(110*time.Second), 11.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(3*time.Minute), start+execute.Time(5*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(4*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(7*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(9*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute), start+execute.Time(10*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "overlapping_aligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use a bounds that are aligned with the every/period durations of the window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(2 * time.Minute),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, 0.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(2*time.Minute), start, 0.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(40*time.Second), 4.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(50*time.Second), 5.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start, start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(110*time.Second), 11.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(1*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(2*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(140*time.Second), 14.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(3*time.Minute), start+execute.Time(5*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(4*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(7*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(9*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute), start+execute.Time(10*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "underlapping_nonaligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use a time that is *not* aligned with the every/period durations of the window
			bounds:      nonalignedBounds,
			every:       values.ConvertDuration(2 * time.Minute),
			period:      values.ConvertDuration(time.Minute),
			offset:      values.ConvertDuration(10*time.Second + 10*time.Nanosecond),
			createEmpty: true,
			num:         24,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(230*time.Second), 23.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "underlapping_aligned",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use a time that is  aligned with the every/period durations of the window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(2 * time.Minute),
			period:      values.ConvertDuration(time.Minute),
			createEmpty: true,
			num:         24,
			want: func(start execute.Time) []*executetest.Table {
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(60*time.Second), 6.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(70*time.Second), 7.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(80*time.Second), 8.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(90*time.Second), 9.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(100*time.Second), 10.0},
							{start + 1*execute.Time(time.Minute), start + 2*execute.Time(time.Minute), start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(3*time.Minute), start + execute.Time(4*time.Minute), start + execute.Time(230*time.Second), 23.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "nonoverlapping_aligned_int",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TInt},
			// Use bounds that are aligned with the every/period durations of the window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(time.Minute),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				testCols := testCols
				testCols[3].Type = flux.TInt
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, int64(0.0)},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), int64(1)},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), int64(2)},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), int64(3)},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), int64(4)},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), int64(5)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), int64(6)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), int64(7)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), int64(8)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), int64(9)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), int64(10)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), int64(11)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), int64(12)},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), int64(13)},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), int64(14)},
						},
					},
					newEmptyWindowTable(start+execute.Time(3*time.Minute), start+execute.Time(4*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(4*time.Minute), start+execute.Time(5*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute), start+execute.Time(6*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute), start+execute.Time(7*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute), start+execute.Time(8*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute), start+execute.Time(9*time.Minute), testCols),
					newEmptyWindowTable(start+execute.Time(9*time.Minute), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "don't create empty",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TInt},
			// Use bounds that are aligned with the every/period durations of the window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(time.Minute),
			createEmpty: false,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				testCols := testCols
				testCols[3].Type = flux.TInt
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start, start + execute.Time(time.Minute), start, int64(0.0)},
							{start, start + execute.Time(time.Minute), start + execute.Time(10*time.Second), int64(1)},
							{start, start + execute.Time(time.Minute), start + execute.Time(20*time.Second), int64(2)},
							{start, start + execute.Time(time.Minute), start + execute.Time(30*time.Second), int64(3)},
							{start, start + execute.Time(time.Minute), start + execute.Time(40*time.Second), int64(4)},
							{start, start + execute.Time(time.Minute), start + execute.Time(50*time.Second), int64(5)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(60*time.Second), int64(6)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(70*time.Second), int64(7)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(80*time.Second), int64(8)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(90*time.Second), int64(9)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(100*time.Second), int64(10)},
							{start + execute.Time(1*time.Minute), start + execute.Time(2*time.Minute), start + execute.Time(110*time.Second), int64(11)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(120*time.Second), int64(12)},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(130*time.Second), int64(13)},
							{start + execute.Time(2*time.Minute), start + execute.Time(3*time.Minute), start + execute.Time(140*time.Second), int64(14)},
						},
					},
				}
			},
		},
		{
			name:     "empty bounds start == stop",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TInt},
			every:    values.ConvertDuration(time.Minute),
			period:   values.ConvertDuration(time.Minute),
			num:      15,
			bounds: execute.Bounds{
				Start: execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
				Stop:  execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
			},
			want: func(start execute.Time) []*executetest.Table {
				return nil
			},
		},
		{
			name:     "negative offset",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TFloat},
			// Use bounds that are aligned with the every/period durations of the window
			bounds:      alignedBounds,
			every:       values.ConvertDuration(time.Minute),
			period:      values.ConvertDuration(time.Minute),
			offset:      values.ConvertDuration(-15 * time.Second),
			createEmpty: true,
			num:         15,
			want: func(start execute.Time) []*executetest.Table {
				testCols := testCols
				testCols[3].Type = flux.TFloat
				return []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							// truncated initial window due to unaligned offset
							{start, start + execute.Time(45*time.Second), start, 0.0},
							{start, start + execute.Time(45*time.Second), start + execute.Time(10*time.Second), 1.0},
							{start, start + execute.Time(45*time.Second), start + execute.Time(20*time.Second), 2.0},
							{start, start + execute.Time(45*time.Second), start + execute.Time(30*time.Second), 3.0},
							{start, start + execute.Time(45*time.Second), start + execute.Time(40*time.Second), 4.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(50*time.Second), 5.0},
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(45*time.Second), start + execute.Time(time.Minute+45*time.Second), start + execute.Time(100*time.Second), 10.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(time.Minute+45*time.Second), start + execute.Time(2*time.Minute+45*time.Second), start + execute.Time(110*time.Second), 11.0},
							{start + execute.Time(time.Minute+45*time.Second), start + execute.Time(2*time.Minute+45*time.Second), start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(time.Minute+45*time.Second), start + execute.Time(2*time.Minute+45*time.Second), start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(time.Minute+45*time.Second), start + execute.Time(2*time.Minute+45*time.Second), start + execute.Time(140*time.Second), 14.0},
						},
					},
					newEmptyWindowTable(start+execute.Time(2*time.Minute+45*time.Second), start+execute.Time(3*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(3*time.Minute+45*time.Second), start+execute.Time(4*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(4*time.Minute+45*time.Second), start+execute.Time(5*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(5*time.Minute+45*time.Second), start+execute.Time(6*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(6*time.Minute+45*time.Second), start+execute.Time(7*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(7*time.Minute+45*time.Second), start+execute.Time(8*time.Minute+45*time.Second), testCols),
					newEmptyWindowTable(start+execute.Time(8*time.Minute+45*time.Second), start+execute.Time(9*time.Minute+45*time.Second), testCols),
					// truncated final window due to unaligned offset
					newEmptyWindowTable(start+execute.Time(9*time.Minute+45*time.Second), start+execute.Time(10*time.Minute), testCols),
				}
			},
		},
		{
			name:     "empty bounds start > stop",
			valueCol: flux.ColMeta{Label: "_value", Type: flux.TInt},
			every:    values.ConvertDuration(time.Minute),
			period:   values.ConvertDuration(time.Minute),
			num:      15,
			bounds: execute.Bounds{
				Start: execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
				Stop:  execute.Time(time.Date(2017, 9, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
			},
			want: func(start execute.Time) []*executetest.Table {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(plan.DefaultTriggerSpec)

			fw := universe.NewFixedWindowTransformation(
				d,
				c,
				tc.bounds,
				execute.NewWindow(tc.every, tc.period, tc.offset),
				execute.DefaultTimeColLabel,
				execute.DefaultStartColLabel,
				execute.DefaultStopColLabel,
				tc.createEmpty,
			)

			table0 := &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					tc.valueCol,
				},
			}

			for i := 0; i < tc.num; i++ {
				var v interface{}
				switch tc.valueCol.Type {
				case flux.TBool:
					v = bool(i%2 == 0)
				case flux.TInt:
					v = int64(i)
				case flux.TUInt:
					v = uint64(i)
				case flux.TFloat:
					v = float64(i)
				case flux.TString:
					v = strconv.Itoa(i)
				}
				table0.Data = append(table0.Data, []interface{}{
					tc.bounds.Start,
					tc.bounds.Stop,
					tc.bounds.Start + execute.Time(time.Duration(i)*10*time.Second),
					v,
				})
			}

			parentID := executetest.RandomDatasetID()
			if err := fw.Process(parentID, table0); err != nil {
				t.Fatal(err)
			}

			got, err := executetest.TablesFromCache(c)
			if err != nil {
				t.Fatal(err)
			}

			want := tc.want(tc.bounds.Start)

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(want)

			sort.Sort(executetest.SortedTables(got))
			sort.Sort(executetest.SortedTables(want))

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func windowOp(id string) plan.Node {
	return plan.CreatePhysicalNode(plan.NodeID(id), &universe.WindowProcedureSpec{})
}

func boundsOp(id string) plan.Node {
	return plan.CreatePhysicalNode(plan.NodeID(id), &universe.RangeProcedureSpec{})
}

func rangeOp(id string) plan.Node {
	return plan.CreatePhysicalNode(plan.NodeID(id), &universe.RangeProcedureSpec{})
}

func mockPred(id string) plan.Node {
	// Create a dummy physical operation that is just a wrapper around a filter.
	// Filter is one of the operators that is allowed to be a predecessor of window
	// in order to perform the trigger optimization.
	return plan.CreatePhysicalNode(plan.NodeID(id), &universe.FilterProcedureSpec{})
}

func TestWindowRewriteRule(t *testing.T) {
	testcases := []struct {
		name string
		spec *plantest.PlanSpec
		// list of rewritten window operations
		want []plan.NodeID
	}{
		// In the following test cases, the following definitions hold:
		//    w: window transformation
		//    r: range transformation
		//    b: bounded source
		//    W: window transformation with narrow trigger spec
		{
			name: "bounded source",
			// w       W
			// |       |
			// 1  ==>  1
			// |       |
			// b       b
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					mockPred("1"),
					windowOp("2"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			want: []plan.NodeID{"2"},
		},
		{
			name: "unbounded source",
			// w       w
			// |       |
			// 1  ==>  1
			// |       |
			// 0       0
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					mockPred("0"),
					mockPred("1"),
					windowOp("2"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
		{
			name: "dependent window",
			// w       w
			// |       |
			// 3       3
			// |       |
			// w  ==>  W
			// |       |
			// 1       1
			// |       |
			// b       b
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					mockPred("1"),
					windowOp("2"),
					mockPred("3"),
					windowOp("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: []plan.NodeID{"2"},
		},
		{
			name: "range after window",
			// w       W
			// |       |
			// r       r
			// |       |
			// w  ==>  W
			// |       |
			// 1       1
			// |       |
			// b       b
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					mockPred("1"),
					windowOp("2"),
					rangeOp("3"),
					windowOp("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: []plan.NodeID{"2", "4"},
		},
		{
			name: "range after window",
			// w       W
			// |       |
			// r       r
			// |       |
			// w  ==>  w
			// |       |
			// 1       1
			// |       |
			// 0       0
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					mockPred("0"),
					mockPred("1"),
					windowOp("2"),
					rangeOp("3"),
					windowOp("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: []plan.NodeID{"4"},
		},
		{
			name: "multiple sources",
			//   w           w
			//   |           |
			//   4           4
			//  / \         / \
			// w   w  ==>  W   W
			// |   |       |   |
			// b   b       b   b
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					windowOp("1"),
					boundsOp("2"),
					windowOp("3"),
					mockPred("4"),
					windowOp("5"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 4},
					{2, 3},
					{3, 4},
					{4, 5},
				},
			},
			want: []plan.NodeID{"1", "3"},
		},
		{
			name: "multiple sources",
			//   w           w
			//   |           |
			//   4           4
			//  / \         / \
			// w   w  ==>  W   w
			// |   |       |   |
			// b   2       b   2
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					windowOp("1"),
					mockPred("2"),
					windowOp("3"),
					mockPred("4"),
					windowOp("5"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 4},
					{2, 3},
					{3, 4},
					{4, 5},
				},
			},
			want: []plan.NodeID{"1"},
		},
		{
			name: "multiple sources",
			//   w           W
			//   |           |
			//   r           r
			//   |           |
			//   4           4
			//  / \         / \
			// w   w  ==>  W   w
			// |   |       |   |
			// b   2       b   2
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					boundsOp("0"),
					windowOp("1"),
					mockPred("2"),
					windowOp("3"),
					mockPred("4"),
					rangeOp("5"),
					windowOp("6"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 4},
					{2, 3},
					{3, 4},
					{4, 5},
					{5, 6},
				},
			},
			want: []plan.NodeID{"1", "6"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			spec := plantest.CreatePlanSpec(tc.spec)

			physicalPlanner := plan.NewPhysicalPlanner(
				plan.OnlyPhysicalRules(universe.WindowTriggerPhysicalRule{}),
				plan.DisableValidation(),
			)

			pp, err := physicalPlanner.Plan(spec)
			if err != nil {
				t.Fatalf("unexpected error during physical planning: %v", err)
			}

			var got []plan.NodeID
			pp.BottomUpWalk(func(node plan.Node) error {
				if _, ok := node.ProcedureSpec().(*universe.WindowProcedureSpec); !ok {
					return nil
				}
				ppn := node.(*plan.PhysicalPlanNode)
				if _, ok := ppn.TriggerSpec.(plan.NarrowTransformationTriggerSpec); !ok {
					return nil
				}
				got = append(got, node.ID())
				return nil
			})

			if !cmp.Equal(tc.want, got) {
				t.Fatalf("unexpected window trigger spec: -want/+got\n- %v\n+ %v", tc.want, got)
			}
		})
	}
}
