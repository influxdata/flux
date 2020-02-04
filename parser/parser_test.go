package parser_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
)

var parserType = "parser-type=rust"
var ignorePolyType = cmpopts.IgnoreFields(semantic.NativeVariableAssignment{}, "Typ")

func TestParseDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestParseDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string][]byte{
		"a.flux": []byte(`
package foo

a = 1
`),
		"b.flux": []byte(`
package foo

b = 2
`),
		"notes.txt": []byte(`
this should be ignored
`),
		"c.flux": []byte(`
c = 3
`)}

	for name, src := range files {
		f, err := os.Create(filepath.Join(tmpDir, name))
		if err != nil {
			t.Fatal(f)
		}
		f.Write(src)
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	fset := new(token.FileSet)
	got, err := parser.ParseDir(fset, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]*ast.Package{
		"foo": &ast.Package{
			Package: "foo",
			Files: []*ast.File{
				{
					Name:     "a.flux",
					Metadata: parserType,
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
					Name:     "b.flux",
					Metadata: parserType,
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
		"main": &ast.Package{
			Package: "main",
			Files: []*ast.File{{
				Name:     "c.flux",
				Metadata: parserType,
				Body: []ast.Statement{
					&ast.VariableAssignment{
						ID:   &ast.Identifier{Name: "c"},
						Init: &ast.IntegerLiteral{Value: 3},
					},
				},
			}},
		},
	}

	if !cmp.Equal(got, want, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("ParseDir unexpected packages -want/+got:\n%s", cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestParseFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestParseDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fname := "a.flux"
	src := []byte(`
package foo

a = 1
`)
	fpath := filepath.Join(tmpDir, fname)
	f, err := os.Create(fpath)
	if err != nil {
		t.Fatal(f)
	}
	f.Write(src)
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	fset := new(token.FileSet)
	got, err := parser.ParseFile(fset, fpath)
	if err != nil {
		t.Fatal(err)
	}
	want := &ast.File{
		Name:     "a.flux",
		Metadata: parserType,
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: "foo"},
		},
		Body: []ast.Statement{
			&ast.VariableAssignment{
				ID:   &ast.Identifier{Name: "a"},
				Init: &ast.IntegerLiteral{Value: 1},
			},
		},
	}
	if !cmp.Equal(got, want, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("ParseFile unexpected file -want/+got:\n%s", cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestParseSource(t *testing.T) {
	src := `
package foo

a = 1
`

	got := parser.ParseSource(src)
	want := &ast.Package{
		Package: "foo",
		Files: []*ast.File{{
			Name:     "",
			Metadata: parserType,
			Package: &ast.PackageClause{
				Name: &ast.Identifier{Name: "foo"},
			},
			Body: []ast.Statement{
				&ast.VariableAssignment{
					ID:   &ast.Identifier{Name: "a"},
					Init: &ast.IntegerLiteral{Value: 1},
				},
			},
		}},
	}
	if !cmp.Equal(got, want, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("ParseSource unexpected package -want/+got:\n%s", cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestParseTimeLiteral(t *testing.T) {
	inputTime := "2018-01-01"
	got, err := parser.ParseTime(inputTime)
	if err != nil {
		t.Errorf(err.Error())
	}
	want := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	if !cmp.Equal(got.Value, want, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("ParseTimeLiteral failed: %s", cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestMergeExternToSemanticHandle(t *testing.T) {
	extern := &ast.File{
		Name:     "",
		Metadata: "parser-type=rust",
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: ""},
		},
		Body: []ast.Statement{
			&ast.VariableAssignment{
				ID:   &ast.Identifier{Name: "a"},
				Init: &ast.IntegerLiteral{Value: 1},
			},
		},
	}

	source := &ast.Package{
		BaseNode: ast.BaseNode{},
		Path:     "",
		Package:  "",
		Files: []*ast.File{
			{
				Name:     "",
				Metadata: "parser-type=rust",
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: ""},
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

	gotHandle, err := parser.MergeExternToSemanticHandle(extern, source) // semantic Handle; semanticPkg
	if err != nil {
		t.Fatalf(err.Error())
	}

	gotFB, err := gotHandle.MarshalFB() // returns FB byte arr
	if err != nil {
		t.Fatal(err)
	}

	got, err := semantic.DeserializeFromFlatBuffer(gotFB) // got is semantic.Package
	if err != nil {
		t.Fatal(err)
	}

	want := &semantic.Package{
		Package: "",
		Files: []*semantic.File{
			{
				Package: &semantic.PackageClause{
					Name: &semantic.Identifier{Name: ""},
				},
				Body: []semantic.Statement{
					&semantic.NativeVariableAssignment{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.IntegerLiteral{Value: 1},
					},
				},
			},
			{
				Package: &semantic.PackageClause{
					Name: &semantic.Identifier{Name: ""},
				},
				Body: []semantic.Statement{
					&semantic.NativeVariableAssignment{
						Identifier: &semantic.Identifier{Name: "b"},
						Init:       &semantic.IntegerLiteral{Value: 1},
					},
				},
			},
		},
	}

	opts := append(semantictest.CmpOptions, ignorePolyType)

	if !cmp.Equal(got, want, opts...) {
		t.Errorf("ParseASTFileToHandle failed: %s", cmp.Diff(want, got, opts...))
	}
}
