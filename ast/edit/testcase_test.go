package edit_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/parser"
)

func TestTestcaseTransform(t *testing.T) {
	expected := []*ast.Package{
		parser.ParseSource(`package main

import "testing"

myVar = 4

testing.assertEqual(got: 2 + 2, want: 4)`),
		parser.ParseSource(`package main
		
import "testing"

myVar = 4

testing.assertEqual(got: 4 - 2, want: 2)`),
	}

	testFile := `package an_test

import "testing"

myVar = 4

testcase test_addition {
	testing.assertEqual(got: 2 + 2, want: 4)
}

testcase test_subtraction {
	testing.assertEqual(got: 4 - 2, want: 2)
}`

	d := parser.ParseSource(testFile)

	names, transformed, err := edit.TestcaseTransform(context.Background(), d, nil)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := []string{"test_addition", "test_subtraction"}, names; !cmp.Equal(want, got) {
		t.Errorf("unexpected test names: -want/+got:\n%s", cmp.Diff(want, got))
	}
	if !cmp.Equal(expected, transformed, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("unexpected match result: -want/+got:\n%s", cmp.Diff(expected, transformed, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestTestcaseTransformNoTestcase(t *testing.T) {
	testFile := `package an_test

import "testing"

myVar = 4`

	d := parser.ParseSource(testFile)

	_, transformed, err := edit.TestcaseTransform(context.Background(), d, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(transformed) != 0 {
		t.Errorf("unexpected package count: want: 0, got: %d", len(transformed))
	}
}

func TestTestcaseTransformImport(t *testing.T) {
	fs := &memoryFilesystem{
		files: map[string]string{
			"a/a_test.flux": `package a_test
import "testing/assert"
want = 4
testcase a {
	assert.equal(want: want, got: 2 + 2)
}
`,
			"b/b_test.flux": `package b_test
import "testing/assert"
testcase b extends "flux/a/a_test" {
	a_test.a()
	assert.equal(want: 6, got: 3 + 3)
}
`,
		},
	}

	ctx := filesystem.Inject(context.Background(), fs)
	data, err := filesystem.ReadFile(ctx, "b/b_test.flux")
	if err != nil {
		t.Fatal(err)
	}
	pkg := parser.ParseSource(string(data))
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error parsing b/b_test.flux: %s", ast.GetError(pkg))
	}
	pkg.Files[0].Name = "b/b_test.flux"

	names, pkgs, err := edit.TestcaseTransform(ctx, pkg, edit.TestModules{
		"flux": fs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := []string{"b"}, names; !cmp.Equal(want, got) {
		t.Fatalf("unexpected testcase names -want/+got:\n%s", cmp.Diff(want, got))
	}

	files := make([]string, len(pkgs[0].Files))
	for i, file := range pkgs[0].Files {
		files[i] = ast.Format(file)
	}

	want := []string{
		`package main


import "testing/assert"

a_test = () => {
	want = 4
	a = () => {
		assert.equal(want: want, got: 2 + 2)

		return {}
	}

	return {want, a}
}()`,
		`package main


import "testing/assert"

a_test.a()
assert.equal(want: 6, got: 3 + 3)`,
	}
	if got := files; !cmp.Equal(want, got) {
		t.Fatalf("unexpected file contents -want/+got:\n%s", cmp.Diff(want, got))
	}
}

type memoryFilesystem struct {
	files map[string]string
}

func (i *memoryFilesystem) Open(fpath string) (filesystem.File, error) {
	file, ok := i.files[filepath.Clean(fpath)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &memoryFile{
		name: fpath,
		r:    strings.NewReader(file),
	}, nil
}

type memoryFile struct {
	name string
	r    *strings.Reader
}

func (i *memoryFile) Read(p []byte) (n int, err error) {
	return i.r.Read(p)
}

func (i *memoryFile) Close() error {
	return nil
}

func (i *memoryFile) Stat() (os.FileInfo, error) {
	return i, nil
}

func (i *memoryFile) Name() string {
	return i.name
}

func (i *memoryFile) Size() int64 {
	return i.r.Size()
}

func (i *memoryFile) Mode() os.FileMode {
	return os.FileMode(0777)
}

func (i *memoryFile) ModTime() time.Time {
	return time.Time{}
}

func (i *memoryFile) IsDir() bool {
	return false
}

func (i *memoryFile) Sys() interface{} {
	return nil
}
