package ast_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/parser"
	"github.com/pkg/errors"
)

func withEachFluxFile(t *testing.T, fn func(caseName, fileContent string)) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "testdata")

	fluxFiles, err := filepath.Glob(filepath.Join(path, "*.flux"))
	if err != nil {
		t.Fatalf("error searching for Flux files: %s", err)
	}

	for _, fluxFile := range fluxFiles {
		ext := filepath.Ext(fluxFile)
		prefix := fluxFile[0 : len(fluxFile)-len(ext)]
		_, caseName := filepath.Split(prefix)

		content, err := ioutil.ReadFile(fluxFile)
		if err != nil {
			t.Fatal(err)
		}

		fn(caseName, string(content))
	}
}

func TestFormat(t *testing.T) {
	// we compare the semantic (we also check that we got a valid query)
	withEachFluxFile(t, func(caseName, content string) {
		t.Run(caseName, func(t *testing.T) {
			originalProgram, err := parser.NewAST(content)
			if err != nil {
				t.Fatal(errors.Wrapf(err, "original program has bad syntax:\n%s", content))
			}

			stringResult := ast.Format(originalProgram)

			newProgram, err := parser.NewAST(stringResult)
			if err != nil {
				t.Fatal(errors.Wrapf(err, "new program has bad syntax:\n%s", stringResult))
			}

			if !cmp.Equal(originalProgram, newProgram, asttest.CompareOptions...) {
				t.Errorf("to string conversion error:\nin:\t\t%s\nout:\t\t%s\n", content, stringResult)
			}
		})
	})
}
