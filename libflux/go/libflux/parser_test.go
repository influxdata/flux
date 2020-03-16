package libflux_test

import (
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
	fooPkg := libflux.ParseString(`
package foo

a = 1`)
	barPkg := libflux.ParseString(`
package bar

c = 3`)
	noClausePkg := libflux.ParseString(``)

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
			wantErr: errors.New(`failed to merge packages: file's package clause: "foo" does not match package output package clause: "bar"`),
		},
		{
			name:    "input package has no package clause",
			outPkg:  barPkg,
			inPkg:   noClausePkg,
			wantErr: errors.New(`failed to merge packages: current file does not have a package clause, but output package has package clause "bar"`),
		},
		{
			name:    "output package has no package clause",
			outPkg:  noClausePkg,
			inPkg:   fooPkg,
			wantErr: errors.New(`failed to merge packages: output package does not have a package clause, but current file has package clause "foo"`),
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
