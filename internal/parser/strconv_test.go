package parser_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/parser"
)

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		testName string
		duration string
		want     []ast.Duration
		err      error
	}{
		{
			testName: "All durations",
			duration: "1y3mo2w1d4h1m30s1ms2µs70ns",
			want: []ast.Duration{
				{
					Magnitude: 1,
					Unit:      "y",
				},
				{
					Magnitude: 3,
					Unit:      "mo",
				},
				{
					Magnitude: 2,
					Unit:      "w",
				},
				{
					Magnitude: 1,
					Unit:      "d",
				},
				{
					Magnitude: 4,
					Unit:      "h",
				},
				{
					Magnitude: 1,
					Unit:      "m",
				},
				{
					Magnitude: 30,
					Unit:      "s",
				},
				{
					Magnitude: 1,
					Unit:      "ms",
				},
				{
					Magnitude: 2,
					Unit:      "us",
				},
				{
					Magnitude: 70,
					Unit:      "ns",
				},
			},
		},
		{
			testName: "Leading zero durations",
			duration: "01y03mo02w01d04h01m03s01ms02µs07ns",
			want: []ast.Duration{
				{
					Magnitude: 1,
					Unit:      "y",
				},
				{
					Magnitude: 3,
					Unit:      "mo",
				},
				{
					Magnitude: 2,
					Unit:      "w",
				},
				{
					Magnitude: 1,
					Unit:      "d",
				},
				{
					Magnitude: 4,
					Unit:      "h",
				},
				{
					Magnitude: 1,
					Unit:      "m",
				},
				{
					Magnitude: 3,
					Unit:      "s",
				},
				{
					Magnitude: 1,
					Unit:      "ms",
				},
				{
					Magnitude: 2,
					Unit:      "us",
				},
				{
					Magnitude: 7,
					Unit:      "ns",
				},
			},
		},
		{
			testName: "Many leading zeros duration",
			duration: "000000000001234d",
			want: []ast.Duration{
				{
					Magnitude: 1234,
					Unit:      "d",
				},
			},
		},
		{
			testName: "Missing duration magnitude",
			duration: "d",
			err:      errors.New(codes.Invalid, "invalid duration d"),
		},
		{
			testName: "Repeated duration units",
			duration: "1s2d3s4d",
			want: []ast.Duration{
				{
					Magnitude: 1,
					Unit:      "s",
				},
				{
					Magnitude: 2,
					Unit:      "d",
				},
				{
					Magnitude: 3,
					Unit:      "s",
				},
				{
					Magnitude: 4,
					Unit:      "d",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			result, err := parser.ParseDuration(tc.duration)

			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.err != nil && tc.err == nil {
				t.Errorf("Expected error but got nil: %v", tc.err)
			} else if tc.err != nil && !cmp.Equal(err, tc.err) {
				t.Errorf("Expected duration error: %v", cmp.Diff(err, tc.err))
			} else if !cmp.Equal(result, tc.want) {
				t.Errorf("Expected duration values to be eq: %v", cmp.Diff(result, tc.want))
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	testCases := []struct {
		testName string
		time     string
		want     time.Time
		err      error
	}{
		{
			testName: "RFC3339Nano",
			time:     "2022-09-14T04:37:17.123456789Z",
			want:     time.Date(2022, 9, 14, 4, 37, 17, 123456789, time.UTC),
		},
		{
			testName: "millis",
			time:     "2022-09-14T04:37:17.123Z",
			want:     time.Date(2022, 9, 14, 4, 37, 17, 123000000, time.UTC),
		},
		{
			testName: "date time offset",
			time:     "2022-09-14T04:37:17.123456789-07:00",
			want:     time.Date(2022, 9, 14, 4, 37, 17, 123456789, time.FixedZone("", -7*60*60)),
		},
		{
			testName: "date only",
			time:     "2022-09-14",
			want:     time.Date(2022, 9, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			testName: "date only error",
			time:     "2022-00-14",
			err:      errors.New(codes.Invalid, "cannot parse date"),
		},
		{
			testName: "date time no offset",
			time:     "2022-09-14T04:37:17.123456789",
			err:      errors.New(codes.Invalid, "cannot parse date time"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			result, err := parser.ParseTime(tc.time)

			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.err != nil && tc.err == nil {
				t.Errorf("Expected error but got nil: %v", tc.err)
			} else if tc.err != nil && !cmp.Equal(err, tc.err) {
				t.Errorf("Expected time error: %v", cmp.Diff(err, tc.err))
			} else if !cmp.Equal(result, tc.want) {
				t.Errorf("Expected time values to be eq: %v", cmp.Diff(result, tc.want))
			}
		})
	}
}

func TestParseString(t *testing.T) {
	testCases := []struct {
		testName string
		str      string
		want     string
		err      error
	}{
		{
			testName: "normal",
			str:      `"hello world"`,
			want:     "hello world",
		},
		{
			testName: "escape sequences",
			str: `"newline\n
carriage return\r
horizontal tab\t
double quote \"
backslash \\
dollar curly braket \${
"`,

			want: "newline\n\ncarriage return\r\nhorizontal tab\t\ndouble quote \"\nbackslash \\\ndollar curly braket ${\n",
		},
		{
			testName: "hex escape sequences",
			str:      `"\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e"`,
			want:     "日本語",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			result, err := parser.ParseString(tc.str)

			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.err != nil && tc.err == nil {
				t.Errorf("Expected error but got nil: %v", tc.err)
			} else if tc.err != nil && !cmp.Equal(err, tc.err) {
				t.Errorf("Expected string error: %v", cmp.Diff(err, tc.err))
			} else if !cmp.Equal(result, tc.want) {
				t.Errorf("Expected string values to be eq: %v", cmp.Diff(result, tc.want))
			}
		})
	}
}

func TestParseRegex(t *testing.T) {
	testCases := []struct {
		testName string
		str      string
		want     string
		err      error
	}{
		{
			testName: "normal",
			str:      `/hello world/`,
			want:     "hello world",
		},
		{
			testName: "escape sequences",
			str:      `/forward slash \/ character classes: \w\s\d/`,
			want:     `forward slash / character classes: \w\s\d`,
		},
		{
			testName: "hex escape sequences",
			str:      `/\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e/`,
			want:     "日本語",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			regex, err := parser.ParseRegexp(tc.str)
			var result string
			if regex != nil {
				result = regex.String()
			}

			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tc.err != nil && tc.err == nil {
				t.Errorf("Expected error but got nil: %v", tc.err)
			} else if tc.err != nil && !cmp.Equal(err, tc.err) {
				t.Errorf("Expected regexp error: %v", cmp.Diff(err, tc.err))
			} else if !cmp.Equal(result, tc.want) {
				t.Errorf("Expected regexp values to be eq: %v", cmp.Diff(result, tc.want))
			}
		})
	}
}
