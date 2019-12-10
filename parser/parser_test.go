package parser_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/parser"
)

func wantMetadata() string {
	if os.Getenv("FLUX_PARSER_TYPE") == "rust" {
		return "parser-type=rust"
	} else {
		return "parser-type=go"
	}
}

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
	wantMeta := wantMetadata()
	want := map[string]*ast.Package{
		"foo": &ast.Package{
			Package: "foo",
			Files: []*ast.File{
				{
					Name:     "a.flux",
					Metadata: wantMeta,
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
					Metadata: wantMeta,
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
				Metadata: wantMeta,
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
		Metadata: wantMetadata(),
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
			Metadata: wantMetadata(),
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
