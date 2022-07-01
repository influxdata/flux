package ast_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
)

func TestWalk(t *testing.T) {
	testCases := []struct {
		name  string
		node  ast.Node
		order func(node ast.Node) []ast.Node
	}{
		{
			name: "package",
			node: &ast.File{
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: "foo"},
				},
			},
			order: func(node ast.Node) []ast.Node {
				file := node.(*ast.File)
				return []ast.Node{
					file,
					file.Package,
					file.Package.Name,
				}
			},
		},
		{
			name: "imports",
			node: &ast.File{
				Imports: []*ast.ImportDeclaration{
					{
						As:   &ast.Identifier{Name: "foo"},
						Path: &ast.StringLiteral{Value: "path/foo"},
					},
					{
						Path: &ast.StringLiteral{Value: "path/bar"},
					},
				},
			},
			order: func(node ast.Node) []ast.Node {
				file := node.(*ast.File)
				return []ast.Node{
					file,
					file.Imports[0],
					file.Imports[0].As,
					file.Imports[0].Path,
					file.Imports[1],
					file.Imports[1].Path,
				}
			},
		},
		{
			name: "body",
			node: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "foo"},
					},
				},
			},
			order: func(node ast.Node) []ast.Node {
				file := node.(*ast.File)
				return []ast.Node{
					file,
					file.Body[0],
					file.Body[0].(*ast.ExpressionStatement).Expression,
				}
			},
		},
		{
			name: "package imports body",
			node: &ast.File{
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: "foo"},
				},
				Imports: []*ast.ImportDeclaration{
					{
						As:   &ast.Identifier{Name: "foo"},
						Path: &ast.StringLiteral{Value: "path/foo"},
					},
					{
						Path: &ast.StringLiteral{Value: "path/bar"},
					},
				},
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "foo"},
					},
				},
			},
			order: func(node ast.Node) []ast.Node {
				file := node.(*ast.File)
				return []ast.Node{
					file,
					file.Package,
					file.Package.Name,
					file.Imports[0],
					file.Imports[0].As,
					file.Imports[0].Path,
					file.Imports[1],
					file.Imports[1].Path,
					file.Body[0],
					file.Body[0].(*ast.ExpressionStatement).Expression,
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			want := tc.order(tc.node)
			v := new(orderVisitor)
			ast.Walk(v, tc.node)
			got := v.order
			if !cmp.Equal(want, got, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("unexpected walk order: -want/+got:\n%s", cmp.Diff(want, got, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}

}

// orderVisitor records the walk order
type orderVisitor struct {
	order []ast.Node
}

func (o *orderVisitor) Visit(node ast.Node) (w ast.Visitor) {
	o.order = append(o.order, node)
	return o
}
func (o *orderVisitor) Done(node ast.Node) {}
