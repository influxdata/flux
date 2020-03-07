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
	ast := libflux.Parse(text)

	jsonBuf, err := ast.ToJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("json has %v bytes:\n%v\n", len(jsonBuf.Buffer), string(jsonBuf.Buffer))
	jsonBuf.Free()

	mbuf, err := ast.MarshalFB()
	if err != nil {
		panic(err)
	}
	fmt.Printf("flatbuffer has %v bytes, offset %v.\n", len(mbuf.Buffer), mbuf.Offset)
	mbuf.Free()
	ast.Free()
}

func TestMergePackages(t *testing.T) {
	outPkg := libflux.Parse(`
package foo

a = 1`)
	inPkg := libflux.Parse(`
package foo

b = 1`)
	err := libflux.MergePackages(outPkg, inPkg)
	if err != nil {
		t.Fatal(err)
	}

	mbuf, err := outPkg.MarshalFB()
	if err != nil {
		t.Fatal(err)
	}
	got := ast.DeserializeFromFlatBuffer(mbuf)
	mbuf.Free()

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
	fooPkg := libflux.Parse(`
package foo

a = 1`)
	barPkg := libflux.Parse(`
package bar

c = 3`)
	noClausePkg := libflux.Parse(``)

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
			wantErr: errors.New("failed to merge packages: file's package clause: foo does not match package output package clause: bar"),
		},
		{
			name:    "input package has no package clause",
			outPkg:  barPkg,
			inPkg:   noClausePkg,
			wantErr: errors.New("failed to merge packages: current file does not have a package clause"),
		},
		{
			name:    "output package has no package clause",
			outPkg:  noClausePkg,
			inPkg:   fooPkg,
			wantErr: errors.New("failed to merge packages: output package does not have a package clause"),
		},
	}

	for _, testCase := range testCases {
		gotErr := libflux.MergePackages(testCase.outPkg, testCase.inPkg)
		if gotErr == nil {
			t.Fatal("\nGot no error, expected:", testCase.wantErr, "\ngot: \n", gotErr)
		}
		if diff := cmp.Diff(testCase.wantErr.Error(), gotErr.Error()); diff != "" {
			t.Fatalf("unexpected error: -want/+got: %v", diff)
		}
	}
}
