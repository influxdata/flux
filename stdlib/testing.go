package stdlib

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	ast "github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/parser"
)

// TestingRunCalls constructs an ast.File that calls testing.run for each test case within the package.
func TestingRunCalls(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "run")
}

// TestingInspectCalls constructs an ast.File that calls testing.inspect for each test case within the package.
func TestingInspectCalls(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "inspect")
}

// TestingBenchmarkCalls constructs an ast.File that calls testing.benchmark for each test case within the package.
func TestingBenchmarkCalls(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "benchmark")
}

func genCalls(pkg *ast.Package, fn string) *ast.File {
	callFile := new(ast.File)
	callFile.Imports = []*ast.ImportDeclaration{{
		Path: &ast.StringLiteral{Value: "testing"},
	}}
	visitor := testStmtVisitor{
		fn: func(tc *ast.TestStatement) {
			callFile.Body = append(callFile.Body, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "testing"},
						Property: &ast.StringLiteral{Value: fn},
					},
					Arguments: []ast.Expression{
						&ast.ObjectExpression{
							Properties: []*ast.Property{{
								Key:   &ast.Identifier{Name: "case"},
								Value: tc.Assignment.ID,
							}},
						},
					},
				},
			})
		},
	}
	ast.Walk(visitor, pkg)
	return callFile
}

type testStmtVisitor struct {
	fn func(*ast.TestStatement)
}

func (v testStmtVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TestStatement:
		v.fn(n)
		return nil
	}
	return v
}

func (v testStmtVisitor) Done(node ast.Node) {}

/// Scans `rootDir` for all packages that contain `testcase` statements and returns them
func FindTestPackages(rootDir string) ([]*ast.Package, error) {
	var testPackages []*ast.Package
	pkgName := "github.com/influxdata/flux/stdlib"
	err := walkDirs(rootDir, func(dir string) error {
		// Determine the absolute flux package path
		fluxPath, err := filepath.Rel(rootDir, dir)
		if err != nil {
			return err
		}

		fset := new(token.FileSet)
		pkgs, err := parser.ParseDir(fset, dir)
		if err != nil {
			return err
		}

		// Annotate the packages with the absolute flux package path.
		for _, pkg := range pkgs {
			pkg.Path = fluxPath
		}

		var testPkg *ast.Package
		switch len(pkgs) {
		case 0:
			return nil
		case 1:
			for k, v := range pkgs {
				if strings.HasSuffix(k, "_test") {
					testPkg = v
				}
			}
		case 2:
			for k, v := range pkgs {
				if strings.HasSuffix(k, "_test") {
					testPkg = v
					continue
				}
			}
			if testPkg == nil {
				return fmt.Errorf("cannot have two distinct non-test Flux packages in the same directory")
			}
		default:
			keys := make([]string, 0, len(pkgs))
			for k := range pkgs {
				keys = append(keys, k)
			}
			return fmt.Errorf("found more than 2 flux packages in directory %s; packages %v", dir, keys)
		}

		if testPkg != nil {
			// Strip out test files with the testcase statement.
			validFiles := []*ast.File{}
			for _, file := range testPkg.Files {
				valid := true
				for _, item := range file.Body {
					if _, ok := item.(*ast.TestCaseStatement); ok {
						valid = false
					}
				}
				if valid {
					validFiles = append(validFiles, file)
				}
			}
			if len(validFiles) < len(testPkg.Files) {
				testPkg.Files = validFiles
			}

			if ast.Check(testPkg) > 0 {
				return errors.Wrapf(ast.GetError(testPkg), codes.Inherit, "failed to parse test package %q", testPkg.Package)
			}
			// Validate test package file use _test.flux suffix for the file name
			for _, f := range testPkg.Files {
				if !strings.HasSuffix(f.Name, "_test.flux") {
					return fmt.Errorf("flux test files must use the _test.flux suffix in their file name, found %q", path.Join(dir, f.Name))
				}
			}
			// Track go import path
			importPath := path.Join(pkgName, dir)
			if importPath != pkgName {
				testPackages = append(testPackages, testPkg)
			}

		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return testPackages, nil
}

func walkDirs(path string, f func(dir string) error) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	if err := f(path); err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			if err := walkDirs(filepath.Join(path, file.Name()), f); err != nil {
				return err
			}
		}
	}
	return nil
}
