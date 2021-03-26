package edit

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

// TestcaseTransform will transform an *ast.Package into a set of *ast.Package values
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
//     testcase addition_v2 extends "math_test" {
//         option math.enable_v2 = true
//         math_test.test_addition()
//     }
//
// This transforms the `math_test` file with the addition testcase into:
//
//     import "testing/assert"
//     math_test = () => {
//         myVar = 4
//         test_addition = () => {
//             assert.equal(want: 2 + 2, got: myVar)
//             return {}
//         }
//         return {myVar, test_addition}
//     }()
//
// The extended test file will be prepended to the list of files in the package as its own file.
//
// If a testcase extends another testcase, it will be replaced with the given body.
//
//     test_invalid_import = () => {
//         die(msg: "cannot extend an extended testcase")
//     }
//
// It is allowed for an imported testcase to have an option, but no attempt is made
// to remove duplicate options. If there is a duplicate option, this will likely
// cause an error when the test is actually run.
func TestcaseTransform(ctx context.Context, pkg *ast.Package, modules TestModules) ([]string, []*ast.Package, error) {
	if len(pkg.Files) != 1 {
		return nil, nil, errors.Newf(codes.FailedPrecondition, "unsupported number of files in test case package, got %d", len(pkg.Files))
	}
	file := pkg.Files[0]

	var (
		predicate []ast.Statement
		n         int
	)
	for _, item := range file.Body {
		if _, ok := item.(*ast.TestCaseStatement); ok {
			n++
			continue
		}
		predicate = append(predicate, item)
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

		testpkg, err := newTestPackage(ctx, pkg, predicate, testcase, modules)
		if err != nil {
			return nil, nil, err
		}
		names = append(names, testcase.ID.Name)
		pkgs = append(pkgs, testpkg)
	}

	return names, pkgs, nil
}

func newTestPackage(ctx context.Context, basePkg *ast.Package, predicate []ast.Statement, tc *ast.TestCaseStatement, modules TestModules) (*ast.Package, error) {
	pkg := basePkg.Copy().(*ast.Package)
	pkg.Package = "main"
	pkg.Files = nil

	file := basePkg.Files[0].Copy().(*ast.File)
	file.Package.Name.Name = "main"

	file.Body = make([]ast.Statement, 0, len(predicate)+len(tc.Block.Body))
	file.Body = append(file.Body, predicate...)
	if tc.Extends != nil {
		f, err := extendTest(file, tc.Extends.Value, modules)
		if err != nil {
			return nil, err
		}
		pkg.Files = append(pkg.Files, f)
	}
	file.Body = append(file.Body, tc.Block.Body...)
	pkg.Files = append(pkg.Files, file)
	return pkg, nil
}

func extendTest(file *ast.File, extends string, modules TestModules) (*ast.File, error) {
	components := strings.Split(extends, "/")
	if len(components) <= 1 {
		return nil, errors.New(codes.Invalid, "testcase extension requires a test module name and at least one other path component")
	}

	moduleName := components[0]
	module, ok := modules[moduleName]
	if !ok {
		return nil, errors.Newf(codes.FailedPrecondition, "test module %q not found", moduleName)
	}

	fpath := filepath.Join(components[1:]...) + ".flux"
	f, err := module.Open(fpath)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(f)
	_ = f.Close()
	if err != nil {
		return nil, err
	}

	pkgpath := filepath.Base(extends)
	impAst := parser.ParseSource(string(contents))
	if ast.Check(impAst) > 0 {
		return nil, ast.GetError(impAst)
	}

	// Construct the new file where we will place the imported package.
	newFile := impAst.Files[0].Copy().(*ast.File)
	newFile.Package = file.Package

	// We must construct a new body for the parsed ast.
	// The body must be placed into a function and then that function
	// must have a return argument.
	// Any testcase statements must be transformed similarly into a function.
	var (
		body    = impAst.Files[0].Body
		newBody = make([]ast.Statement, 0, len(body))
		vars    []*ast.Property
	)
	for _, stmt := range body {
		switch stmt := stmt.(type) {
		case *ast.OptionStatement:
			// Extract into the top level of the file.
			newFile.Body = append(newFile.Body, stmt)
		case *ast.TestCaseStatement:
			body := stmt.Block.Body
			if valid, reason := isValidTestcase(stmt); !valid {
				body = []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "die"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{{
										Key:   &ast.Identifier{Name: "msg"},
										Value: ast.StringLiteralFromValue(reason),
									}},
								},
							},
						},
					},
				}
			}
			body = append(body, &ast.ReturnStatement{
				Argument: &ast.ObjectExpression{},
			})
			varname := stmt.ID
			vars = append(vars, &ast.Property{Key: varname})
			newBody = append(newBody, &ast.VariableAssignment{
				ID: varname,
				Init: &ast.FunctionExpression{
					Body: &ast.Block{
						Body: body,
					},
				},
			})
		case *ast.VariableAssignment:
			vars = append(vars, &ast.Property{Key: stmt.ID})
			newBody = append(newBody, stmt)
		default:
			newBody = append(newBody, stmt)
		}
	}

	// Add a final statement to the body with a return statement
	// holding the variables.
	newBody = append(newBody, &ast.ReturnStatement{
		Argument: &ast.ObjectExpression{
			Properties: vars,
		},
	})

	// Assign the new file body to the single variable assignment.
	newFile.Body = []ast.Statement{
		&ast.VariableAssignment{
			ID: &ast.Identifier{Name: pkgpath},
			Init: &ast.CallExpression{
				Callee: &ast.FunctionExpression{
					Body: &ast.Block{
						Body: newBody,
					},
				},
			},
		},
	}
	return newFile, nil
}

func isValidTestcase(tc *ast.TestCaseStatement) (valid bool, reason string) {
	if tc.Extends != nil {
		return false, "cannot extend an extended testcase"
	}

	for _, stmt := range tc.Block.Body {
		if _, ok := stmt.(*ast.OptionStatement); ok {
			return false, "option statements may not be used in an extended testcase"
		}
	}
	return true, ""
}

type TestModules map[string]filesystem.Service

func (m *TestModules) Add(name string, fs filesystem.Service) error {
	if *m == nil {
		*m = make(map[string]filesystem.Service)
	}

	if _, ok := (*m)[name]; ok {
		return errors.Newf(codes.FailedPrecondition, "duplicate test module %q", name)
	}
	(*m)[name] = fs
	return nil
}

func (m *TestModules) Merge(other TestModules) error {
	for name, fs := range other {
		if err := m.Add(name, fs); err != nil {
			return err
		}
	}
	return nil
}
