package parser_test

import (
	"testing"

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
