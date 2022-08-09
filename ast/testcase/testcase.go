package testcase

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
)

// Transform will transform an *ast.Package into a set of *ast.Package values
// that represent each testcase defined within the original package.
//
// A testcase is defined with the testcase statement such as below.
//
//     import "testing/assert"
//     myVar = 4
//     testcase addition {
//         assert.equal(want: 2 + 2, got: myVar)
//     }
//
// This gets transformed into a package that looks like this:
//
//     import "testing/assert"
//     myVar = 4
//     assert.equal(want: 2 + 2, got: myVar)
//
// It is allowed to include options within the testcase block as they will be extracted
// to the top level.
//
// In addition to this syntax, testcase blocks may also extend another test file.
// This will transform the the extended testcase in a slightly different way.
// The syntax for extending is as such:
//
//     import "math"
//     testcase addition_v2 extends "math_test.addition" {
//         option math.enable_v2 = true
//         super()
//     }
//
// The extending test case is then transformed into a single file combining both the parent
// statements and the current statements.
//
//     import "testing/assert"
//     import "math"
//
//     option math.enable_v2 = true
//
//     myVar = 4
//     assert.equal(want: 2 + 2, got: myVar)
//
//
// The call to `super()` is replaced with the body of the parent test case.
//
// It is an error to extend an extended testcase.
//
// It is allowed for an imported testcase to have an option, but no attempt is made
// to remove duplicate options. If there is a duplicate option, this will likely
// cause an error when the test is actually run.
func Transform(ctx context.Context, pkg *ast.Package, modules TestModules) ([]string, []*ast.Package, error) {
	if len(pkg.Files) != 1 {
		return nil, nil, errors.Newf(codes.FailedPrecondition, "unsupported number of files in test case package, got %d", len(pkg.Files))
	}
	file := pkg.Files[0]

	var (
		preamble []ast.Statement
		n        int
	)
	for _, item := range file.Body {
		if _, ok := item.(*ast.TestCaseStatement); ok {
			n++
			continue
		}
		preamble = append(preamble, item)
	}

	var (
		names = make([]string, 0, n)
		pkgs  = make([]*ast.Package, 0, n)
	)
	for _, item := range file.Body {
		testcase, ok := item.(*ast.TestCaseStatement)
		if !ok {
			continue
		}

		testpkg, err := newTestPackage(ctx, pkg, preamble, testcase, modules)
		if err != nil {
			return nil, nil, err
		}
		names = append(names, testcase.ID.Name)
		pkgs = append(pkgs, testpkg)
	}

	return names, pkgs, nil
}

func newTestPackage(ctx context.Context, basePkg *ast.Package, preamble []ast.Statement, tc *ast.TestCaseStatement, modules TestModules) (*ast.Package, error) {
	pkg := basePkg.Copy().(*ast.Package)
	pkg.Package = "main"
	pkg.Files = nil

	file := basePkg.Files[0].Copy().(*ast.File)
	file.Package.Name.Name = "main"

	file.Body = make([]ast.Statement, 0, len(preamble)+len(tc.Block.Body))
	file.Body = append(file.Body, preamble...)
	if tc.Extends != nil {
		parentImports, parentPreamble, parentTC, err := extendTest(file, tc.Extends.Value, modules)
		if err != nil {
			return nil, err
		}
		file.Imports = mergeImports(file.Imports, parentImports)
		file.Body = append(file.Body, parentPreamble...)
		// Copy test case statements into body replacing the super statement
		// with the parent test case statements
		found := false
		for _, s := range tc.Block.Body {
			if !found {
				if es, ok := s.(*ast.ExpressionStatement); ok {
					if call, ok := es.Expression.(*ast.CallExpression); ok {
						if id, ok := call.Callee.(*ast.Identifier); ok && len(call.Arguments) == 0 {
							if id.Name == "super" {
								file.Body = append(file.Body, parentTC...)
								found = true
								continue
							}
						}
					}
				}
			}
			file.Body = append(file.Body, s)
		}
	} else {
		// Simply copy test case body into file
		file.Body = append(file.Body, tc.Block.Body...)
	}
	pkg.Files = append(pkg.Files, file)
	return pkg, nil
}

func extendTest(file *ast.File, extends string, modules TestModules) ([]*ast.ImportDeclaration, []ast.Statement, []ast.Statement, error) {
	testcaseName, f, err := modules.Open(extends)
	if err != nil {
		return nil, nil, nil, err
	}
	contents, err := ioutil.ReadAll(f)
	_ = f.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	impAst := parser.ParseSourceWithFileName(string(contents), file.Name)
	if ast.Check(impAst) > 0 {
		return nil, nil, nil, ast.GetError(impAst)
	}

	// Find the preamble statements and the test case statements
	var (
		preamble           = make([]ast.Statement, 0, len(impAst.Files[0].Body))
		testCaseStatements []ast.Statement
	)
	for _, item := range impAst.Files[0].Body {
		if tc, ok := item.(*ast.TestCaseStatement); ok {
			if tc.ID.Name == testcaseName {
				testCaseStatements = tc.Block.Body
			}
			continue
		}
		preamble = append(preamble, item)
	}

	return impAst.Files[0].Imports, preamble, testCaseStatements, nil
}

func mergeImports(a, b []*ast.ImportDeclaration) []*ast.ImportDeclaration {
	dst := make([]*ast.ImportDeclaration, len(a), len(a)+len(b))
	copy(dst, a)

B:
	for _, imp := range b {
		for _, existingImp := range a {
			if imp.Path.Value == existingImp.Path.Value {
				continue B
			}
		}
		dst = append(dst, imp)
	}
	return dst
}

type TestModules map[string]TestModule

// TestModule represents a single fluxtest root
type TestModule struct {
	filesystem.Service
	// Tags is the set of valid tags tests in this module may contain
	Tags []string
}

func (m *TestModules) Tags(name string) []string {
	if m, ok := (*m)[name]; ok {
		return m.Tags
	}
	return nil
}

// Open returns a File for the testcase path
//
// A test case path has the following format:
//
// <module_name>/<filesystem_path>/<testfile_path>.<testcase_name>
//
// where `filesystem_path` may contain multiple path elements.
func (m *TestModules) Open(fpath string) (string, filesystem.File, error) {
	components := strings.Split(fpath, "/")
	if len(components) <= 1 {
		return "", nil, errors.New(codes.Invalid, "testcase path must contain a module name")
	}

	moduleName := components[0]
	module, ok := (*m)[moduleName]
	if !ok {
		return "", nil, errors.Newf(codes.FailedPrecondition, "test module %q not found", moduleName)
	}

	// Extenstion is the test case name
	// Swap for '.flux' extension
	ext := filepath.Ext(fpath)
	name := strings.TrimPrefix(ext, ".")
	last := len(components) - 1
	components[last] = strings.TrimSuffix(components[last], ext)
	fp := filepath.Join(components[1:]...) + ".flux"

	f, err := module.Open(fp)
	return name, f, err
}

func (m *TestModules) Add(name string, mod TestModule) error {
	if *m == nil {
		*m = make(map[string]TestModule)
	}

	if _, ok := (*m)[name]; ok {
		return errors.Newf(codes.FailedPrecondition, "duplicate test module %q", name)
	}
	(*m)[name] = mod
	return nil
}

func (m *TestModules) Merge(other TestModules) error {
	for name, mod := range other {
		if err := m.Add(name, mod); err != nil {
			return err
		}
	}
	return nil
}
