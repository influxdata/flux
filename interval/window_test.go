package interval_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/values"
)

func TestNewWindow(t *testing.T) {
	var testcases = []struct {
		name    string
		every   values.Duration
		period  values.Duration
		offset  values.Duration
		wantErr bool
	}{
		{
			name:   "valid nanoseconds every",
			every:  values.ConvertDurationNsecs(time.Minute),
			period: values.ConvertDurationNsecs(time.Minute),
			offset: values.ConvertDurationNsecs(time.Minute),
		},
		{
			name:   "valid months every",
			every:  values.MakeDuration(0, 1, false),
			period: values.ConvertDurationNsecs(time.Minute),
			offset: values.ConvertDurationNsecs(time.Minute),
		},
		{
			name:    "invalid zero every",
			every:   values.ConvertDurationNsecs(0),
			period:  values.ConvertDurationNsecs(time.Minute),
			offset:  values.ConvertDurationNsecs(time.Minute),
			wantErr: true,
		},
		{
			name:    "invalid mixed every",
			every:   values.MakeDuration(1, 1, false),
			period:  values.ConvertDurationNsecs(time.Minute),
			offset:  values.ConvertDurationNsecs(time.Minute),
			wantErr: true,
		},
		{
			name:    "invalid negative every",
			every:   values.MakeDuration(0, 1, true),
			period:  values.ConvertDurationNsecs(time.Minute),
			offset:  values.ConvertDurationNsecs(time.Minute),
			wantErr: true,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := interval.NewWindow(tc.every, tc.period, tc.offset)
			hasErr := err != nil
			if tc.wantErr != hasErr {
				if tc.wantErr {
					t.Error("missing expected error")
				} else {
					t.Errorf("unexpected error: %s", err)
				}
			}
		})
	}
}

type testBounds struct {
	Start values.Time
	Stop  values.Time
}

func TestWindow_GetLatestBounds(t *testing.T) {
	var testcases = []struct {
		name string
		w    interval.Window
		t    values.Time
		want testBounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0)),
			t: values.Time(6 * time.Minute),
			want: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(10 * time.Minute),
			},
		},
		{
			name: "simple with negative period",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-5*time.Minute),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(5 * time.Minute),
			want: testBounds{
				Start: values.Time(30 * time.Second),
				Stop:  values.Time(5*time.Minute + 30*time.Second),
			},
		},
		{
			name: "simple with offset",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(5 * time.Minute),
			want: testBounds{
				Start: values.Time(30 * time.Second),
				Stop:  values.Time(5*time.Minute + 30*time.Second),
			},
		},
		{
			name: "simple with negative offset",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-30*time.Second)),
			t: values.Time(5 * time.Minute),
			want: testBounds{
				Start: values.Time(4*time.Minute + 30*time.Second),
				Stop:  values.Time(9*time.Minute + 30*time.Second),
			},
		},
		{
			name: "simple with equal offset before",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute)),
			t: values.Time(0),
			want: testBounds{
				Start: values.Time(0 * time.Minute),
				Stop:  values.Time(5 * time.Minute),
			},
		},
		{
			name: "simple with equal offset after",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute)),
			t: values.Time(7 * time.Minute),
			want: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(10 * time.Minute),
			},
		},
		{
			name: "simple months",
			w: mustWindow(
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(0)),
			t: mustTime("1970-01-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-01-01T00:00:00Z"),
				Stop:  mustTime("1970-06-01T00:00:00Z"),
			},
		},
		{
			name: "simple months with offset",
			w: mustWindow(
				values.ConvertDurationMonths(3),
				values.ConvertDurationMonths(3),
				values.ConvertDurationMonths(1)),
			t: mustTime("1970-01-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1969-11-01T00:00:00Z"),
				Stop:  mustTime("1970-02-01T00:00:00Z"),
			},
		},
		{
			name: "months with equal offset",
			w: mustWindow(
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5)),
			t: mustTime("1970-01-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-01-01T00:00:00Z"),
				Stop:  mustTime("1970-06-01T00:00:00Z"),
			},
		},
		{
			name: "underlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(3 * time.Minute),
			want: testBounds{
				Start: values.Time(2*time.Minute + 30*time.Second),
				Stop:  values.Time(3*time.Minute + 30*time.Second),
			},
		},
		{
			name: "underlapping not contained",
			w: mustWindow(
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(2*time.Minute + 15*time.Second),
			want: testBounds{
				Start: values.Time(0*time.Minute + 30*time.Second),
				Stop:  values.Time(1*time.Minute + 30*time.Second),
			},
		},
		{
			name: "overlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(30 * time.Second),
			want: testBounds{
				Start: values.Time(30 * time.Second),
				Stop:  values.Time(2*time.Minute + 30*time.Second),
			},
		},
		{
			name: "partially overlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(3*time.Minute+30*time.Second),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(5*time.Minute + 45*time.Second),
			want: testBounds{
				Start: values.Time(5*time.Minute + 30*time.Second),
				Stop:  values.Time(9 * time.Minute),
			},
		},
		{
			name: "partially overlapping (t on boundary)",
			w: mustWindow(
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(3*time.Minute+30*time.Second),
				values.ConvertDurationNsecs(30*time.Second)),
			t: values.Time(5 * time.Minute),
			want: testBounds{
				Start: values.Time(4*time.Minute + 30*time.Second),
				Stop:  values.Time(8 * time.Minute),
			},
		},
		{
			name: "overlapping with negative period on boundary",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-15*time.Minute),
				values.ConvertDurationNsecs(0*time.Second)),
			t: values.Time(5 * time.Minute),
			want: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(20 * time.Minute),
			},
		},
		{
			name: "overlapping with negative period",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-15*time.Minute),
				values.ConvertDurationNsecs(0*time.Second)),
			t: values.Time(6 * time.Minute),
			want: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(20 * time.Minute),
			},
		},
		{
			name: "truncate before offset",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Second),
				values.ConvertDurationNsecs(5*time.Second),
				values.ConvertDurationNsecs(2*time.Second)),
			t: values.Time(1 * time.Second),
			want: testBounds{
				Start: values.Time(-3 * time.Second),
				Stop:  values.Time(2 * time.Second),
			},
		},
		{
			name: "truncate after offset",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Second),
				values.ConvertDurationNsecs(5*time.Second),
				values.ConvertDurationNsecs(2*time.Second)),
			t: values.Time(3 * time.Second),
			want: testBounds{
				Start: values.Time(2 * time.Second),
				Stop:  values.Time(7 * time.Second),
			},
		},
		{
			name: "truncate before calendar offset",
			w: mustWindow(
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(2)),
			t: mustTime("1970-02-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1969-10-01T00:00:00Z"),
				Stop:  mustTime("1970-03-01T00:00:00Z"),
			},
		},
		{
			name: "truncate after calendar offset",
			w: mustWindow(
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(2)),
			t: mustTime("1970-04-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-03-01T00:00:00Z"),
				Stop:  mustTime("1970-08-01T00:00:00Z"),
			},
		},
		{
			name: "negative calendar offset",
			w: mustWindow(
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(5),
				values.ConvertDurationMonths(-2)),
			t: mustTime("1970-02-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1969-11-01T00:00:00Z"),
				Stop:  mustTime("1970-04-01T00:00:00Z"),
			},
		},
		{
			name: "calendar overlapping with negative period on boundary",
			w: mustWindow(
				values.ConvertDurationMonths(4),
				values.ConvertDurationMonths(-10),
				values.ConvertDurationMonths(0)),
			t: mustTime("1970-03-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-03-01T00:00:00Z"),
				Stop:  mustTime("1971-01-01T00:00:00Z"),
			},
		},
		{
			name: "calendar overlapping with negative period",
			w: mustWindow(
				values.ConvertDurationMonths(4),
				values.ConvertDurationMonths(-10),
				values.ConvertDurationMonths(0)),
			t: mustTime("1970-03-01T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-03-01T00:00:00Z"),
				Stop:  mustTime("1971-01-01T00:00:00Z"),
			},
		},
		{
			name: "mixed period",
			w: mustWindow(
				values.ConvertDurationMonths(2),
				values.MakeDuration(int64(10*time.Hour), 1, false),
				values.ConvertDurationNsecs(0)),
			t: mustTime("1970-07-10T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-07-01T00:00:00Z"),
				Stop:  mustTime("1970-08-01T10:00:00Z"),
			},
		},
		{
			name: "mixed negative period",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.MakeDuration(int64(24*time.Hour), 1, true),
				values.ConvertDurationNsecs(0)),
			t: mustTime("1970-07-10T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-06-30T00:00:00Z"),
				Stop:  mustTime("1970-08-01T00:00:00Z"),
			},
		},
		{
			name: "mixed offset",
			w: mustWindow(
				values.ConvertDurationMonths(2),
				values.ConvertDurationMonths(2),
				values.MakeDuration(int64(10*time.Hour), 1, false),
			),
			t: mustTime("1970-07-10T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-06-01T10:00:00Z"),
				Stop:  mustTime("1970-08-01T10:00:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset before by days",
			w: mustWindow(
				values.ConvertDurationMonths(2),
				values.ConvertDurationMonths(2),
				values.MakeDuration(int64(24*time.Hour), 1, true),
			),
			t: mustTime("1970-07-10T00:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-05-30T00:00:00Z"),
				Stop:  mustTime("1970-07-30T00:00:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset before by hours",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Hour),
			),
			t: mustTime("1970-07-31T21:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-06-30T22:00:00Z"),
				Stop:  mustTime("1970-07-30T22:00:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset after by hours",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Hour),
			),
			t: mustTime("1970-07-31T23:00:00Z"),
			want: testBounds{
				Start: mustTime("1970-07-31T22:00:00Z"),
				Stop:  mustTime("1970-08-31T22:00:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset before by minutes",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Minute),
			),
			t: mustTime("1970-07-31T23:57:00Z"),
			want: testBounds{
				Start: mustTime("1970-06-30T23:58:00Z"),
				Stop:  mustTime("1970-07-30T23:58:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset after by minutes",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Minute),
			),
			t: mustTime("1970-07-31T23:59:00Z"),
			want: testBounds{
				Start: mustTime("1970-07-31T23:58:00Z"),
				Stop:  mustTime("1970-08-31T23:58:00Z"),
			},
		},
		{
			name: "calendar mixed negative offset before by seconds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Second),
			),
			t: mustTime("1970-07-31T23:59:57Z"),
			want: testBounds{
				Start: mustTime("1970-06-30T23:59:58Z"),
				Stop:  mustTime("1970-07-30T23:59:58Z"),
			},
		},
		{
			name: "calendar mixed negative offset after by seconds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2*time.Second),
			),
			t: mustTime("1970-07-31T23:59:59Z"),
			want: testBounds{
				Start: mustTime("1970-07-31T23:59:58Z"),
				Stop:  mustTime("1970-08-31T23:59:58Z"),
			},
		},
		{
			name: "calendar mixed negative offset before by nanoseconds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2),
			),
			t: mustTime("1970-07-31T23:59:59.999999997Z"),
			want: testBounds{
				Start: mustTime("1970-06-30T23:59:59.999999998Z"),
				Stop:  mustTime("1970-07-30T23:59:59.999999998Z"),
			},
		},
		{
			name: "calendar mixed negative offset after by nanoseconds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2),
			),
			t: mustTime("1970-07-31T23:59:59.999999999Z"),
			want: testBounds{
				Start: mustTime("1970-07-31T23:59:59.999999998Z"),
				Stop:  mustTime("1970-08-31T23:59:59.999999998Z"),
			},
		},
		{
			name: "calendar mixed negative offset equal to nanoseconds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-2),
			),
			t: mustTime("1970-07-31T23:59:59.999999998Z"),
			want: testBounds{
				Start: mustTime("1970-07-31T23:59:59.999999998Z"),
				Stop:  mustTime("1970-08-31T23:59:59.999999998Z"),
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.w.GetLatestBounds(tc.t)
			if got.Start() != tc.want.Start {
				t.Errorf("unexpected start boundary: got %s want %s", got.Start(), tc.want.Start)
			}
			if got.Stop() != tc.want.Stop {
				t.Errorf("unexpected stop boundary:  got %s want %s", got.Stop(), tc.want.Stop)
			}
		})
	}
}

func TestWindow_GetOverlappingBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    interval.Window
		b    testBounds
		want []testBounds
	}{
		{
			name: "empty",
			w: mustWindow(
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(0),
			),
			b: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(5 * time.Minute),
			},
			want: []testBounds{},
		},
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(0),
			),
			b: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(8 * time.Minute),
			},
			want: []testBounds{
				{Start: values.Time(7 * time.Minute), Stop: values.Time(8 * time.Minute)},
				{Start: values.Time(6 * time.Minute), Stop: values.Time(7 * time.Minute)},
				{Start: values.Time(5 * time.Minute), Stop: values.Time(6 * time.Minute)},
			},
		},
		{
			name: "simple with offset",
			w: mustWindow(
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(15*time.Second),
			),
			b: testBounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(7 * time.Minute),
			},
			want: []testBounds{
				{
					Start: values.Time(6*time.Minute + 15*time.Second),
					Stop:  values.Time(7*time.Minute + 15*time.Second),
				},
				{
					Start: values.Time(5*time.Minute + 15*time.Second),
					Stop:  values.Time(6*time.Minute + 15*time.Second),
				},
				{
					Start: values.Time(4*time.Minute + 15*time.Second),
					Stop:  values.Time(5*time.Minute + 15*time.Second),
				},
			},
		},
		{
			name: "underlapping, bounds in gap",
			w: mustWindow(
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(0),
			),
			b: testBounds{
				Start: values.Time(1*time.Minute + 30*time.Second),
				Stop:  values.Time(1*time.Minute + 45*time.Second),
			},
			want: []testBounds{},
		},
		{
			name: "underlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(30*time.Second),
			),
			b: testBounds{
				Start: values.Time(1*time.Minute + 45*time.Second),
				Stop:  values.Time(4*time.Minute + 35*time.Second),
			},
			want: []testBounds{
				{
					Start: values.Time(4*time.Minute + 30*time.Second),
					Stop:  values.Time(5*time.Minute + 30*time.Second),
				},
				{
					Start: values.Time(2*time.Minute + 30*time.Second),
					Stop:  values.Time(3*time.Minute + 30*time.Second),
				},
			},
		},
		{
			name: "overlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(1*time.Minute),
				values.ConvertDurationNsecs(2*time.Minute+15*time.Second),
				values.ConvertDurationNsecs(0),
			),
			b: testBounds{
				Start: values.Time(10 * time.Minute),
				Stop:  values.Time(12 * time.Minute),
			},
			want: []testBounds{
				{
					Start: values.Time(11 * time.Minute),
					Stop:  values.Time(13*time.Minute + 15*time.Second),
				},
				{
					Start: values.Time(10 * time.Minute),
					Stop:  values.Time(12*time.Minute + 15*time.Second),
				},
				{
					Start: values.Time(9 * time.Minute),
					Stop:  values.Time(11*time.Minute + 15*time.Second),
				},
				{
					Start: values.Time(8 * time.Minute),
					Stop:  values.Time(10*time.Minute + 15*time.Second),
				},
			},
		},
		{
			name: "by day",
			b: testBounds{
				Start: mustTime("2019-10-01T00:00:00Z"),
				Stop:  mustTime("2019-10-08T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1d"),
				mustDuration("1d"),
				values.ConvertDurationNsecs(0),
			),
			want: []testBounds{
				{Start: mustTime("2019-10-07T00:00:00Z"), Stop: mustTime("2019-10-08T00:00:00Z")},
				{Start: mustTime("2019-10-06T00:00:00Z"), Stop: mustTime("2019-10-07T00:00:00Z")},
				{Start: mustTime("2019-10-05T00:00:00Z"), Stop: mustTime("2019-10-06T00:00:00Z")},
				{Start: mustTime("2019-10-04T00:00:00Z"), Stop: mustTime("2019-10-05T00:00:00Z")},
				{Start: mustTime("2019-10-03T00:00:00Z"), Stop: mustTime("2019-10-04T00:00:00Z")},
				{Start: mustTime("2019-10-02T00:00:00Z"), Stop: mustTime("2019-10-03T00:00:00Z")},
				{Start: mustTime("2019-10-01T00:00:00Z"), Stop: mustTime("2019-10-02T00:00:00Z")},
			},
		},
		{
			name: "by month",
			b: testBounds{
				Start: mustTime("2019-01-01T00:00:00Z"),
				Stop:  mustTime("2020-01-01T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1mo"),
				mustDuration("1mo"),
				values.ConvertDurationNsecs(0),
			),
			want: []testBounds{
				{Start: mustTime("2019-12-01T00:00:00Z"), Stop: mustTime("2020-01-01T00:00:00Z")},
				{Start: mustTime("2019-11-01T00:00:00Z"), Stop: mustTime("2019-12-01T00:00:00Z")},
				{Start: mustTime("2019-10-01T00:00:00Z"), Stop: mustTime("2019-11-01T00:00:00Z")},
				{Start: mustTime("2019-09-01T00:00:00Z"), Stop: mustTime("2019-10-01T00:00:00Z")},
				{Start: mustTime("2019-08-01T00:00:00Z"), Stop: mustTime("2019-09-01T00:00:00Z")},
				{Start: mustTime("2019-07-01T00:00:00Z"), Stop: mustTime("2019-08-01T00:00:00Z")},
				{Start: mustTime("2019-06-01T00:00:00Z"), Stop: mustTime("2019-07-01T00:00:00Z")},
				{Start: mustTime("2019-05-01T00:00:00Z"), Stop: mustTime("2019-06-01T00:00:00Z")},
				{Start: mustTime("2019-04-01T00:00:00Z"), Stop: mustTime("2019-05-01T00:00:00Z")},
				{Start: mustTime("2019-03-01T00:00:00Z"), Stop: mustTime("2019-04-01T00:00:00Z")},
				{Start: mustTime("2019-02-01T00:00:00Z"), Stop: mustTime("2019-03-01T00:00:00Z")},
				{Start: mustTime("2019-01-01T00:00:00Z"), Stop: mustTime("2019-02-01T00:00:00Z")},
			},
		},
		{
			name: "overlapping by month",
			b: testBounds{
				Start: mustTime("2019-01-01T00:00:00Z"),
				Stop:  mustTime("2020-01-01T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1mo"),
				mustDuration("3mo"),
				values.ConvertDurationNsecs(0),
			),
			want: []testBounds{
				{Start: mustTime("2019-12-01T00:00:00Z"), Stop: mustTime("2020-03-01T00:00:00Z")},
				{Start: mustTime("2019-11-01T00:00:00Z"), Stop: mustTime("2020-02-01T00:00:00Z")},
				{Start: mustTime("2019-10-01T00:00:00Z"), Stop: mustTime("2020-01-01T00:00:00Z")},
				{Start: mustTime("2019-09-01T00:00:00Z"), Stop: mustTime("2019-12-01T00:00:00Z")},
				{Start: mustTime("2019-08-01T00:00:00Z"), Stop: mustTime("2019-11-01T00:00:00Z")},
				{Start: mustTime("2019-07-01T00:00:00Z"), Stop: mustTime("2019-10-01T00:00:00Z")},
				{Start: mustTime("2019-06-01T00:00:00Z"), Stop: mustTime("2019-09-01T00:00:00Z")},
				{Start: mustTime("2019-05-01T00:00:00Z"), Stop: mustTime("2019-08-01T00:00:00Z")},
				{Start: mustTime("2019-04-01T00:00:00Z"), Stop: mustTime("2019-07-01T00:00:00Z")},
				{Start: mustTime("2019-03-01T00:00:00Z"), Stop: mustTime("2019-06-01T00:00:00Z")},
				{Start: mustTime("2019-02-01T00:00:00Z"), Stop: mustTime("2019-05-01T00:00:00Z")},
				{Start: mustTime("2019-01-01T00:00:00Z"), Stop: mustTime("2019-04-01T00:00:00Z")},
				{Start: mustTime("2018-12-01T00:00:00Z"), Stop: mustTime("2019-03-01T00:00:00Z")},
				{Start: mustTime("2018-11-01T00:00:00Z"), Stop: mustTime("2019-02-01T00:00:00Z")},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := transformBounds(tc.w.GetOverlappingBounds(tc.b.Start, tc.b.Stop))
			if !cmp.Equal(tc.want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}
func TestWindow_NextBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    interval.Window
		t    values.Time
		want []testBounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(10 * time.Minute),
			want: []testBounds{
				{Start: values.Time(10 * time.Minute), Stop: values.Time(15 * time.Minute)},
				{Start: values.Time(15 * time.Minute), Stop: values.Time(20 * time.Minute)},
				{Start: values.Time(20 * time.Minute), Stop: values.Time(25 * time.Minute)},
				{Start: values.Time(25 * time.Minute), Stop: values.Time(30 * time.Minute)},
				{Start: values.Time(30 * time.Minute), Stop: values.Time(35 * time.Minute)},
				{Start: values.Time(35 * time.Minute), Stop: values.Time(40 * time.Minute)},
			},
		},
		{
			name: "simple negative period",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(10 * time.Minute),
			want: []testBounds{
				{Start: values.Time(10 * time.Minute), Stop: values.Time(15 * time.Minute)},
				{Start: values.Time(15 * time.Minute), Stop: values.Time(20 * time.Minute)},
				{Start: values.Time(20 * time.Minute), Stop: values.Time(25 * time.Minute)},
				{Start: values.Time(25 * time.Minute), Stop: values.Time(30 * time.Minute)},
				{Start: values.Time(30 * time.Minute), Stop: values.Time(35 * time.Minute)},
				{Start: values.Time(35 * time.Minute), Stop: values.Time(40 * time.Minute)},
			},
		},
		{
			name: "beginning of month",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(0),
			),
			t: mustTime("2020-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2020-10-01T00:00:00Z"), Stop: mustTime("2020-11-01T00:00:00Z")},
				{Start: mustTime("2020-11-01T00:00:00Z"), Stop: mustTime("2020-12-01T00:00:00Z")},
				{Start: mustTime("2020-12-01T00:00:00Z"), Stop: mustTime("2021-01-01T00:00:00Z")},
				{Start: mustTime("2021-01-01T00:00:00Z"), Stop: mustTime("2021-02-01T00:00:00Z")},
				{Start: mustTime("2021-02-01T00:00:00Z"), Stop: mustTime("2021-03-01T00:00:00Z")},
				{Start: mustTime("2021-03-01T00:00:00Z"), Stop: mustTime("2021-04-01T00:00:00Z")},
				{Start: mustTime("2021-04-01T00:00:00Z"), Stop: mustTime("2021-05-01T00:00:00Z")},
				{Start: mustTime("2021-05-01T00:00:00Z"), Stop: mustTime("2021-06-01T00:00:00Z")},
				{Start: mustTime("2021-06-01T00:00:00Z"), Stop: mustTime("2021-07-01T00:00:00Z")},
			},
		},
		{
			name: "end of month",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-24*time.Hour),
			),
			t: mustTime("2020-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2020-09-30T00:00:00Z"), Stop: mustTime("2020-10-30T00:00:00Z")},
				{Start: mustTime("2020-10-31T00:00:00Z"), Stop: mustTime("2020-11-30T00:00:00Z")},
				{Start: mustTime("2020-11-30T00:00:00Z"), Stop: mustTime("2020-12-30T00:00:00Z")},
				{Start: mustTime("2020-12-31T00:00:00Z"), Stop: mustTime("2021-01-31T00:00:00Z")},
				{Start: mustTime("2021-01-31T00:00:00Z"), Stop: mustTime("2021-02-28T00:00:00Z")},
				{Start: mustTime("2021-02-28T00:00:00Z"), Stop: mustTime("2021-03-28T00:00:00Z")},
				// This is the case that is fixed by adding index.
				// If we were to simply add a month to 2-28 the next window would start on 3-28 instead of 3-31.
				{Start: mustTime("2021-03-31T00:00:00Z"), Stop: mustTime("2021-04-30T00:00:00Z")},
				{Start: mustTime("2021-04-30T00:00:00Z"), Stop: mustTime("2021-05-30T00:00:00Z")},
				{Start: mustTime("2021-05-31T00:00:00Z"), Stop: mustTime("2021-06-30T00:00:00Z")},
				{Start: mustTime("2021-06-30T00:00:00Z"), Stop: mustTime("2021-07-30T00:00:00Z")},
				{Start: mustTime("2021-07-31T00:00:00Z"), Stop: mustTime("2021-08-31T00:00:00Z")},
				{Start: mustTime("2021-08-31T00:00:00Z"), Stop: mustTime("2021-09-30T00:00:00Z")},
			},
		},
		{
			name: "end of month far from bounds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-24*time.Hour),
			),
			t: mustTime("2121-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2121-09-30T00:00:00Z"), Stop: mustTime("2121-10-30T00:00:00Z")},
				{Start: mustTime("2121-10-31T00:00:00Z"), Stop: mustTime("2121-11-30T00:00:00Z")},
				{Start: mustTime("2121-11-30T00:00:00Z"), Stop: mustTime("2121-12-30T00:00:00Z")},
				{Start: mustTime("2121-12-31T00:00:00Z"), Stop: mustTime("2122-01-31T00:00:00Z")},
				{Start: mustTime("2122-01-31T00:00:00Z"), Stop: mustTime("2122-02-28T00:00:00Z")},
				{Start: mustTime("2122-02-28T00:00:00Z"), Stop: mustTime("2122-03-28T00:00:00Z")},
				{Start: mustTime("2122-03-31T00:00:00Z"), Stop: mustTime("2122-04-30T00:00:00Z")},
				{Start: mustTime("2122-04-30T00:00:00Z"), Stop: mustTime("2122-05-30T00:00:00Z")},
				{Start: mustTime("2122-05-31T00:00:00Z"), Stop: mustTime("2122-06-30T00:00:00Z")},
				{Start: mustTime("2122-06-30T00:00:00Z"), Stop: mustTime("2122-07-30T00:00:00Z")},
				{Start: mustTime("2122-07-31T00:00:00Z"), Stop: mustTime("2122-08-31T00:00:00Z")},
				{Start: mustTime("2122-08-31T00:00:00Z"), Stop: mustTime("2122-09-30T00:00:00Z")},
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b := tc.w.GetLatestBounds(tc.t)
			got := make([]testBounds, 0, len(tc.want))
			for range tc.want {
				got = append(got, testBounds{
					Start: b.Start(),
					Stop:  b.Stop(),
				})
				b = tc.w.NextBounds(b)
			}
			if !cmp.Equal(tc.want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}
func TestWindow_PrevBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    interval.Window
		t    values.Time
		want []testBounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(36 * time.Minute),
			want: []testBounds{
				{Start: values.Time(35 * time.Minute), Stop: values.Time(40 * time.Minute)},
				{Start: values.Time(30 * time.Minute), Stop: values.Time(35 * time.Minute)},
				{Start: values.Time(25 * time.Minute), Stop: values.Time(30 * time.Minute)},
				{Start: values.Time(20 * time.Minute), Stop: values.Time(25 * time.Minute)},
				{Start: values.Time(15 * time.Minute), Stop: values.Time(20 * time.Minute)},
				{Start: values.Time(10 * time.Minute), Stop: values.Time(15 * time.Minute)},
			},
		},
		{
			name: "simple negative period",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(-5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(36 * time.Minute),
			want: []testBounds{
				{Start: values.Time(35 * time.Minute), Stop: values.Time(40 * time.Minute)},
				{Start: values.Time(30 * time.Minute), Stop: values.Time(35 * time.Minute)},
				{Start: values.Time(25 * time.Minute), Stop: values.Time(30 * time.Minute)},
				{Start: values.Time(20 * time.Minute), Stop: values.Time(25 * time.Minute)},
				{Start: values.Time(15 * time.Minute), Stop: values.Time(20 * time.Minute)},
				{Start: values.Time(10 * time.Minute), Stop: values.Time(15 * time.Minute)},
			},
		},
		{
			name: "beginning of month",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(0),
			),
			t: mustTime("2020-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2020-10-01T00:00:00Z"), Stop: mustTime("2020-11-01T00:00:00Z")},
				{Start: mustTime("2020-09-01T00:00:00Z"), Stop: mustTime("2020-10-01T00:00:00Z")},
				{Start: mustTime("2020-08-01T00:00:00Z"), Stop: mustTime("2020-09-01T00:00:00Z")},
				{Start: mustTime("2020-07-01T00:00:00Z"), Stop: mustTime("2020-08-01T00:00:00Z")},
				{Start: mustTime("2020-06-01T00:00:00Z"), Stop: mustTime("2020-07-01T00:00:00Z")},
				{Start: mustTime("2020-05-01T00:00:00Z"), Stop: mustTime("2020-06-01T00:00:00Z")},
				{Start: mustTime("2020-04-01T00:00:00Z"), Stop: mustTime("2020-05-01T00:00:00Z")},
				{Start: mustTime("2020-03-01T00:00:00Z"), Stop: mustTime("2020-04-01T00:00:00Z")},
				{Start: mustTime("2020-02-01T00:00:00Z"), Stop: mustTime("2020-03-01T00:00:00Z")},
				{Start: mustTime("2020-01-01T00:00:00Z"), Stop: mustTime("2020-02-01T00:00:00Z")},
				{Start: mustTime("2019-12-01T00:00:00Z"), Stop: mustTime("2020-01-01T00:00:00Z")},
			},
		},
		{
			name: "end of month",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(-24*time.Hour),
			),
			t: mustTime("2020-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2020-09-30T00:00:00Z"), Stop: mustTime("2020-10-30T00:00:00Z")},
				{Start: mustTime("2020-08-31T00:00:00Z"), Stop: mustTime("2020-09-30T00:00:00Z")},
				{Start: mustTime("2020-07-31T00:00:00Z"), Stop: mustTime("2020-08-31T00:00:00Z")},
				{Start: mustTime("2020-06-30T00:00:00Z"), Stop: mustTime("2020-07-30T00:00:00Z")},
				{Start: mustTime("2020-05-31T00:00:00Z"), Stop: mustTime("2020-06-30T00:00:00Z")},
				{Start: mustTime("2020-04-30T00:00:00Z"), Stop: mustTime("2020-05-30T00:00:00Z")},
				{Start: mustTime("2020-03-31T00:00:00Z"), Stop: mustTime("2020-04-30T00:00:00Z")},
				{Start: mustTime("2020-02-29T00:00:00Z"), Stop: mustTime("2020-03-29T00:00:00Z")},
				{Start: mustTime("2020-01-31T00:00:00Z"), Stop: mustTime("2020-02-29T00:00:00Z")},
				{Start: mustTime("2019-12-31T00:00:00Z"), Stop: mustTime("2020-01-31T00:00:00Z")},
			},
		},
		{
			name: "far from bounds",
			w: mustWindow(
				values.ConvertDurationMonths(1),
				values.ConvertDurationMonths(1),
				values.ConvertDurationNsecs(0),
			),
			t: mustTime("2100-10-01T00:00:00Z"),
			want: []testBounds{
				{Start: mustTime("2100-10-01T00:00:00Z"), Stop: mustTime("2100-11-01T00:00:00Z")},
				{Start: mustTime("2100-09-01T00:00:00Z"), Stop: mustTime("2100-10-01T00:00:00Z")},
				{Start: mustTime("2100-08-01T00:00:00Z"), Stop: mustTime("2100-09-01T00:00:00Z")},
				{Start: mustTime("2100-07-01T00:00:00Z"), Stop: mustTime("2100-08-01T00:00:00Z")},
				{Start: mustTime("2100-06-01T00:00:00Z"), Stop: mustTime("2100-07-01T00:00:00Z")},
				{Start: mustTime("2100-05-01T00:00:00Z"), Stop: mustTime("2100-06-01T00:00:00Z")},
				{Start: mustTime("2100-04-01T00:00:00Z"), Stop: mustTime("2100-05-01T00:00:00Z")},
				{Start: mustTime("2100-03-01T00:00:00Z"), Stop: mustTime("2100-04-01T00:00:00Z")},
				{Start: mustTime("2100-02-01T00:00:00Z"), Stop: mustTime("2100-03-01T00:00:00Z")},
				{Start: mustTime("2100-01-01T00:00:00Z"), Stop: mustTime("2100-02-01T00:00:00Z")},
				{Start: mustTime("2099-12-01T00:00:00Z"), Stop: mustTime("2100-01-01T00:00:00Z")},
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b := tc.w.GetLatestBounds(tc.t)
			got := make([]testBounds, 0, len(tc.want))
			for range tc.want {
				got = append(got, testBounds{
					Start: b.Start(),
					Stop:  b.Stop(),
				})
				b = tc.w.PrevBounds(b)
			}
			if !cmp.Equal(tc.want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}

func mustWindow(every, period, offset values.Duration) interval.Window {
	w, err := interval.NewWindow(every, period, offset)
	if err != nil {
		panic(err)
	}
	return w
}

func mustTime(s string) values.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return values.ConvertTime(t)
}

func mustDuration(s string) values.Duration {
	d, err := values.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func transformBounds(b []interval.Bounds) []testBounds {
	bs := make([]testBounds, 0, len(b))
	for i := range b {
		bs = append(bs, testBounds{
			Start: b[i].Start(),
			Stop:  b[i].Stop(),
		})
	}
	return bs
}
