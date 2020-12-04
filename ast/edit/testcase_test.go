package edit_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/ast/edit"
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
	expected[0].Files[0].Name = "test_addition"
	expected[1].Files[0].Name = "test_subtraction"

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

	transformed, err := edit.TestcaseTransform(d)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(transformed, expected, asttest.IgnoreBaseNodeOptions...) {
		t.Errorf("unexpected match result: -want/+got:\n%s", cmp.Diff(transformed, expected, asttest.IgnoreBaseNodeOptions...))
	}
}

func TestTestcaseTransformNoTestcase(t *testing.T) {
	testFile := `package an_test

import "testing"

myVar = 4`

	d := parser.ParseSource(testFile)

	transformed, err := edit.TestcaseTransform(d)
	if err != nil {
		t.Fatal(err)
	}

	if len(transformed) != 0 {
		t.Errorf("unexpected package count: want: 0, got: %d", len(transformed))
	}
}
