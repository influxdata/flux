package execute_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

func TestNewWindow(t *testing.T) {
	t.Run("normal offset", func(t *testing.T) {
		want := execute.Window{
			Every:  values.ConvertDuration(time.Minute),
			Period: values.ConvertDuration(time.Minute),
			Offset: values.ConvertDuration(time.Second),
		}
		got := MustWindow(values.ConvertDuration(time.Minute), values.ConvertDuration(time.Minute), values.ConvertDuration(time.Second))
		if !cmp.Equal(want, got) {
			t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
		}
	})

	// offset larger than "every" duration will be normalized
	t.Run("larger offset", func(t *testing.T) {
		want := execute.Window{
			Every:  values.ConvertDuration(time.Minute),
			Period: values.ConvertDuration(time.Minute),
			Offset: values.ConvertDuration(30 * time.Second),
		}
		got := MustWindow(
			values.ConvertDuration(time.Minute),
			values.ConvertDuration(time.Minute),
			values.ConvertDuration(2*time.Minute+30*time.Second))
		if !cmp.Equal(want, got) {
			t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
		}
	})

	// Negative offset will be normalized
	t.Run("negative offset", func(t *testing.T) {
		want := execute.Window{
			Every:  values.ConvertDuration(time.Minute),
			Period: values.ConvertDuration(time.Minute),
			Offset: values.ConvertDuration(30 * time.Second),
		}
		got := MustWindow(
			values.ConvertDuration(time.Minute),
			values.ConvertDuration(time.Minute),
			values.ConvertDuration(-2*time.Minute+30*time.Second))
		if !cmp.Equal(want, got) {
			t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
		}
	})

	// Mixed base duration units.
	t.Run("mixed units", func(t *testing.T) {
		wantErr := errors.New(codes.Invalid, "duration used as an interval cannot mix month and nanosecond units")
		_, gotErr := execute.NewWindow(
			mustParseDuration("1mo2w"),
			mustParseDuration("1mo2w"),
			values.Duration{},
		)
		if want, got := errAsString(wantErr), errAsString(gotErr); want != got {
			t.Errorf("window error different; -want/+got:\n%v\n", cmp.Diff(want, got))
		}
	})

	// Zero values.
	t.Run("zero values", func(t *testing.T) {
		wantErr := errors.New(codes.Invalid, "duration used as an interval cannot be zero")
		_, gotErr := execute.NewWindow(
			values.Duration{},
			values.Duration{},
			values.Duration{},
		)
		if want, got := errAsString(wantErr), errAsString(gotErr); want != got {
			t.Errorf("window error different; -want/+got:\n%v\n", cmp.Diff(want, got))
		}
	})
}

func TestWindow_GetEarliestBounds(t *testing.T) {
	var testcases = []struct {
		name string
		w    execute.Window
		t    execute.Time
		want execute.Bounds
	}{
		{
			name: "simple",
			w: MustWindow(
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(0)),
			t: execute.Time(6 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(10 * time.Minute),
			},
		},
		{
			name: "simple with offset",
			w: MustWindow(
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(30 * time.Second),
				Stop:  execute.Time(5*time.Minute + 30*time.Second),
			},
		},
		{
			name: "underlapping",
			w: MustWindow(
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(3 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(3*time.Minute + 30*time.Second),
				Stop:  execute.Time(4*time.Minute + 30*time.Second),
			},
		},
		{
			name: "underlapping not contained",
			w: MustWindow(
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(2*time.Minute + 45*time.Second),
			want: execute.Bounds{
				Start: execute.Time(3*time.Minute + 30*time.Second),
				Stop:  execute.Time(4*time.Minute + 30*time.Second),
			},
		},
		{
			name: "overlapping",
			w: MustWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(30 * time.Second),
			want: execute.Bounds{
				Start: execute.Time(-30 * time.Second),
				Stop:  execute.Time(1*time.Minute + 30*time.Second),
			},
		},
		{
			name: "partially overlapping",
			w: MustWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(3*time.Minute+30*time.Second),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5*time.Minute + 45*time.Second),
			want: execute.Bounds{
				Start: execute.Time(3 * time.Minute),
				Stop:  execute.Time(6*time.Minute + 30*time.Second),
			},
		},
		{
			name: "partially overlapping (t on boundary)",
			w: MustWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(3*time.Minute+30*time.Second),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(2 * time.Minute),
				Stop:  execute.Time(5*time.Minute + 30*time.Second),
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.w.GetEarliestBounds(tc.t)
			if !cmp.Equal(tc.want, got) {
				t.Errorf("did not get expected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestWindow_GetOverlappingBounds(t *testing.T) {
	ts, ds := mustParseTime, mustParseDuration
	testcases := []struct {
		name string
		w    execute.Window
		b    execute.Bounds
		want []execute.Bounds
	}{
		{
			name: "simple",
			w: execute.Window{
				Every:  values.ConvertDuration(time.Minute),
				Period: values.ConvertDuration(time.Minute),
			},
			b: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(8 * time.Minute),
			},
			want: []execute.Bounds{
				{Start: execute.Time(5 * time.Minute), Stop: execute.Time(6 * time.Minute)},
				{Start: execute.Time(6 * time.Minute), Stop: execute.Time(7 * time.Minute)},
				{Start: execute.Time(7 * time.Minute), Stop: execute.Time(8 * time.Minute)},
			},
		},
		{
			name: "simple with offset",
			w: execute.Window{
				Every:  values.ConvertDuration(time.Minute),
				Period: values.ConvertDuration(time.Minute),
				Offset: values.ConvertDuration(15 * time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(7 * time.Minute),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(4*time.Minute + 15*time.Second),
					Stop:  execute.Time(5*time.Minute + 15*time.Second),
				},
				{
					Start: execute.Time(5*time.Minute + 15*time.Second),
					Stop:  execute.Time(6*time.Minute + 15*time.Second),
				},
				{
					Start: execute.Time(6*time.Minute + 15*time.Second),
					Stop:  execute.Time(7*time.Minute + 15*time.Second),
				},
			},
		},
		{
			name: "underlapping, bounds in gap",
			w: execute.Window{
				Every:  values.ConvertDuration(2 * time.Minute),
				Period: values.ConvertDuration(time.Minute),
			},
			b: execute.Bounds{
				Start: execute.Time(30 * time.Second),
				Stop:  execute.Time(45 * time.Second),
			},
			want: []execute.Bounds{},
		},
		{
			name: "underlapping",
			w: execute.Window{
				Every:  values.ConvertDuration(2 * time.Minute),
				Period: values.ConvertDuration(time.Minute),
				Offset: values.ConvertDuration(30 * time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(time.Minute + 45*time.Second),
				Stop:  execute.Time(4*time.Minute + 35*time.Second),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(1*time.Minute + 30*time.Second),
					Stop:  execute.Time(2*time.Minute + 30*time.Second),
				},
				{
					Start: execute.Time(3*time.Minute + 30*time.Second),
					Stop:  execute.Time(4*time.Minute + 30*time.Second),
				},
			},
		},
		{
			name: "overlapping",
			w: execute.Window{
				Every:  values.ConvertDuration(1 * time.Minute),
				Period: values.ConvertDuration(2*time.Minute + 15*time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(10 * time.Minute),
				Stop:  execute.Time(12 * time.Minute),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(8*time.Minute + 45*time.Second),
					Stop:  execute.Time(11 * time.Minute),
				},
				{
					Start: execute.Time(9*time.Minute + 45*time.Second),
					Stop:  execute.Time(12 * time.Minute),
				},
				{
					Start: execute.Time(10*time.Minute + 45*time.Second),
					Stop:  execute.Time(13 * time.Minute),
				},
				{
					Start: execute.Time(11*time.Minute + 45*time.Second),
					Stop:  execute.Time(14 * time.Minute),
				},
			},
		},
		{
			name: "by day",
			b: execute.Bounds{
				Start: ts("2019-10-01T00:00:00Z"),
				Stop:  ts("2019-10-08T00:00:00Z"),
			},
			w: execute.Window{
				Every:  ds("1d"),
				Period: ds("1d"),
			},
			want: []execute.Bounds{
				{Start: ts("2019-10-01T00:00:00Z"), Stop: ts("2019-10-02T00:00:00Z")},
				{Start: ts("2019-10-02T00:00:00Z"), Stop: ts("2019-10-03T00:00:00Z")},
				{Start: ts("2019-10-03T00:00:00Z"), Stop: ts("2019-10-04T00:00:00Z")},
				{Start: ts("2019-10-04T00:00:00Z"), Stop: ts("2019-10-05T00:00:00Z")},
				{Start: ts("2019-10-05T00:00:00Z"), Stop: ts("2019-10-06T00:00:00Z")},
				{Start: ts("2019-10-06T00:00:00Z"), Stop: ts("2019-10-07T00:00:00Z")},
				{Start: ts("2019-10-07T00:00:00Z"), Stop: ts("2019-10-08T00:00:00Z")},
			},
		},
		{
			name: "by month",
			b: execute.Bounds{
				Start: ts("2019-01-01T00:00:00Z"),
				Stop:  ts("2020-01-01T00:00:00Z"),
			},
			w: execute.Window{
				Every:  ds("1mo"),
				Period: ds("1mo"),
			},
			want: []execute.Bounds{
				{Start: ts("2019-01-01T00:00:00Z"), Stop: ts("2019-02-01T00:00:00Z")},
				{Start: ts("2019-02-01T00:00:00Z"), Stop: ts("2019-03-01T00:00:00Z")},
				{Start: ts("2019-03-01T00:00:00Z"), Stop: ts("2019-04-01T00:00:00Z")},
				{Start: ts("2019-04-01T00:00:00Z"), Stop: ts("2019-05-01T00:00:00Z")},
				{Start: ts("2019-05-01T00:00:00Z"), Stop: ts("2019-06-01T00:00:00Z")},
				{Start: ts("2019-06-01T00:00:00Z"), Stop: ts("2019-07-01T00:00:00Z")},
				{Start: ts("2019-07-01T00:00:00Z"), Stop: ts("2019-08-01T00:00:00Z")},
				{Start: ts("2019-08-01T00:00:00Z"), Stop: ts("2019-09-01T00:00:00Z")},
				{Start: ts("2019-09-01T00:00:00Z"), Stop: ts("2019-10-01T00:00:00Z")},
				{Start: ts("2019-10-01T00:00:00Z"), Stop: ts("2019-11-01T00:00:00Z")},
				{Start: ts("2019-11-01T00:00:00Z"), Stop: ts("2019-12-01T00:00:00Z")},
				{Start: ts("2019-12-01T00:00:00Z"), Stop: ts("2020-01-01T00:00:00Z")},
			},
		},
		{
			name: "overlapping by month",
			b: execute.Bounds{
				Start: ts("2019-01-01T00:00:00Z"),
				Stop:  ts("2020-01-01T00:00:00Z"),
			},
			w: execute.Window{
				Every:  ds("1mo"),
				Period: ds("3mo"),
			},
			want: []execute.Bounds{
				{Start: ts("2018-11-01T00:00:00Z"), Stop: ts("2019-02-01T00:00:00Z")},
				{Start: ts("2018-12-01T00:00:00Z"), Stop: ts("2019-03-01T00:00:00Z")},
				{Start: ts("2019-01-01T00:00:00Z"), Stop: ts("2019-04-01T00:00:00Z")},
				{Start: ts("2019-02-01T00:00:00Z"), Stop: ts("2019-05-01T00:00:00Z")},
				{Start: ts("2019-03-01T00:00:00Z"), Stop: ts("2019-06-01T00:00:00Z")},
				{Start: ts("2019-04-01T00:00:00Z"), Stop: ts("2019-07-01T00:00:00Z")},
				{Start: ts("2019-05-01T00:00:00Z"), Stop: ts("2019-08-01T00:00:00Z")},
				{Start: ts("2019-06-01T00:00:00Z"), Stop: ts("2019-09-01T00:00:00Z")},
				{Start: ts("2019-07-01T00:00:00Z"), Stop: ts("2019-10-01T00:00:00Z")},
				{Start: ts("2019-08-01T00:00:00Z"), Stop: ts("2019-11-01T00:00:00Z")},
				{Start: ts("2019-09-01T00:00:00Z"), Stop: ts("2019-12-01T00:00:00Z")},
				{Start: ts("2019-10-01T00:00:00Z"), Stop: ts("2020-01-01T00:00:00Z")},
				{Start: ts("2019-11-01T00:00:00Z"), Stop: ts("2020-02-01T00:00:00Z")},
				{Start: ts("2019-12-01T00:00:00Z"), Stop: ts("2020-03-01T00:00:00Z")},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.w.GetOverlappingBounds(tc.b)
			if !cmp.Equal(tc.want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestWindow_StartStop(t *testing.T) {
	testcases := []struct {
		name      string
		every     values.Duration
		offset    values.Duration
		timestamp values.Time
		wantStart values.Time
		wantStop  values.Time
	}{
		{
			name:      "one second",
			every:     values.CreateDuration(1, 1000000000, false),
			offset:    values.CreateDuration(1, 1000000000, false),
			timestamp: values.Time(34214401000000000),
			wantStart: values.Time(34214401000000000),
			wantStop:  values.Time(34214402000000000),
		},
		{
			name:      "one thousand seconds",
			every:     values.CreateDuration(1, 1000000000000, false),
			offset:    values.CreateDuration(1, 1000000000000, false),
			timestamp: values.Time(35214401000000000),
			wantStart: values.Time(35214000000000000),
			wantStop:  values.Time(35215000000000000),
		},
		{
			name:      "negative duration",
			every:     values.CreateDuration(1, 10, true),
			offset:    values.CreateDuration(0, 0, true),
			timestamp: values.Time(1000),
			wantStart: values.Time(1000),
			wantStop:  values.Time(1010),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := execute.WindowStart(tc.timestamp, tc.every, tc.offset)
			if !cmp.Equal(tc.wantStart, got) {
				t.Errorf("got unexpected window start; -want/+got:\n%v\n", cmp.Diff(tc.wantStart, got))
			}

			got = execute.WindowStop(tc.timestamp, tc.every, tc.offset)
			if !cmp.Equal(tc.wantStop, got) {
				t.Errorf("got unexpected window stop; -want/+got:\n%v\n", cmp.Diff(tc.wantStop, got))
			}
		})
	}
}

func MustWindow(every, period, offset execute.Duration) execute.Window {
	w, err := execute.NewWindow(every, period, offset)
	if err != nil {
		panic(err)
	}
	return w
}

func mustParseTime(s string) execute.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return values.ConvertTime(t)
}

func mustParseDuration(s string) execute.Duration {
	d, err := values.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func errAsString(err error) (s string) {
	if err != nil {
		s = err.Error()
	}
	return s
}
