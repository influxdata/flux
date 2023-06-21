package parser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/InfluxCommunity/flux/ast"
	"github.com/InfluxCommunity/flux/parser"
)

func TestParseSignedDuration(t *testing.T) {
	t.Run("negative simple duration", func(t *testing.T) {
		durs, err := parser.ParseSignedDuration("-1m")
		if err != nil {
			t.Fatal(err)
		}
		if durs.Values[0].Magnitude != -1 {
			t.Fatalf("expected magnitude of -1 but got %d", durs.Values[0].Magnitude)
		}
	})
	t.Run("positive simple duration", func(t *testing.T) {
		durs, err := parser.ParseSignedDuration("1m")
		if err != nil {
			t.Fatal(err)
		}
		if durs.Values[0].Magnitude != 1 {
			t.Fatalf("expected magnitude of 1 but got %d", durs.Values[0].Magnitude)
		}
	})
	t.Run("negative complex duration", func(t *testing.T) {
		durs, err := parser.ParseSignedDuration("-1d2h1m3s")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff([]ast.Duration{
			{
				Magnitude: -1,
				Unit:      "d",
			},
			{
				Magnitude: -2,
				Unit:      "h",
			},
			{
				Magnitude: -1,
				Unit:      "m",
			},
			{
				Magnitude: -3,
				Unit:      "s",
			},
		}, durs.Values); diff != "" {
			t.Fatal(diff)
		}

	})

}

func TestFormatDuration(t *testing.T) {
	t.Run("format negative duration", func(t *testing.T) {
		node := &ast.DurationLiteral{
			Values: []ast.Duration{
				{
					Magnitude: -1,
					Unit:      "d",
				},
				{
					Magnitude: -2,
					Unit:      "h",
				},
				{
					Magnitude: -1,
					Unit:      "m",
				},
				{
					Magnitude: -3,
					Unit:      "s",
				},
			},
		}
		durs := parser.FormatDuration(node)
		if diff := cmp.Diff("-1d-2h-1m-3s", durs); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("format duration", func(t *testing.T) {
		node := &ast.DurationLiteral{
			Values: []ast.Duration{
				{
					Magnitude: 1,
					Unit:      "d",
				},
				{
					Magnitude: 2,
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
			},
		}
		durs := parser.FormatDuration(node)
		if diff := cmp.Diff("1d2h1m3s", durs); diff != "" {
			t.Fatal(diff)
		}
	})
}
