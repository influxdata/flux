package testcase_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/InfluxCommunity/flux/ast"
	"github.com/InfluxCommunity/flux/ast/asttest"
	"github.com/InfluxCommunity/flux/ast/astutil"
	"github.com/InfluxCommunity/flux/ast/testcase"
	"github.com/InfluxCommunity/flux/dependencies/filesystem"
	"github.com/InfluxCommunity/flux/parser"
	"github.com/google/go-cmp/cmp"
)

func TestTransform(t *testing.T) {
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

	idens, transformed, err := testcase.Transform(context.Background(), d, nil)
	if err != nil {
		t.Fatal(err)
	}
	testNames := make([]string, len(idens))
	for i := range idens {
		testNames[i] = idens[i].Name
	}

	if want, got := []string{"test_addition", "test_subtraction"}, testNames; !cmp.Equal(want, got) {
		t.Errorf("unexpected test names: -want/+got:\n%s", cmp.Diff(want, got))
	}
	if !cmp.Equal(expected, transformed, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("unexpected match result: -want/+got:\n%s", cmp.Diff(expected, transformed, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestTransformNoTestcase(t *testing.T) {
	testFile := `package an_test

import "testing"

myVar = 4`

	d := parser.ParseSource(testFile)

	_, transformed, err := testcase.Transform(context.Background(), d, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(transformed) != 0 {
		t.Errorf("unexpected package count: want: 0, got: %d", len(transformed))
	}
}

func TestTransformImport(t *testing.T) {
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
testcase b extends "flux/a/a_test.a" {
	super()
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

	idens, pkgs, err := testcase.Transform(ctx, pkg, testcase.TestModules{
		"flux": testcase.TestModule{
			Service: fs,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	testNames := make([]string, len(idens))
	for i := range idens {
		testNames[i] = idens[i].Name
	}
	if want, got := []string{"b"}, testNames; !cmp.Equal(want, got) {
		t.Fatalf("unexpected testcase names -want/+got:\n%s", cmp.Diff(want, got))
	}

	files := make([]string, len(pkgs[0].Files))
	for i, file := range pkgs[0].Files {
		fileStr, err := astutil.Format(file)
		if err != nil {
			t.Fatalf("unexpected error from formatter: %s", err)
		}
		files[i] = fileStr
	}

	want := []string{
		`package main


import "testing/assert"

want = 4

assert.equal(want: want, got: 2 + 2)
assert.equal(want: 6, got: 3 + 3)
`,
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
