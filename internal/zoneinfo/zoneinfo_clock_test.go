package zoneinfo_test

import (
	"testing"
	"time"

	"github.com/InfluxCommunity/flux/internal/zoneinfo"
)

func TestLocation_ToLocalClock(t *testing.T) {
	const (
		American_Samoa  = "Pacific/Apia"
		America_Phoenix = "America/Phoenix"
		America_Denver  = "America/Denver"
		Asia_Kolkata    = "Asia/Kolkata"
		Australia_East  = "Australia/Sydney"
		Europe_Moscow   = "Europe/Moscow"
		US_Pacific      = "America/Los_Angeles"
	)
	for _, tt := range []struct {
		name    string
		locname string
		s       string
		want    string
	}{
		{
			name:    "America_Phoenix",
			locname: America_Phoenix,
			s:       "2017-02-24T12:00:00Z",
			want:    "2017-02-24T12:00:00-07:00",
		},
		{
			name:    "America_Phoenix DST", // Phoenix doesn't observe DST
			locname: America_Phoenix,
			s:       "2017-09-03T12:00:00Z",
			want:    "2017-09-03T12:00:00-07:00",
		},
		{
			name:    "America_Denver",
			locname: America_Denver,
			s:       "2017-02-24T12:00:00Z",
			want:    "2017-02-24T12:00:00-07:00",
		},
		{
			name:    "America_Denver DST", // Denver observes DST
			locname: America_Denver,
			s:       "2017-09-03T12:00:00Z",
			want:    "2017-09-03T12:00:00-06:00",
		},
		{
			name:    "Europe_Moscow", // Moscow doesn't observe DST between 2015 - 2019
			locname: Europe_Moscow,
			s:       "2017-03-30T03:00:00Z",
			want:    "2017-03-30T03:00:00+03:00",
		},
		{
			name:    "Europe_Moscow DST", // Moscow observe DST in 2009
			locname: Europe_Moscow,
			s:       "2009-03-30T03:00:00Z",
			want:    "2009-03-30T03:00:00+04:00",
		},
		{
			name:    "Asia_Kolkata",
			locname: Asia_Kolkata,
			s:       "2017-09-03T12:00:00Z",
			want:    "2017-09-03T12:00:00+05:30",
		},
		{
			name:    "US_Pacific",
			locname: US_Pacific,
			s:       "2017-02-24T12:00:00Z",
			want:    "2017-02-24T12:00:00-08:00",
		},
		{
			name:    "US_Pacific DST",
			locname: US_Pacific,
			s:       "2017-09-03T12:00:00Z",
			want:    "2017-09-03T12:00:00-07:00",
		},
		{
			name:    "US_Pacific DST start",
			locname: US_Pacific,
			s:       "2017-03-12T02:30:00Z",
			want:    "2017-03-12T03:00:00-07:00",
		},
		{
			name:    "US_Pacific DST end",
			locname: US_Pacific,
			s:       "2017-11-05T01:30:00Z",
			want:    "2017-11-05T01:30:00-07:00",
		},
		{
			name:    "Australia_East",
			locname: Australia_East,
			s:       "2017-09-17T12:00:00Z",
			want:    "2017-09-17T12:00:00+10:00",
		},
		{
			name:    "Australia_East DST",
			locname: Australia_East,
			s:       "2017-02-24T12:00:00Z",
			want:    "2017-02-24T12:00:00+11:00",
		},
		{
			name:    "Australia_East DST start",
			locname: Australia_East,
			s:       "2017-10-01T02:30:00Z",
			want:    "2017-10-01T03:00:00+11:00",
		},
		{
			name:    "Australia_East DST end",
			locname: Australia_East,
			s:       "2017-04-02T02:30:00Z",
			want:    "2017-04-02T02:30:00+11:00",
		},
		{
			name:    "American_Samoa day skip morning",
			locname: American_Samoa,
			s:       "2011-12-30T06:00:00Z",
			want:    "2011-12-31T00:00:00+14:00",
		},
		{
			name:    "American_Samoa day skip evening",
			locname: American_Samoa,
			s:       "2011-12-30T18:00:00Z",
			want:    "2011-12-31T00:00:00+14:00",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := zoneinfo.LoadLocation(tt.locname)
			if err != nil {
				t.Fatal(err)
			}

			utc := mustParseTime(t, tt.s)
			local := loc.ToLocalClock(utc)

			if want, got := mustParseTimeInLocation(t, tt.want, tt.locname), local; want != got {
				timeLoc, err := time.LoadLocation(tt.locname)
				if err != nil {
					t.Fatal(err)
				}

				t.Errorf("unexpected local time -want/+got:\n\t- %d (%s)\n\t+ %d (%s)",
					want, time.Unix(0, want).In(timeLoc).Format(time.RFC3339),
					got, time.Unix(0, got).In(timeLoc).Format(time.RFC3339))
			}
		})
	}
}

func TestLocation_FromLocalClock(t *testing.T) {
	const (
		US_Pacific     = "America/Los_Angeles"
		Australia_East = "Australia/Sydney"
	)
	for _, tt := range []struct {
		name    string
		locname string
		s       string
		want    string
	}{
		{
			name:    "US_Pacific",
			locname: US_Pacific,
			s:       "2017-02-24T12:00:00-08:00",
			want:    "2017-02-24T12:00:00Z",
		},
		{
			name:    "US_Pacific DST",
			locname: US_Pacific,
			s:       "2017-09-03T12:00:00-07:00",
			want:    "2017-09-03T12:00:00Z",
		},
		// US_Pacific DST start omitted because the time does not exist.
		{
			name:    "US_Pacific DST end",
			locname: US_Pacific,
			s:       "2017-11-05T01:30:00-07:00",
			want:    "2017-11-05T01:30:00Z",
		},
		{
			name:    "Australia_East",
			locname: Australia_East,
			s:       "2017-09-17T12:00:00+10:00",
			want:    "2017-09-17T12:00:00Z",
		},
		{
			name:    "Australia_East DST",
			locname: Australia_East,
			s:       "2017-02-24T12:00:00+11:00",
			want:    "2017-02-24T12:00:00Z",
		},
		// Australia_East DST start omitted because the time does not exist.
		{
			name:    "Australia_East DST end",
			locname: Australia_East,
			s:       "2017-04-02T02:30:00+11:00",
			want:    "2017-04-02T02:30:00Z",
		},
		// American_Samoa day skip morning omitted because the time does not exist.
		// American_Samoa day skip evening omitted because the time does not exist.
	} {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := zoneinfo.LoadLocation(tt.locname)
			if err != nil {
				t.Fatal(err)
			}

			// Parse the time and retrieve the unix timestanp.
			local := mustParseTimeInLocation(t, tt.s, tt.locname)
			utc := loc.FromLocalClock(local)

			if want, got := mustParseTime(t, tt.want), utc; want != got {
				t.Errorf("unexpected utc time -want/+got:\n\t- %d (%s)\n\t+ %d (%s)",
					want, time.Unix(0, want).UTC().Format(time.RFC3339),
					got, time.Unix(0, got).UTC().Format(time.RFC3339))
			}
		})
	}
}

func mustParseTime(t *testing.T, s string) int64 {
	t.Helper()

	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts.UnixNano()
}

func mustParseTimeInLocation(t *testing.T, s, loc string) int64 {
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
