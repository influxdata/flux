package ast_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mvn-trinhnguyen2-dn/flux/ast"
)

func TestDurationLiteralString(t *testing.T) {
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
		durs := node.String()
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
		durs := node.String()
		if diff := cmp.Diff("1d2h1m3s", durs); diff != "" {
			t.Fatal(diff)
		}
	})
}
