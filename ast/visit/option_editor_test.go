package visit_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/visit"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func TestEditor(t *testing.T) {
	testCases := []struct {
		name   string
		in     string
		edited string
		editor *visit.OptionEditor
	}{
		{
			name: "leaves_unchanged",
			in:   `from(bucket:"testdb") |> range(start: 2018-05-23T13:09:22.885021542Z)`,
			edited: `from(bucket: "testdb")
	|> range(start: 2018-05-23T13:09:22.885021542Z)`,
			editor: visit.NewOptionEditor(map[string]values.Value{
				"bucket": values.New("foo"),       // should ignore this
				"start":  values.New("yesterday"), // should ignore this
				"stop":   values.New("tomorrow"),  // should ignore this
			}),
		},
		{
			name: "edits",
			in: `option task = {name: "foo",every: 1h,delay: 10m,cron: "0 2 * * *",retry: 5}
from(bucket:"testdb") |> range(start: 2018-05-23T13:09:22.885021542Z)`,
			edited: `option task = {
	name: "bar",
	every: 7200000000000ns,
	delay: 2520000000000ns,
	cron: "buz",
	retry: 10,
}

from(bucket: "testdb")
	|> range(start: 2018-05-23T13:09:22.885021542Z)`,
			editor: visit.NewOptionEditor(map[string]values.Value{
				"bucket": values.New("foo"),       // should ignore this
				"start":  values.New("yesterday"), // should ignore this
				"name":   values.New("bar"),
				"every":  values.New(duration("2h")),
				"delay":  values.New(duration("42m")),
				"cron":   values.New("buz"),
				"retry":  values.New(int64(10)),
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

			visit.Walk(tc.editor, p)
			out := ast.Format(p)

			if tc.edited != out {
				t.Errorf("\nexpected:\n%s\nedited:\n%s\n", tc.edited, out)
			}
		})
	}
}

func duration(sd string) values.Duration {
	d, err := values.ParseDuration(sd)
	if err != nil {
		panic(err)
	}

	return values.New(d).Duration()
}
