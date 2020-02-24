package semantic_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
)

func TestToAST(t *testing.T) {
	const prelude = "package main\n"
	for _, tt := range []struct {
		name string
		s    string
	}{
		{
			name: "array expression",
			s:    `[1, 2, 3]`,
		},
		{
			name: "object expression",
			s:    `{a: 1, b: "2"}`,
		},
		{
			name: "function expression",
			s:    `(a) => a`,
		},
		{
			name: "function expression with default",
			s:    `(a, b=2) => a + b`,
		},
		{
			name: "function block",
			s: `(a, b) => {
	c = a + b
	return c
}`,
		},
		{
			name: "call expression",
			s:    `(() => 2)()`,
		},
		{
			name: "call expression with argument",
			s:    `((v) => v)(v:2)`,
		},
		{
			name: "identifier",
			s: `my_value = 2
my_value
`,
		},
		{
			name: "boolean literal",
			s:    `true`,
		},
		{
			name: "date time literal",
			s:    `2019-05-15T12:00:00Z`,
		},
		{
			name: "duration literal",
			s:    `5s`,
		},
		{
			name: "float literal",
			s:    `1.0`,
		},
		{
			name: "integer literal",
			s:    `5`,
		},
		{
			name: "regexp literal",
			s:    `/abc/`,
		},
		{
			name: "string literal",
			s:    `"hello world"`,
		},
		{
			name: `binary expression`,
			s:    `"gpu" == "cpu"`,
		},
		{
			name: "logical expression",
			s:    `true or false`,
		},
		{
			name: "imports",
			s:    `import "csv"`,
		},
		{
			name: "option statement",
			s:    `option now = () => 2019-05-22T00:00:00Z`,
		},
		{
			name: "option statement for subpackage",
			s: `import c "csv"
import "testing"
option testing.loadMem = (csv) => c.from(csv: csv)
`,
		},
		{
			name: "variable assignment",
			s:    `a = 1`,
		},
		{
			name: "test statement",
			s:    `test foo = {}`,
		},
		{
			name: "conditional expression",
			s:    `if 1 == 2 then 5 else 3`,
		},
		{
			name: "index expression",
			s:    `[1, 2, 3][1]`,
		},
		{
			name: "unary negative expression",
			s:    `-1d`,
		},
		{
			name: "unary exists operator",
			s: `import "internal/testutil"
r = testutil.makeRecord(o: {a: 1})
exists r.b
`,
		},
		{
			name: "string expression",
			s: `name = "World"
"Hello ${name}!"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			want, err := runtime.AnalyzeSource(prelude + tt.s)
			if err != nil {
				t.Fatalf("unexpected error analyzing source: %s", err)
			}

			got, err := runtime.AnalyzeSource(ast.Format(semantic.ToAST(want)))
			if err != nil {
				t.Fatalf("unexpected error analyzing generated AST: %s", err)
			}

			if !cmp.Equal(want, got, semantictest.CmpOptions...) {
				t.Fatalf("unexpected semantic graph -want/+got:\n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
			}
		})
	}
}
