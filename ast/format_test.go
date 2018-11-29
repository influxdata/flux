package ast_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
	"github.com/pkg/errors"
)

var skip = map[string]string{
	"array_expr":        "without pars -> bad syntax, with pars formatting removes them",
	"conditional":       "how is a conditional expression defined in spec?",
	"duration_multiple": "how are multiple duration values meant to be used?",
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "arrow_fn",
			query: `(r)=>r.user=="user1"`,
		},
		{
			name:  "fn_decl",
			query: `add=(a,b)=>a+b`,
		},
		{
			name:  "fn_call",
			query: `add(a:1,b:2)`,
		},
		{
			name:  "object",
			query: `{a:1,b:{c:11,d:12}}`,
		},
		{
			name:  "member",
			query: `object.property`,
		},
		{
			name: "array",
			query: `a=[1,2,3]
a[i]`,
		},
		{
			name:  "array_expr",
			query: `a[(i+1)]`,
		},
		{
			name:  "conditional",
			query: `test?cons:alt`,
		},
		{
			name:  "float",
			query: `0.1`,
		},
		{
			name:  "duration",
			query: `365d`,
		},
		{
			name:  "duration_multiple",
			query: `1m1d1s`,
		},
		{
			name:  "time",
			query: `2018-05-22T19:53:00Z`,
		},
		{
			name:  "regexp",
			query: `/^\w+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$/`,
		},
		{
			name:  "return",
			query: `return 42`,
		},
		{
			name:  "option",
			query: `option foo={a:1}`,
		},
		{
			name:  "simple",
			query: `from(bucket:"testdb")|>range(start:2018-05-23T13:09:22.885021542Z)`,
		},
		{
			name:  "medium",
			query: `from(bucket:"testdb")|>range(start:2018-05-20T19:53:26Z)|>filter(fn:(r)=>r.name=~/.*0/)|>group(by:["_measurement","_start"])|>map(fn:(r)=>{_time:r._time,io_time:r._value})`,
		},
		{
			name: "complex",
			query: `left=from(bucket:"test")|>range(start:2018-05-22T19:53:00Z,stop:2018-05-22T19:55:00Z)|>drop(columns:["_start","_stop"])|>filter(fn:(r)=>r.user=="user1")|>group(by:["user"])
right=from(bucket:"test")|>range(start:2018-05-22T19:53:00Z,stop:2018-05-22T19:55:00Z)|>drop(columns:["_start","_stop"])|>filter(fn:(r)=>r.user=="user2")|>group(by:["_measurement"])
join(tables:{left:left,right:right},on:["_time","_measurement"])`,
		},
		{
			name: "option",
			query: `option task={name:"foo",every:1h,delay:10m,cron:"02***",retry:5}
from(bucket:"test")|>range(start:2018-05-22T19:53:26Z)|>window(every:task.every)|>group(by:["_field","host"])|>sum()|>to(bucket:"test",tagColumns:["host","_field"])`,
		},
		{
			name: "functions",
			query: `foo=()=>from(bucket:"testdb")
bar=(x=<-)=>x|>filter(fn:(r)=>r.name=~/.*0/)
baz=(y=<-)=>y|>map(fn:(r)=>{_time:r._time,io_time:r._value})
foo()|>bar()|>baz()`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if reason, ok := skip[tc.name]; ok {
				t.Skip(reason)
			}

			originalProgram, err := parser.NewAST(tc.query)
			if err != nil {
				t.Fatal(errors.Wrapf(err, "original program has bad syntax:\n%s", tc.query))
			}

			stringResult := ast.Format(originalProgram)

			if tc.query != stringResult {
				t.Errorf("\nin:\n%s\nout:\n%s\n", tc.query, stringResult)
			}
		})
	}
}
