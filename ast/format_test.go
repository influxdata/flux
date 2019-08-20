package ast_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
)

var skip = map[string]string{
	"array_expr":  "without pars -> bad syntax, with pars formatting removes them",
	"conditional": "how is a conditional expression defined in spec?",
}

type formatTestCase struct {
	name       string
	script     string
	shouldFail bool
}

// formatTestHelper tests that a raw script has valid syntax and
// that it has the same value if parsed and then formatted.
func formatTestHelper(t *testing.T, testCases []formatTestCase) {
	t.Helper()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if reason, ok := skip[tc.name]; ok {
				t.Skip(reason)
			}

			pkg := parser.ParseSource(tc.script)
			if ast.Check(pkg) > 0 {
				err := ast.GetError(pkg)
				t.Fatalf("source has bad syntax: %s\n%s", err, tc.script)
			}

			stringResult := ast.Format(pkg.Files[0])

			if tc.script != stringResult {
				if !tc.shouldFail {
					t.Errorf("unexpected output: -want/+got:\n %s", cmp.Diff(tc.script, stringResult))
				}
			}
		})
	}
}

func TestFormat_Nodes(t *testing.T) {
	testCases := []formatTestCase{
		{
			name:   "string interpolation",
			script: `"a + b = ${a + b}"`,
		},
		{
			name:   "binary_op",
			script: `1 + 1 - 2`,
		},
		{
			name:   "binary_op 2",
			script: `2 ^ 4`,
		},
		{
			name: "arrow_fn",
			script: `(r) =>
	(r.user == "user1")`,
		},
		{
			name: "fn_decl",
			script: `add = (a, b) =>
	(a + b)`,
		},
		{
			name:   "fn_call",
			script: `add(a: 1, b: 2)`,
		},
		{
			name:   "object",
			script: `{a: 1, b: {c: 11, d: 12}}`,
		},
		{
			name:   "object with",
			script: `{foo with a: 1, b: {c: 11, d: 12}}`,
		},
		{
			name:   "implicit key object literal",
			script: `{a, b, c}`,
		},
		{
			name:   "object with string literal keys",
			script: `{"a": 1, "b": 2}`,
		},
		{
			name:   "object with mixed keys",
			script: `{"a": 1, b: 2}`,
		},
		{
			name:   "member ident",
			script: `object.property`,
		},
		{
			name:   "member string literal",
			script: `object["property"]`,
		},
		{
			name: "array",
			script: `a = [1, 2, 3]

a[i]`,
		},
		{
			name:   "array_expr",
			script: `a[(i+1)]`,
		},
		{
			name:   "conditional",
			script: `test?cons:alt`,
		},
		{
			name:   "float",
			script: `0.1`,
		},
		{
			name:   "duration",
			script: `365d`,
		},
		{
			name:   "duration_multiple",
			script: `1d1m1s`,
		},
		{
			name:   "time",
			script: `2018-05-22T19:53:00Z`,
		},
		{
			name:   "regexp",
			script: `/^\w+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$/`,
		},
		{
			name:   "regexp_escape",
			script: `/^http:\/\/\w+\.com$/`,
		},
		{
			name:   "return",
			script: `return 42`,
		},
		{
			name:   "option",
			script: `option foo = {a: 1}`,
		},
		{
			name:   "qualified option",
			script: `option alert.state = "Warning"`,
		},
		{
			name:   "test statement",
			script: `test mean = {want: 0, got: 0}`,
		},
		{
			name:   "conditional",
			script: "if a then b else c",
		},
		{
			name:   "conditional with more complex expressions",
			script: `if not a or b and c then 2 / (3 * 2) else obj.a(par: "foo")`,
		},
		{
			name: "nil_value_as_default",
			script: `foo = (arg=[]) =>
	(1)`,
		},
		{
			name: "non_nil_value_as_default",
			script: `foo = (arg=[1, 2]) =>
	(1)`,
		},
		{
			name: "block",
			script: `foo = () => {
	foo(f: 1)
	1 + 1
}`,
		},
		{
			name:   "string",
			script: `"foo"`,
		},
		{
			name: "string multiline",
			script: `"this is
a string
with multiple lines"`,
		},
		{
			name:   "string with escape",
			script: `"foo \\ \" \r\n"`,
		},
		{
			name:   "string with byte value",
			script: `"\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e"`,
		},
		{
			name:   "package",
			script: "package foo\n",
		},
		{
			name: "imports",
			script: `import "path/foo"
import bar "path/bar"`,
		},
		{
			name: "no_package",
			script: `import foo "path/foo"

foo.from(bucket: "testdb")
	|> range(start: 2018-05-20T19:53:26Z)`,
		},
		{
			name: "no_import",
			script: `package foo


from(bucket: "testdb")
	|> range(start: 2018-05-20T19:53:26Z)`,
		},
		{
			name: "package_import",
			script: `package foo


import "path/foo"
import bar "path/bar"

from(bucket: "testdb")
	|> range(start: 2018-05-20T19:53:26Z)`,
		},
		{
			name: "simple",
			script: `from(bucket: "testdb")
	|> range(start: 2018-05-20T19:53:26Z)
	|> filter(fn: (r) =>
		(r.name =~ /.*0/))
	|> group(by: ["_measurement", "_start"])
	|> map(fn: (r) =>
		({_time: r._time, io_time: r._value}))`,
		},
		{
			name: "medium",
			script: `from(bucket: "testdb")
	|> range(start: 2018-05-20T19:53:26Z)
	|> filter(fn: (r) =>
		(r.name =~ /.*0/))
	|> group(by: ["_measurement", "_start"])
	|> map(fn: (r) =>
		({_time: r._time, io_time: r._value}))`,
		},
		{
			name: "complex",
			script: `left = from(bucket: "test")
	|> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
	|> drop(columns: ["_start", "_stop"])
	|> filter(fn: (r) =>
		(r.user == "user1"))
	|> group(by: ["user"])
right = from(bucket: "test")
	|> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
	|> drop(columns: ["_start", "_stop"])
	|> filter(fn: (r) =>
		(r.user == "user2"))
	|> group(by: ["_measurement"])

join(tables: {left: left, right: right}, on: ["_time", "_measurement"])`,
		},
		{
			name: "option",
			script: `option task = {
	name: "foo",
	every: 1h,
	delay: 10m,
	cron: "02***",
	retry: 5,
}

from(bucket: "test")
	|> range(start: 2018-05-22T19:53:26Z)
	|> window(every: task.every)
	|> group(by: ["_field", "host"])
	|> sum()
	|> to(bucket: "test", tagColumns: ["host", "_field"])`,
		},
		{
			name: "functions",
			script: `foo = () =>
	(from(bucket: "testdb"))
bar = (x=<-) =>
	(x
		|> filter(fn: (r) =>
			(r.name =~ /.*0/)))
baz = (y=<-) =>
	(y
		|> map(fn: (r) =>
			({_time: r._time, io_time: r._value})))

foo()
	|> bar()
	|> baz()`,
		},
		{
			name: "multi_indent",
			script: `_sortLimit = (n, desc, columns=["_value"], tables=<-) =>
	(tables
		|> sort(columns: columns, desc: desc)
		|> limit(n: n))
_highestOrLowest = (n, _sortLimit, reducer, columns=["_value"], by, tables=<-) =>
	(tables
		|> group(by: by)
		|> reducer()
		|> group(none: true)
		|> _sortLimit(n: n, columns: columns))
highestAverage = (n, columns=["_value"], by, tables=<-) =>
	(tables
		|> _highestOrLowest(
			n: n,
			columns: columns,
			by: by,
			reducer: (tables=<-) =>
				(tables
					|> mean(columns: [columns[0]])),
			_sortLimit: top,
		))`,
		},
	}

	formatTestHelper(t, testCases)
}

func TestFormat_Associativity(t *testing.T) {
	testCases := []formatTestCase{
		{
			name:   "math no pars",
			script: `a * b + c / d - e * f`,
		},
		{
			name:   "math with pars",
			script: `(a * b + c / d - e) * f`,
		},
		{
			name:   "minus before parens",
			script: `r._value - (1 * 2 + 4 / 6 - 10)`,
		},
		{
			name:       "plus with unintended parens 1",
			script:     `(1 + 2) + 3`,
			shouldFail: true,
		},
		{
			name:       "plus with unintended parens 2",
			script:     `1 + (2 + 3)`,
			shouldFail: true,
		},
		{
			name:       "minus with unintended parens",
			script:     `(1 - 2) - 3`,
			shouldFail: true,
		},
		{
			name:   "minus with intended parens",
			script: `1 - (2 - 3)`,
		},
		{
			name:   "minus no parens",
			script: `1 - 2 - 3`,
		},
		{
			name:   "div with parens",
			script: `1 / (2 * 3)`,
		},
		{
			name:   "div no parens",
			script: `1 / 2 * 3`,
		},
		{
			name:       "div with unintended parens",
			script:     `(1 / 2) * 3`,
			shouldFail: true,
		},
		{
			name:   "math with more pars",
			script: `(a * (b + c) / d / e * (f + g) - h) * i * j / (k + l)`,
		},
		{
			name:   "logic",
			script: `a or b and c`,
		},
		{
			name:   "logic with pars",
			script: `(a or b) and c`,
		},
		{
			name:   "logic with comparison",
			script: `a == 0 or b != 1 and c > 2`,
		},
		{
			name:   "logic with comparison with pars",
			script: `(a == 0 or b != 1) and c > 2`,
		},
		{
			name:   "logic and math",
			script: `a * b + c * d != 0 or not e == 1 and f == g`,
		},
		{
			name:   "logic and math with pars",
			script: `(a * (b + c) * d != 0 or not e == 1) and f == g`,
		},
		{
			name:   "unary",
			script: `not b and c`,
		},
		{
			name:   "unary with pars",
			script: `not (b and c) and exists d or exists (e and f)`,
		},
		{
			name:   "unary negative duration",
			script: `-30s`,
		},
		{
			name:   "unary positive duration",
			script: `+30s`,
		},
		{
			name:   "function call with pars",
			script: `(a + b * c == 0)(foo: "bar")`,
		},
		{
			name:   "member with pars",
			script: `((a + b) * c)._value`,
		},
		{
			name:   "index with pars",
			script: `((a - b) / (c + d))[3]`,
		},
		{
			name: "misc",
			script: `foo = (a) =>
	((bar or buz)(arg: a + 1) + (a / (b + c))[42])

foo(a: (obj1 and obj2 or obj3).idk)`,
		},
	}

	formatTestHelper(t, testCases)
}

func TestFormat_Raw(t *testing.T) {
	testCases := []struct {
		name   string
		node   ast.Node
		script string
	}{
		{
			name: "string escape",
			node: &ast.StringLiteral{
				Value: "foo \\ \" \r\n",
			},
			script: "\"foo \\\\ \\\" \r\n\"",
		},
		{
			name: "package multiple files",
			node: &ast.Package{
				Package: "foo",
				Files: []*ast.File{
					{
						Name: "a.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "foo"},
						},
						Body: []ast.Statement{
							&ast.VariableAssignment{
								ID:   &ast.Identifier{Name: "a"},
								Init: &ast.IntegerLiteral{Value: 1},
							},
						},
					},
					{
						Name: "b.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "foo"},
						},
						Body: []ast.Statement{
							&ast.VariableAssignment{
								ID:   &ast.Identifier{Name: "b"},
								Init: &ast.IntegerLiteral{Value: 2},
							},
						},
					},
				},
			},
			script: `package foo
// a.flux
a = 1

// b.flux
b = 2`,
		},
		{
			name: "package file no name",
			node: &ast.Package{
				Package: "foo",
				Files: []*ast.File{
					{
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "foo"},
						},
						Body: []ast.Statement{
							&ast.VariableAssignment{
								ID:   &ast.Identifier{Name: "a"},
								Init: &ast.IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
			script: `package foo
a = 1`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ast.Format(tc.node)
			if tc.script != got {
				t.Errorf("unexpected output: -want/+got:\n %s", cmp.Diff(tc.script, got))
			}
		})
	}
}
