package ast_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
	"github.com/pkg/errors"
)

func TestEditor(t *testing.T) {
	testCases := []struct {
		name   string
		in     string
		edited string
		editor *ast.OptionEditor
	}{
		{
			name:   "leaves_unchanged",
			in:     `from(bucket:"testdb") |> range(start: 2018-05-23T13:09:22.885021542Z)`,
			edited: `from(bucket:"testdb")|>range(start:2018-05-23T13:09:22.885021542Z)`,
			editor: ast.NewOptionEditor(map[string]interface{}{
				"bucket": "foo",
				"start":  "yesterday", // should ignore this
				"stop":   "tomorrow",  // should ignore this
			}),
		},
		{
			name: "edits",
			in: `option task = {name: "foo",every: 1h,delay: 10m,cron: "0 2 * * *",retry: 5}
from(bucket:"testdb") |> range(start: 2018-05-23T13:09:22.885021542Z)`,
			edited: `option task={name:"bar",every:2d,delay:42m,cron:"buz",retry:10}
from(bucket:"testdb")|>range(start:2018-05-23T13:09:22.885021542Z)`,
			editor: ast.NewOptionEditor(map[string]interface{}{
				"bucket": "foo",       // should ignore this
				"start":  "yesterday", // should ignore this
				"name":   "bar",
				"every":  ast.Duration{Magnitude: 2, Unit: "d"},
				"delay":  ast.Duration{Magnitude: 42, Unit: "m"},
				"cron":   "buz",
				"retry":  10,
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := parser.NewAST(tc.in)
			if err != nil {
				t.Fatal(errors.Wrapf(err, "input program has bad syntax:\n%s", tc.in))
			}

			ast.Walk(tc.editor, p)
			out := ast.Format(p)

			if tc.edited != out {
				t.Errorf("\nexpected:\n%s\nedited:\n%s\n", tc.edited, out)
			}
		})
	}
}
