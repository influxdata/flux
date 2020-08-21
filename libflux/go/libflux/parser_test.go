package libflux_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/libflux/go/libflux"
)

func TestParse(t *testing.T) {
	text := `
package main

from(bucket: "telegraf")
	|> range(start: -5m)
	|> mean()
`
	ast := libflux.ParseString(text)
	if err := ast.GetError(); err != nil {
		t.Fatal(err)
	}

	jsonBuf, err := ast.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("json has %v bytes:\n%v\n", len(jsonBuf), string(jsonBuf))

	fbBuf, err := ast.MarshalFB()
	if err != nil {
		panic(err)
	}
	fmt.Printf("flatbuffer has %v bytes\n", len(fbBuf))

	ast.Free()
}

func TestASTPkg_GetError(t *testing.T) {
	src := `x = 1 + / 3`
	ast := libflux.ParseString(src)
	defer ast.Free()
	err := ast.GetError()
	if err == nil {
		t.Fatal("expected parse error, got none")
	}
	if want, got := "error at @1:9-1:10: invalid expression: invalid token for primary expression: DIV", err.Error(); want != got {
		t.Error("unexpected parse error; -want/+got:\n ", cmp.Diff(want, got))
	}

}

func TestMergePackages(t *testing.T) {
	outPkg := libflux.ParseString(`
package foo

a = 1`)
	inPkg := libflux.ParseString(`
package foo

b = 1`)
	err := libflux.MergePackages(outPkg, inPkg)
	if err != nil {
		t.Fatal(err)
	}

	gotFB, err := outPkg.MarshalFB()
	if err != nil {
		t.Fatal(err)
	}
	got := ast.DeserializeFromFlatBuffer(gotFB)

	want := &ast.Package{
		Package: "foo",
		Files: []*ast.File{
			{
				Name:     "",
				Metadata: "parser-type=rust",
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
				Name:     "",
				Metadata: "parser-type=rust",
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: "foo"},
				},
				Body: []ast.Statement{
					&ast.VariableAssignment{
						ID:   &ast.Identifier{Name: "b"},
						Init: &ast.IntegerLiteral{Value: 1},
					},
				},
			},
		},
	}

	if !cmp.Equal(got, want, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf(
			"MergePackages unexpected packages -want/+got:\n %v",
			cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...),
		)
	}
}

func TestMergePackagesWithErrors(t *testing.T) {
	fooPkg := libflux.Parse("foo.flux", `
package foo

a = 1`)
	barPkg := libflux.Parse("bar.flux", `
package bar

c = 3`)
	noClausePkg := libflux.Parse("no_pkg.flux", `d = 7`)

	testCases := []struct {
		name    string
		outPkg  *libflux.ASTPkg
		inPkg   *libflux.ASTPkg
		wantErr error
	}{
		{
			name:    "packages clauses don't match",
			outPkg:  barPkg,
			inPkg:   fooPkg,
			wantErr: errors.New(`failed to merge packages: error at foo.flux@2:1-2:12: file is in package "foo", but other files are in package "bar"`),
		},
		{
			name:    "input package has no package clause",
			outPkg:  barPkg,
			inPkg:   noClausePkg,
			wantErr: errors.New(`failed to merge packages: error at no_pkg.flux@1:1-1:6: file is in default package "main", but other files are in package "bar"`),
		},
		{
			name:    "output package has no package clause",
			outPkg:  noClausePkg,
			inPkg:   fooPkg,
			wantErr: errors.New(`failed to merge packages: error at foo.flux@2:1-2:12: file is in package "foo", but other files are in package "main"`),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			gotErr := libflux.MergePackages(testCase.outPkg, testCase.inPkg)
			if gotErr == nil {
				t.Fatal("\nGot no error, expected:", testCase.wantErr, "\ngot: \n", gotErr)
			}
			if diff := cmp.Diff(testCase.wantErr.Error(), gotErr.Error()); diff != "" {
				t.Fatalf("unexpected error: -want/+got: %v", diff)
			}
		})
	}
}

func TestASTPkg_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		fluxFile string
	}{
		{
			name: "simple",
			fluxFile: `
import "foo"
x = foo.y
`,
		},
		{
			name: "every AST node 1",
			fluxFile: `
package mypkg
import "my_other_pkg"
import "yet_another_pkg"
option now = () => (2030-01-01T00:00:00Z)
option foo.bar = "baz"
builtin foo : int

test aggregate_window_empty = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) =>
        table
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:55:00Z)
            |> aggregateWindow(every: 30s, fn: sum),
})
`,
		},
		{
			name: "every AST node 2",
			fluxFile: `
a

arr = [0, 1, 2]
f = (i) => i
ff = (i=<-, j) => {
  k = i + j
  return k
}
b = z and y
b = z or y
o = {red: "red", "blue": 30}
empty_obj = {}
m = o.red
i = arr[0]
n = 10 - 5 + 10
n = 10 / 5 * 10
m = 13 % 3
p = 2^10
b = 10 < 30
b = 10 <= 30
b = 10 > 30
b = 10 >= 30
eq = 10 == 10
neq = 11 != 10
b = not false
e = exists o.red
tables |> f()
fncall = id(v: 20)
fncall2 = foo(v: 20, w: "bar")
fncall_short_form_arg(arg)
fncall_short_form_args(arg0, arg1)
v = if true then 70.0 else 140.0
ans = "the answer is ${v}"
paren = (1)

i = 1
f = 1.0
s = "foo"
d = 10s
b = true
dt = 2030-01-01T00:00:00Z
re =~ /foo/
re !~ /foo/
`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// src -> rust AST -> rustJSONA -> Go AST -> goJSON -> Rust AST -> rustJSONB
			// Compare rustJSONA and rustJSONB
			astPkgA := libflux.ParseString(tc.fluxFile)
			rustJSONA, err := astPkgA.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var goAST ast.Package
			if err := json.Unmarshal(rustJSONA, &goAST); err != nil {
				t.Fatal(err)
			}
			goJSON, err := json.Marshal(&goAST)
			if err != nil {
				t.Fatal(err)
			}
			astPkgB, err := libflux.ParseJSON(goJSON)
			if err != nil {
				t.Fatal(err)
			}
			rustJSONB, err := astPkgB.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			compareIndentedJSON(t, rustJSONA, rustJSONB)
		})
	}
}

func mustIndent(t *testing.T, bs []byte) []byte {
	t.Helper()
	var bb bytes.Buffer
	if err := json.Indent(&bb, bs, "", "    "); err != nil {
		t.Fatalf("could not indent: %v", err)
	}
	return bb.Bytes()
}

func compareIndentedJSON(t *testing.T, jsonA, jsonB []byte) {
	t.Helper()
	if diff := cmp.Diff(string(mustIndent(t, jsonA)), string(mustIndent(t, jsonB))); diff != "" {
		t.Errorf("JSON A and JSON B differed; -a/+b:\n%v", diff)
	}
}
