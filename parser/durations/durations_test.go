package durations

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
)

var tests = []struct {
	name      string
	raw       string
	want      *ast.Program
	wantErr   bool
	errString string
}{
	{
		name: "duration literal, all units",
		raw:  `1y3mo2w1d4h1m30s5ms2μs70ns`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 1, Unit: "y"},
						{Magnitude: 3, Unit: "mo"},
						{Magnitude: 2, Unit: "w"},
						{Magnitude: 1, Unit: "d"},
						{Magnitude: 4, Unit: "h"},
						{Magnitude: 1, Unit: "m"},
						{Magnitude: 30, Unit: "s"},
						{Magnitude: 5, Unit: "ms"},
						{Magnitude: 2, Unit: "us"},
						{Magnitude: 70, Unit: "ns"},
					},
				},
			}},
		},
	},
	{
		name: "months",
		raw:  `6mo`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 6, Unit: "mo"},
					},
				},
			}},
		},
	},
	{
		name: "milliseconds",
		raw:  `500ms`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 500, Unit: "ms"},
					},
				},
			}},
		},
	},
	{
		name: "nanoseconds (us)",
		raw:  `22us`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 22, Unit: "us"},
					},
				},
			}},
		},
	},
	{
		name: "nanoseconds (μs)",
		raw:  `22μs`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 22, Unit: "us"},
					},
				},
			}},
		},
	},
	{
		name: "handle months, min, and ms",
		raw:  `6mo30m500ms`,
		want: &ast.Program{
			Body: []ast.Statement{&ast.ExpressionStatement{
				Expression: &ast.DurationLiteral{
					Values: []ast.Duration{
						{Magnitude: 6, Unit: "mo"},
						{Magnitude: 30, Unit: "m"},
						{Magnitude: 500, Unit: "ms"},
					},
				},
			}},
		},
	},
	// Do we want to accept unordered durations?
	{
		name:      "unordered units are not allowed",
		raw:       `3m6mo`,
		wantErr:   true,
		errString: "unable to match [col 4]",
	},
	{
		name:      "duplicated units are not allowed",
		raw:       `6mo3mo`, //6mo30m3mo
		wantErr:   true,
		errString: "unable to match [col 5]",
	},
	{
		name:      "duplicated units are not allowed",
		raw:       `5mo30m1mo`,
		wantErr:   true,
		errString: "unable to match [col 8]",
	},
}

func TestParse(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mach := NewMachine()
			got := mach.Parse([]byte(tt.raw))

			if !tt.wantErr {
				if !cmp.Equal(tt.want, got, asttest.CompareOptions...) {
					t.Errorf("-want/+got %s", cmp.Diff(tt.want, got, asttest.CompareOptions...))
				}
			} else {
				if got != nil {
					t.Errorf("-want/+got %s", cmp.Diff(nil, got))
				}
				if tt.errString != mach.Err().Error() {
					t.Errorf("-want/+got %s", cmp.Diff(tt.errString, mach.Err().Error()))
				}
			}

		})
	}
}
