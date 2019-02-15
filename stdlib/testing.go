package stdlib

import ast "github.com/influxdata/flux/ast"

// TestingRunCalls constructs an ast.File that calls testing.run for each test case within the package.
func TestingRunCalls(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "run")
}

// TestingInspectCalls constructs an ast.File that calls testing.inspect for each test case within the package.
func TestingInspectCalls(pkg *ast.Package) *ast.File {
	return genCalls(pkg, "inspect")
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
