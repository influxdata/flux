package ast_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
)

func TestCopy(t *testing.T) {
	testCases := []struct {
		node ast.Node
	}{
		{
			node: &ast.Package{},
		},
		{
			node: &ast.Package{
				Package: "foo",
				Files: []*ast.File{{
					Name: "a.flux",
				}},
			},
		},
		{
			node: &ast.File{},
		},
		{
			node: &ast.File{
				Name: "a.flux",
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: "a"},
				},
				Imports: []*ast.ImportDeclaration{{
					As:   &ast.Identifier{Name: "a"},
					Path: &ast.StringLiteral{Value: "a/b"},
				}},
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BooleanLiteral{Value: true},
					},
				},
			},
		},
		{
			node: &ast.PackageClause{},
		},
		{
			node: &ast.PackageClause{
				Name: &ast.Identifier{Name: "a"},
			},
		},
		{
			node: &ast.ImportDeclaration{},
		},
		{
			node: &ast.ImportDeclaration{
				As:   &ast.Identifier{Name: "a"},
				Path: &ast.StringLiteral{Value: "a/b"},
			},
		},
		{
			node: &ast.BadStatement{},
		},
		{
			node: &ast.BadStatement{
				Text: "foo",
			},
		},
		{
			node: &ast.Block{},
		},
		{
			node: &ast.Block{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BooleanLiteral{Value: true},
					},
				},
			},
		},
		{
			node: &ast.ExpressionStatement{},
		},
		{
			node: &ast.ExpressionStatement{
				Expression: &ast.BooleanLiteral{Value: true},
			},
		},
		{
			node: &ast.ReturnStatement{},
		},
		{
			node: &ast.ReturnStatement{
				Argument: &ast.BooleanLiteral{Value: true},
			},
		},
		{
			node: &ast.OptionStatement{},
		},
		{
			node: &ast.OptionStatement{
				Assignment: &ast.VariableAssignment{
					ID:   &ast.Identifier{Name: "a"},
					Init: &ast.BooleanLiteral{Value: true},
				},
			},
		},
		{
			node: &ast.OptionStatement{
				Assignment: &ast.MemberAssignment{
					Member: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "alert",
						},
						Property: &ast.Identifier{
							Name: "state",
						},
					},
					Init: &ast.StringLiteral{
						Value: "Warning",
					},
				},
			},
		},
		{
			node: &ast.TestStatement{},
		},
		{
			node: &ast.TestStatement{
				Assignment: &ast.VariableAssignment{
					ID: &ast.Identifier{
						Name: "mean",
					},
					Init: &ast.ObjectExpression{
						Properties: []*ast.Property{
							{
								Key: &ast.Identifier{
									Name: "want",
								},
								Value: &ast.IntegerLiteral{
									Value: 0,
								},
							},
							{
								Key: &ast.Identifier{
									Name: "got",
								},
								Value: &ast.IntegerLiteral{
									Value: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			node: &ast.VariableAssignment{},
		},
		{
			node: &ast.VariableAssignment{
				ID:   &ast.Identifier{Name: "a"},
				Init: &ast.BooleanLiteral{Value: true},
			},
		},
		{
			node: &ast.CallExpression{},
		},
		{
			node: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "a"},
				Arguments: []ast.Expression{
					&ast.BooleanLiteral{Value: true},
				},
			},
		},
		{
			node: &ast.PipeExpression{},
		},
		{
			node: &ast.PipeExpression{
				Argument: &ast.BooleanLiteral{Value: true},
				Call: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "a"},
					Arguments: []ast.Expression{
						&ast.BooleanLiteral{Value: true},
					},
				},
			},
		},
		{
			node: &ast.MemberExpression{},
		},
		{
			node: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "o"},
				Property: &ast.StringLiteral{Value: "a"},
			},
		},
		{
			node: &ast.IndexExpression{},
		},
		{
			node: &ast.IndexExpression{
				Array: &ast.Identifier{Name: "o"},
				Index: &ast.IntegerLiteral{Value: 3},
			},
		},
		{
			node: &ast.FunctionExpression{},
		},
		{
			node: &ast.FunctionExpression{
				Params: []*ast.Property{{
					Key:   &ast.Identifier{Name: "a"},
					Value: &ast.IntegerLiteral{Value: 3},
				}},
				Body: &ast.Identifier{Name: "a"},
			},
		},
		{
			node: &ast.BinaryExpression{},
		},
		{
			node: &ast.BinaryExpression{
				Operator: ast.AdditionOperator,
				Left:     &ast.IntegerLiteral{Value: 3},
				Right:    &ast.IntegerLiteral{Value: 3},
			},
		},
		{
			node: &ast.UnaryExpression{},
		},
		{
			node: &ast.UnaryExpression{
				Operator: ast.AdditionOperator,
				Argument: &ast.IntegerLiteral{Value: 3},
			},
		},
		{
			node: &ast.LogicalExpression{},
		},
		{
			node: &ast.LogicalExpression{
				Operator: ast.AndOperator,
				Left:     &ast.BooleanLiteral{Value: false},
				Right:    &ast.BooleanLiteral{Value: true},
			},
		},
		{
			node: &ast.ArrayExpression{},
		},
		{
			node: &ast.ArrayExpression{
				Elements: []ast.Expression{&ast.BooleanLiteral{Value: false}},
			},
		},
		{
			node: &ast.ObjectExpression{},
		},
		{
			node: &ast.ObjectExpression{
				Properties: []*ast.Property{{
					Key:   &ast.Identifier{Name: "a"},
					Value: &ast.IntegerLiteral{Value: 3},
				}},
			},
		},
		{
			node: &ast.ConditionalExpression{},
		},
		{
			node: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "a"},
				Alternate:  &ast.Identifier{Name: "b"},
				Consequent: &ast.Identifier{Name: "c"},
			},
		},
		{
			node: &ast.Property{},
		},
		{
			node: &ast.Property{
				Key:   &ast.Identifier{Name: "a"},
				Value: &ast.IntegerLiteral{Value: 3},
			},
		},
		{
			node: &ast.Identifier{},
		},
		{
			node: &ast.Identifier{
				Name: "f",
			},
		},
		{
			node: &ast.PipeLiteral{},
		},
		{
			node: &ast.StringLiteral{},
		},
		{
			node: &ast.StringLiteral{
				Value: "a",
			},
		},
		{
			node: &ast.BooleanLiteral{},
		},
		{
			node: &ast.BooleanLiteral{
				Value: true,
			},
		},
		{
			node: &ast.FloatLiteral{},
		},
		{
			node: &ast.FloatLiteral{
				Value: 1,
			},
		},
		{
			node: &ast.IntegerLiteral{},
		},
		{
			node: &ast.IntegerLiteral{
				Value: 1,
			},
		},
		{
			node: &ast.UnsignedIntegerLiteral{},
		},
		{
			node: &ast.UnsignedIntegerLiteral{
				Value: 1,
			},
		},
		{
			node: &ast.RegexpLiteral{},
		},
		{
			node: &ast.RegexpLiteral{
				Value: regexp.MustCompile(".*"),
			},
		},
		{
			node: &ast.DurationLiteral{},
		},
		{
			node: &ast.DurationLiteral{
				Values: []ast.Duration{{
					Magnitude: 1,
					Unit:      "s",
				}},
			},
		},
		{
			node: &ast.DateTimeLiteral{},
		},
		{
			node: &ast.DateTimeLiteral{
				Value: time.Now(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%T", tc.node), func(t *testing.T) {
			cpy := tc.node.Copy()
			if !cmp.Equal(cpy, tc.node, asttest.CmpOptions...) {
				t.Errorf("copy not equal -want/+got:\n%s", cmp.Diff(tc.node, cpy, asttest.CmpOptions...))
			}
		})
	}
}
