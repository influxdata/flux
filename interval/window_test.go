package interval_test

import (
	"testing"
	"time"

	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/interval"
	"github.com/InfluxCommunity/flux/values"
	"github.com/google/go-cmp/cmp"
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

func TestWindow_GetLatestBounds(t *testing.T) {
	var testcases = []struct {
		name string
		w    interval.Window
		t    values.Time
		want execute.Bounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0)),
			t: values.Time(6 * time.Minute),
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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
			want: execute.Bounds{
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

func TestWindow_GetLatestBounds_InLocation(t *testing.T) {
	const (
		American_Samoa  = "Pacific/Apia"
		America_Phoenix = "America/Phoenix"
		America_Denver  = "America/Denver"
		Australia_East  = "Australia/Sydney"
		Europe_Moscow   = "Europe/Moscow"
		US_Pacific      = "America/Los_Angeles"
	)

	type window struct {
		every  string
		period string
		offset string
	}

	var testcases = []struct {
		name string
		loc  string
		w    window
		t    string
		want [2]string
	}{
		{
			name: "America_Phoenix",
			loc:  America_Phoenix,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-02-24T12:00:00-07:00",
			want: [2]string{
				"2017-02-24T00:00:00-07:00",
				"2017-02-25T00:00:00-07:00",
			},
		},
		{
			name: "America_Phoenix DST", // Phoenix doesn't observe DST
			loc:  America_Phoenix,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-09-03T12:00:00-07:00",
			want: [2]string{
				"2017-09-03T00:00:00-07:00",
				"2017-09-04T00:00:00-07:00",
			},
		},
		{
			name: "America_Denver",
			loc:  America_Denver,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-02-24T12:00:00-07:00",
			want: [2]string{
				"2017-02-24T00:00:00-07:00",
				"2017-02-25T00:00:00-07:00",
			},
		},
		{
			name: "America_Denver DST", // Denver observes DST
			loc:  America_Denver,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-09-03T12:00:00-06:00",
			want: [2]string{
				"2017-09-03T00:00:00-06:00",
				"2017-09-04T00:00:00-06:00",
			},
		},
		{
			name: "Europe_Moscow", // Moscow doesn't observe DST between 2015 - 2019
			loc:  Europe_Moscow,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-09-03T12:00:00+03:00",
			want: [2]string{
				"2017-09-03T00:00:00+03:00",
				"2017-09-04T00:00:00+03:00",
			},
		},
		{
			name: "Europe_Moscow DST", // Moscow observe DST in 2009
			loc:  Europe_Moscow,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2009-03-30T12:00:00+04:00",
			want: [2]string{
				"2009-03-30T00:00:00+04:00",
				"2009-03-31T00:00:00+04:00",
			},
		},
		{
			name: "US_Pacific",
			loc:  US_Pacific,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-02-24T12:00:00-08:00",
			want: [2]string{
				"2017-02-24T00:00:00-08:00",
				"2017-02-25T00:00:00-08:00",
			},
		},
		{
			name: "US_Pacific DST",
			loc:  US_Pacific,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-09-03T12:00:00-07:00",
			want: [2]string{
				"2017-09-03T00:00:00-07:00",
				"2017-09-04T00:00:00-07:00",
			},
		},
		{
			name: "US_Pacific DST start",
			loc:  US_Pacific,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-03-12T03:00:00-07:00",
			want: [2]string{
				"2017-03-12T00:00:00-08:00",
				"2017-03-13T00:00:00-07:00",
			},
		},
		{
			name: "US_Pacific DST start every 1h offset 30m",
			loc:  US_Pacific,
			w: window{
				every:  "1h",
				period: "1h",
				offset: "30m",
			},
			t: "2017-03-12T01:45:00-08:00",
			want: [2]string{
				"2017-03-12T01:30:00-08:00",
				"2017-03-12T03:00:00-07:00",
			},
		},
		{
			name: "US_Pacific DST end",
			loc:  US_Pacific,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-11-05T01:30:00-08:00",
			want: [2]string{
				"2017-11-05T00:00:00-07:00",
				"2017-11-06T00:00:00-08:00",
			},
		},
		{
			name: "US_Pacific DST end every 1h offset 30m",
			loc:  US_Pacific,
			w: window{
				every:  "1h",
				period: "1h",
				offset: "30m",
			},
			t: "2017-11-05T01:45:00-08:00",
			want: [2]string{
				"2017-11-05T01:30:00-07:00",
				"2017-11-05T02:30:00-08:00",
			},
		},
		{
			name: "Australia_East",
			loc:  Australia_East,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-09-03T12:00:00+10:00",
			want: [2]string{
				"2017-09-03T00:00:00+10:00",
				"2017-09-04T00:00:00+10:00",
			},
		},
		{
			name: "Australia_East DST",
			loc:  Australia_East,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-02-24T12:00:00+11:00",
			want: [2]string{
				"2017-02-24T00:00:00+11:00",
				"2017-02-25T00:00:00+11:00",
			},
		},
		{
			name: "Australia_East DST start",
			loc:  Australia_East,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-10-01T03:00:00+11:00",
			want: [2]string{
				"2017-10-01T00:00:00+10:00",
				"2017-10-02T00:00:00+11:00",
			},
		},
		{
			name: "Australia_East DST start every 1h offset 30m",
			loc:  Australia_East,
			w: window{
				every:  "1h",
				period: "1h",
				offset: "30m",
			},
			t: "2017-10-01T01:45:00+10:00",
			want: [2]string{
				"2017-10-01T01:30:00+10:00",
				"2017-10-01T03:00:00+11:00",
			},
		},
		{
			name: "Australia_East DST end",
			loc:  Australia_East,
			w: window{
				every:  "1d",
				period: "1d",
			},
			t: "2017-04-02T02:30:00+11:00",
			want: [2]string{
				"2017-04-02T00:00:00+11:00",
				"2017-04-03T00:00:00+10:00",
			},
		},
		{
			name: "Australia_East DST end every 1h offset 30m",
			loc:  Australia_East,
			w: window{
				every:  "1h",
				period: "1h",
				offset: "30m",
			},
			t: "2017-04-02T02:45:00+10:00",
			want: [2]string{
				"2017-04-02T02:30:00+11:00",
				"2017-04-02T03:30:00+10:00",
			},
		},
		{
			name: "American_Samoa day skip start",
			loc:  American_Samoa,
			w: window{
				every:  "1d",
				period: "1d",
				offset: "12h",
			},
			t: "2011-12-29T16:00:00-10:00",
			want: [2]string{
				"2011-12-29T12:00:00-10:00",
				"2011-12-31T00:00:00+14:00",
			},
		},
		{
			name: "American_Samoa day skip end",
			loc:  American_Samoa,
			w: window{
				every:  "1d",
				period: "1d",
				offset: "12h",
			},
			t: "2011-12-31T04:00:00+14:00",
			want: [2]string{
				"2011-12-31T00:00:00+14:00",
				"2011-12-31T12:00:00+14:00",
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			loc, err := interval.LoadLocation(tc.loc)
			if err != nil {
				t.Fatal(err)
			}

			every, err := values.ParseDuration(tc.w.every)
			if err != nil {
				t.Fatal(err)
			}

			period, err := values.ParseDuration(tc.w.period)
			if err != nil {
				t.Fatal(err)
			}

			var offset values.Duration
			if tc.w.offset != "" {
				offset, err = values.ParseDuration(tc.w.offset)
				if err != nil {
					t.Fatal(err)
				}
			}

			want := execute.Bounds{
				Start: values.Time(mustTimeInLocation(t, tc.want[0], tc.loc)),
				Stop:  values.Time(mustTimeInLocation(t, tc.want[1], tc.loc)),
			}

			window, err := interval.NewWindowInLocation(every, period, offset, loc)
			if err != nil {
				t.Fatal(err)
			}

			ts := values.Time(mustTimeInLocation(t, tc.t, tc.loc))
			got := window.GetLatestBounds(ts)
			if got.Start() != want.Start {
				t.Errorf("unexpected start boundary: got %s want %s", got.Start(), want.Start)
			}
			if got.Stop() != want.Stop {
				t.Errorf("unexpected stop boundary:  got %s want %s", got.Stop(), want.Stop)
			}
		})
	}
}

func TestWindow_GetOverlappingBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    interval.Window
		b    execute.Bounds
		want []execute.Bounds
	}{
		{
			name: "empty",
			w: mustWindow(
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(0),
			),
			b: execute.Bounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(5 * time.Minute),
			},
			want: []execute.Bounds{},
		},
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(0),
			),
			b: execute.Bounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(8 * time.Minute),
			},
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: values.Time(5 * time.Minute),
				Stop:  values.Time(7 * time.Minute),
			},
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: values.Time(1*time.Minute + 30*time.Second),
				Stop:  values.Time(1*time.Minute + 45*time.Second),
			},
			want: []execute.Bounds{},
		},
		{
			name: "underlapping",
			w: mustWindow(
				values.ConvertDurationNsecs(2*time.Minute),
				values.ConvertDurationNsecs(time.Minute),
				values.ConvertDurationNsecs(30*time.Second),
			),
			b: execute.Bounds{
				Start: values.Time(1*time.Minute + 45*time.Second),
				Stop:  values.Time(4*time.Minute + 35*time.Second),
			},
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: values.Time(10 * time.Minute),
				Stop:  values.Time(12 * time.Minute),
			},
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: mustTime("2019-10-01T00:00:00Z"),
				Stop:  mustTime("2019-10-08T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1d"),
				mustDuration("1d"),
				values.ConvertDurationNsecs(0),
			),
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: mustTime("2019-01-01T00:00:00Z"),
				Stop:  mustTime("2020-01-01T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1mo"),
				mustDuration("1mo"),
				values.ConvertDurationNsecs(0),
			),
			want: []execute.Bounds{
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
			b: execute.Bounds{
				Start: mustTime("2019-01-01T00:00:00Z"),
				Stop:  mustTime("2020-01-01T00:00:00Z"),
			},
			w: mustWindow(
				mustDuration("1mo"),
				mustDuration("3mo"),
				values.ConvertDurationNsecs(0),
			),
			want: []execute.Bounds{
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

func TestWindow_GetOverlappingBounds_InLocation(t *testing.T) {
	const (
		US_Pacific     = "America/Los_Angeles"
		Australia_East = "Australia/Sydney"
		American_Samoa = "Pacific/Apia"
	)

	type window struct {
		every  string
		period string
		offset string
	}

	var testcases = []struct {
		name string
		loc  string
		w    window
		t    string
		want [][2]string
	}{
		{
			name: "US_Pacific",
			loc:  US_Pacific,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-02-24T12:00:00-08:00",
			want: [][2]string{
				{"2017-02-23T16:00:00-08:00", "2017-02-24T16:00:00-08:00"},
				{"2017-02-24T00:00:00-08:00", "2017-02-25T00:00:00-08:00"},
				{"2017-02-24T08:00:00-08:00", "2017-02-25T08:00:00-08:00"},
			},
		},
		{
			name: "US_Pacific DST",
			loc:  US_Pacific,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-09-03T12:00:00-07:00",
			want: [][2]string{
				{"2017-09-02T16:00:00-07:00", "2017-09-03T16:00:00-07:00"},
				{"2017-09-03T00:00:00-07:00", "2017-09-04T00:00:00-07:00"},
				{"2017-09-03T08:00:00-07:00", "2017-09-04T08:00:00-07:00"},
			},
		},
		{
			name: "US_Pacific DST start",
			loc:  US_Pacific,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-03-12T03:00:00-07:00",
			want: [][2]string{
				{"2017-03-11T08:00:00-08:00", "2017-03-12T08:00:00-07:00"},
				{"2017-03-11T16:00:00-08:00", "2017-03-12T16:00:00-07:00"},
				{"2017-03-12T00:00:00-08:00", "2017-03-13T00:00:00-07:00"},
			},
		},
		{
			name: "US_Pacific DST start every 1h offset 30m",
			loc:  US_Pacific,
			w: window{
				every:  "1h",
				period: "4h",
				offset: "30m",
			},
			t: "2017-03-12T01:45:00-08:00",
			want: [][2]string{
				{"2017-03-11T22:30:00-08:00", "2017-03-12T03:00:00-07:00"},
				{"2017-03-11T23:30:00-08:00", "2017-03-12T03:30:00-07:00"},
				{"2017-03-12T00:30:00-08:00", "2017-03-12T04:30:00-07:00"},
				{"2017-03-12T01:30:00-08:00", "2017-03-12T05:30:00-07:00"},
			},
		},
		{
			name: "US_Pacific DST end",
			loc:  US_Pacific,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-11-05T01:30:00-08:00",
			want: [][2]string{
				{"2017-11-04T08:00:00-07:00", "2017-11-05T08:00:00-08:00"},
				{"2017-11-04T16:00:00-07:00", "2017-11-05T16:00:00-08:00"},
				{"2017-11-05T00:00:00-07:00", "2017-11-06T00:00:00-08:00"},
			},
		},
		{
			name: "US_Pacific DST end every 1h offset 30m",
			loc:  US_Pacific,
			w: window{
				every:  "1h",
				period: "4h",
				offset: "30m",
			},
			t: "2017-11-05T01:45:00-08:00",
			want: [][2]string{
				{"2017-11-04T22:30:00-07:00", "2017-11-05T02:30:00-08:00"},
				{"2017-11-04T23:30:00-07:00", "2017-11-05T03:30:00-08:00"},
				{"2017-11-05T00:30:00-07:00", "2017-11-05T04:30:00-08:00"},
				{"2017-11-05T01:30:00-07:00", "2017-11-05T05:30:00-08:00"},
			},
		},
		{
			name: "Australia_East",
			loc:  Australia_East,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-09-17T12:00:00+10:00",
			want: [][2]string{
				{"2017-09-16T16:00:00+10:00", "2017-09-17T16:00:00+10:00"},
				{"2017-09-17T00:00:00+10:00", "2017-09-18T00:00:00+10:00"},
				{"2017-09-17T08:00:00+10:00", "2017-09-18T08:00:00+10:00"},
			},
		},
		{
			name: "Australia_East DST",
			loc:  Australia_East,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-02-24T12:00:00+11:00",
			want: [][2]string{
				{"2017-02-23T16:00:00+11:00", "2017-02-24T16:00:00+11:00"},
				{"2017-02-24T00:00:00+11:00", "2017-02-25T00:00:00+11:00"},
				{"2017-02-24T08:00:00+11:00", "2017-02-25T08:00:00+11:00"},
			},
		},
		{
			name: "Australia_East DST start",
			loc:  Australia_East,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-10-01T03:00:00+11:00",
			want: [][2]string{
				{"2017-09-30T08:00:00+10:00", "2017-10-01T08:00:00+11:00"},
				{"2017-09-30T16:00:00+10:00", "2017-10-01T16:00:00+11:00"},
				{"2017-10-01T00:00:00+10:00", "2017-10-02T00:00:00+11:00"},
			},
		},
		{
			name: "Australia_East DST start every 1h offset 30m",
			loc:  Australia_East,
			w: window{
				every:  "1h",
				period: "4h",
				offset: "30m",
			},
			t: "2017-10-01T01:45:00+10:00",
			want: [][2]string{
				{"2017-09-30T22:30:00+10:00", "2017-10-01T03:00:00+11:00"},
				{"2017-09-30T23:30:00+10:00", "2017-10-01T03:30:00+11:00"},
				{"2017-10-01T00:30:00+10:00", "2017-10-01T04:30:00+11:00"},
				{"2017-10-01T01:30:00+10:00", "2017-10-01T05:30:00+11:00"},
			},
		},
		{
			name: "Australia_East DST end",
			loc:  Australia_East,
			w: window{
				every:  "8h",
				period: "1d",
			},
			t: "2017-04-02T02:30:00+11:00",
			want: [][2]string{
				{"2017-04-01T08:00:00+11:00", "2017-04-02T08:00:00+10:00"},
				{"2017-04-01T16:00:00+11:00", "2017-04-02T16:00:00+10:00"},
				{"2017-04-02T00:00:00+11:00", "2017-04-03T00:00:00+10:00"},
			},
		},
		{
			name: "Australia_East DST end every 1h offset 30m",
			loc:  Australia_East,
			w: window{
				every:  "1h",
				period: "4h",
				offset: "30m",
			},
			t: "2017-04-02T02:45:00+10:00",
			want: [][2]string{
				{"2017-04-01T23:30:00+11:00", "2017-04-02T03:30:00+10:00"},
				{"2017-04-02T00:30:00+11:00", "2017-04-02T04:30:00+10:00"},
				{"2017-04-02T01:30:00+11:00", "2017-04-02T05:30:00+10:00"},
				{"2017-04-02T02:30:00+11:00", "2017-04-02T06:30:00+10:00"},
			},
		},
		{
			name: "American_Samoa day skip start",
			loc:  American_Samoa,
			w: window{
				every:  "1d",
				period: "1w",
				offset: "2h",
			},
			t: "2011-12-29T16:00:00-10:00",
			want: [][2]string{
				{"2011-12-23T02:00:00-10:00", "2011-12-31T00:00:00+14:00"},
				{"2011-12-24T02:00:00-10:00", "2011-12-31T02:00:00+14:00"},
				{"2011-12-25T02:00:00-10:00", "2012-01-01T02:00:00+14:00"},
				{"2011-12-26T02:00:00-10:00", "2012-01-02T02:00:00+14:00"},
				{"2011-12-27T02:00:00-10:00", "2012-01-03T02:00:00+14:00"},
				{"2011-12-28T02:00:00-10:00", "2012-01-04T02:00:00+14:00"},
				{"2011-12-29T02:00:00-10:00", "2012-01-05T02:00:00+14:00"},
			},
		},
		{
			name: "American_Samoa day skip end",
			loc:  American_Samoa,
			w: window{
				every:  "1d",
				period: "1w",
				offset: "2h",
			},
			t: "2011-12-31T04:00:00+14:00",
			want: [][2]string{
				{"2011-12-25T02:00:00-10:00", "2012-01-01T02:00:00+14:00"},
				{"2011-12-26T02:00:00-10:00", "2012-01-02T02:00:00+14:00"},
				{"2011-12-27T02:00:00-10:00", "2012-01-03T02:00:00+14:00"},
				{"2011-12-28T02:00:00-10:00", "2012-01-04T02:00:00+14:00"},
				{"2011-12-29T02:00:00-10:00", "2012-01-05T02:00:00+14:00"},
				{"2011-12-31T00:00:00+14:00", "2012-01-06T02:00:00+14:00"},
				{"2011-12-31T02:00:00+14:00", "2012-01-07T02:00:00+14:00"},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			loc, err := interval.LoadLocation(tc.loc)
			if err != nil {
				t.Fatal(err)
			}

			every, err := values.ParseDuration(tc.w.every)
			if err != nil {
				t.Fatal(err)
			}

			period, err := values.ParseDuration(tc.w.period)
			if err != nil {
				t.Fatal(err)
			}

			var offset values.Duration
			if tc.w.offset != "" {
				offset, err = values.ParseDuration(tc.w.offset)
				if err != nil {
					t.Fatal(err)
				}
			}

			want := make([]execute.Bounds, len(tc.want))
			for i, b := range tc.want {
				want[i] = execute.Bounds{
					Start: values.Time(mustTimeInLocation(t, b[0], tc.loc)),
					Stop:  values.Time(mustTimeInLocation(t, b[1], tc.loc)),
				}
			}

			window, err := interval.NewWindowInLocation(every, period, offset, loc)
			if err != nil {
				t.Fatal(err)
			}

			ts := values.Time(mustTimeInLocation(t, tc.t, tc.loc))
			got := transformBounds(window.GetOverlappingBounds(ts, ts+1))
			for i, j := 0, len(got)-1; i < j; i, j = i+1, j-1 {
				got[i], got[j] = got[j], got[i]
			}
			if !cmp.Equal(want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(want, got))
			}
		})
	}
}

func TestWindow_NextBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    interval.Window
		t    values.Time
		want []execute.Bounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(10 * time.Minute),
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			got := make([]execute.Bounds, 0, len(tc.want))
			for range tc.want {
				got = append(got, execute.Bounds{
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
		want []execute.Bounds
	}{
		{
			name: "simple",
			w: mustWindow(
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(5*time.Minute),
				values.ConvertDurationNsecs(0),
			),
			t: values.Time(36 * time.Minute),
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			want: []execute.Bounds{
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
			got := make([]execute.Bounds, 0, len(tc.want))
			for range tc.want {
				got = append(got, execute.Bounds{
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

func mustTimeInLocation(t *testing.T, s, loc string) int64 {
	t.Helper()

	// Load location from the time library and parse
	// the location. Then verify that the location output
	// is the same after we parse it. This is to prevent
	// developer errors where we input the offset incorrectly.
	// The Go library won't detect that, so we do it by
	// round-tripping the parsing.
	timeLoc, err := time.LoadLocation(loc)
	if err != nil {
		t.Fatal(err)
	}

	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	ts = ts.In(timeLoc)

	if want, got := s, ts.Format(time.RFC3339); want != got {
		t.Fatalf("unexpected output from time parse -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
	return ts.UnixNano()
}

func mustDuration(s string) values.Duration {
	d, err := values.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func transformBounds(b []interval.Bounds) []execute.Bounds {
	bs := make([]execute.Bounds, 0, len(b))
	for i := range b {
		bs = append(bs, execute.Bounds{
			Start: b[i].Start(),
			Stop:  b[i].Stop(),
		})
	}
	return bs
}
