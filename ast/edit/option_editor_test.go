package edit_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func TestEditor(t *testing.T) {
	testCases := []struct {
		name        string
		in          string
		edited      string
		unchanged   bool
		errorWanted bool
		edit        func(node ast.Node) (bool, error)
	}{
		{
			name: "no_option",
			in: `from(bucket: "test")
	|> range(start: 2018-05-23T13:09:22.885021542Z)`,
			unchanged: true,
			edit: func(node ast.Node) (bool, error) {
				return edit.Option(node, "from", nil)
			},
		},
		{
			name:      "option_wrong_id",
			in:        `option foo = 1`,
			unchanged: true,
			edit: func(node ast.Node) (bool, error) {
				return edit.Option(node, "bar", nil)
			},
		},
		{
			name:   "qualified_option",
			in:     `option alert.state = 0`,
			edited: `option alert.state = 1`,
			edit: func(node ast.Node) (bool, error) {
				literal := &ast.IntegerLiteral{Value: int64(1)}
				return edit.Option(node, "alert.state", edit.OptionValueFn(literal))
			},
		},
		{
			name: "sets_option",
			in: `option foo = 1
option bar = 1`,
			edited: `option foo = 1
option bar = 42`,
			edit: func(node ast.Node) (bool, error) {
				literal := &ast.IntegerLiteral{Value: int64(42)}
				return edit.Option(node, "bar", edit.OptionValueFn(literal))
			},
		},
		{
			name: "updates_object",
			in: `option foo = 1
option task = {
	name: "bar",
	every: 1m,
	delay: 1m,
	cron: "20 * * *",
	retry: 5,
}`,
			edited: `option foo = 1
option task = {
	name: "bar",
	every: 2hr3m10s,
	delay: 42m,
	cron: "buz",
	retry: 10,
}`,
			edit: func(node ast.Node) (bool, error) {
				every, err := ast.ParseDuration("2hr3m10s")
				if err != nil {
					t.Fatal(err)
				}
				delay, err := ast.ParseDuration("42m")
				if err != nil {
					t.Fatal(err)
				}
				return edit.Option(node, "task", edit.OptionObjectFn(map[string]ast.Expression{
					"every": &ast.DurationLiteral{Values: every},
					"delay": &ast.DurationLiteral{Values: delay},
					"cron":  &ast.StringLiteral{Value: "buz"},
					"retry": &ast.IntegerLiteral{Value: 10},
				}))
			},
		},
		{
			name: "error_key_not_found",
			in: `option foo = 1
option task = {
	name: "bar",
	every: 1m,
	delay: 1m,
	cron: "20 * * *",
	retry: 5,
}`,
			errorWanted: true,
			edit: func(node ast.Node) (bool, error) {
				every, err := ast.ParseDuration("2hr")
				if err != nil {
					t.Fatal(err)
				}
				return edit.Option(node, "task", edit.OptionObjectFn(map[string]ast.Expression{
					"foo":   &ast.StringLiteral{Value: "foo"}, // should cause error
					"every": &ast.DurationLiteral{Values: every},
				}))
			},
		},
		{
			name:   "sets_option_to_array",
			in:     `option foo = "edit me"`,
			edited: `option foo = [1, 2, 3, 4]`,
			edit: func(node ast.Node) (bool, error) {
				literal := &ast.ArrayExpression{Elements: []ast.Expression{
					&ast.IntegerLiteral{Value: 1},
					&ast.IntegerLiteral{Value: 2},
					&ast.IntegerLiteral{Value: 3},
					&ast.IntegerLiteral{Value: 4},
				}}
				return edit.Option(node, "foo", edit.OptionValueFn(literal))
			},
		},
		{
			name:   "sets_option_to_object",
			in:     `option foo = "edit me"`,
			edited: `option foo = {x: "x", y: "y"}`,
			edit: func(node ast.Node) (bool, error) {
				literal := &ast.ObjectExpression{
					Properties: []*ast.Property{{
						Key:   &ast.Identifier{Name: "x"},
						Value: &ast.StringLiteral{Value: "x"},
					}, {
						Key:   &ast.Identifier{Name: "y"},
						Value: &ast.StringLiteral{Value: "y"},
					},
					}}
				return edit.Option(node, "foo", edit.OptionValueFn(literal))
			},
		},
		{
			name:   "sets_option_mixed",
			in:     `option foo = "edit me"`,
			edited: `option foo = {x: {a: [1, 2, 3]}, y: [[1], [2, 3]], z: [{a: 1}, {b: 2}]}`,
			edit: func(node ast.Node) (bool, error) {
				x := &ast.ObjectExpression{
					Properties: []*ast.Property{{
						Key: &ast.Identifier{Name: "a"},
						Value: &ast.ArrayExpression{Elements: []ast.Expression{
							&ast.IntegerLiteral{Value: 1},
							&ast.IntegerLiteral{Value: 2},
							&ast.IntegerLiteral{Value: 3},
						},
						},
					}}}
				y := &ast.ArrayExpression{Elements: []ast.Expression{
					&ast.ArrayExpression{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
					&ast.ArrayExpression{Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 2},
						&ast.IntegerLiteral{Value: 3},
					}},
				}}
				z := &ast.ArrayExpression{Elements: []ast.Expression{
					&ast.ObjectExpression{Properties: []*ast.Property{{Key: &ast.Identifier{Name: "a"}, Value: &ast.IntegerLiteral{Value: 1}}}},
					&ast.ObjectExpression{Properties: []*ast.Property{{Key: &ast.Identifier{Name: "b"}, Value: &ast.IntegerLiteral{Value: 2}}}},
				}}

				literal := &ast.ObjectExpression{
					Properties: []*ast.Property{{
						Key:   &ast.Identifier{Name: "x"},
						Value: x,
					}, {
						Key:   &ast.Identifier{Name: "y"},
						Value: y,
					}, {
						Key:   &ast.Identifier{Name: "z"},
						Value: z,
					},
					}}
				return edit.Option(node, "foo", edit.OptionValueFn(literal))
			},
		},
		{
			name:   "sets_option_to_function_call",
			in:     `option location = "edit me"`,
			edited: `option location = loadLocation(name: "America/Denver")`,
			edit: func(node ast.Node) (bool, error) {
				literal := &ast.CallExpression{
					Callee: &ast.Identifier{Name: "loadLocation"},
					Arguments: []ast.Expression{&ast.ObjectExpression{Properties: []*ast.Property{
						{
							Key:   &ast.Identifier{Name: "name"},
							Value: &ast.StringLiteral{Value: "America/Denver"},
						},
					}}},
				}
				return edit.Option(node, "location", edit.OptionValueFn(literal))
			},
		},
		{
			name: "sets_option_to_function",
			in:   `option now = "edit me"`,
			edited: `option now = () =>
	(2018-12-03T20:52:48.464942Z)`,
			edit: func(node ast.Node) (bool, error) {
				t, err := values.ParseTime("2018-12-03T20:52:48.464942000Z")
				if err != nil {
					panic(err)
				}

				literal := &ast.FunctionExpression{
					Params: []*ast.Property{},
					Body:   &ast.DateTimeLiteral{Value: t.Time()},
				}
				return edit.Option(node, "now", edit.OptionValueFn(literal))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		if tc.unchanged {
			tc.edited = tc.in
		}

		t.Run(tc.name, func(t *testing.T) {
			p := parser.ParseSource(tc.in)
			if ast.Check(p) > 0 {
				err := ast.GetError(p)
				t.Fatal(errors.Wrapf(err, "input source has bad syntax:\n%s", tc.in))
			}

			edited, err := tc.edit(p)
			if err != nil && tc.errorWanted {
				return
			}

			if err != nil {
				t.Fatal(errors.Wrap(err, "got unexpected error from edit"))
			}

			if edited && tc.unchanged {
				t.Fatal("unexpected option edit")
			}

			out := ast.Format(p.Files[0])

			if out != tc.edited {
				t.Errorf("\nexpected:\n%s\nedited:\n%s\n", tc.edited, out)
			}
		})
	}
}
