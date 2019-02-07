package edit_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
	"github.com/influxdata/flux/ast/edit"
)

func TestMatch(t *testing.T) {
	tcs := []struct {
		name   string
		in     ast.Node
		match  ast.Node
		result []ast.Node
		fuzzy  bool
	}{
		{
			name: "id",
			in: &ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ast.AdditionOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "b"},
						Operator: ast.MultiplicationOperator,
						Right: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "c"},
							Operator: ast.SubtractionOperator,
							Right:    &ast.Identifier{Name: "a"},
						},
					},
				},
			},
			match: &ast.Identifier{Name: "a"},
			result: []ast.Node{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "a"},
			},
		},
		{
			name: "empty id",
			in: &ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ast.AdditionOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "b"},
						Operator: ast.MultiplicationOperator,
						Right: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "c"},
							Operator: ast.SubtractionOperator,
							Right:    &ast.Identifier{Name: "a"},
						},
					},
				},
			},
			match: &ast.Identifier{},
			result: []ast.Node{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
				&ast.Identifier{Name: "c"},
				&ast.Identifier{Name: "a"},
			},
		},
		{
			name: "binary expression, ignore operator",
			in: &ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ast.AdditionOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "b"},
						Operator: ast.MultiplicationOperator,
						Right: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "c"},
							Operator: ast.SubtractionOperator,
							Right:    &ast.Identifier{Name: "a"},
						},
					},
				},
			},
			match: &ast.BinaryExpression{
				Operator: -1,
				Right:    &ast.Identifier{Name: "a"},
			},
			result: []ast.Node{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "c"},
					Operator: ast.SubtractionOperator,
					Right:    &ast.Identifier{Name: "a"},
				},
			},
		},
		{
			name: "recursive binary expression",
			in: &ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ast.AdditionOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "b"},
						Operator: ast.MultiplicationOperator,
						Right: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "c"},
							Operator: ast.SubtractionOperator,
							Right:    &ast.Identifier{Name: "a"},
						},
					},
				},
			},
			match: &ast.BinaryExpression{
				Operator: -1,
			},
			result: []ast.Node{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ast.AdditionOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "b"},
						Operator: ast.MultiplicationOperator,
						Right: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "c"},
							Operator: ast.SubtractionOperator,
							Right:    &ast.Identifier{Name: "a"},
						},
					},
				},
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "b"},
					Operator: ast.MultiplicationOperator,
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "c"},
						Operator: ast.SubtractionOperator,
						Right:    &ast.Identifier{Name: "a"},
					},
				},
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "c"},
					Operator: ast.SubtractionOperator,
					Right:    &ast.Identifier{Name: "a"},
				},
			},
		},
		{
			name: "property",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.ObjectExpression{
							Properties: []*ast.Property{
								{
									Key: &ast.StringLiteral{
										Value: "a",
									},
									Value: &ast.IntegerLiteral{
										Value: 10,
									},
								},
								{
									Key: &ast.Identifier{
										Name: "b",
									},
									Value: &ast.IntegerLiteral{
										Value: 11,
									},
								},
							},
						},
					},
				},
			}}},
			match: &ast.Property{
				Key: &ast.Identifier{
					Name: "b",
				},
			},
			result: []ast.Node{
				&ast.Property{
					Key: &ast.Identifier{
						Name: "b",
					},
					Value: &ast.IntegerLiteral{
						Value: 11,
					},
				},
			},
		},
		{
			name: "double match property",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.ObjectExpression{
							Properties: []*ast.Property{
								{
									Key: &ast.StringLiteral{
										Value: "a",
									},
									Value: &ast.IntegerLiteral{
										Value: 10,
									},
								},
								{
									Key: &ast.Identifier{
										Name: "b",
									},
									Value: &ast.IntegerLiteral{
										Value: 11,
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.ObjectExpression{
							Properties: []*ast.Property{
								{
									Key: &ast.Identifier{
										Name: "b",
									},
									Value: &ast.IntegerLiteral{
										Value: 12,
									},
								},
							},
						},
					},
				},
			}}},
			match: &ast.Property{
				Key: &ast.Identifier{
					Name: "b",
				},
			},
			result: []ast.Node{
				&ast.Property{
					Key: &ast.Identifier{
						Name: "b",
					},
					Value: &ast.IntegerLiteral{
						Value: 11,
					},
				},
				&ast.Property{
					Key: &ast.Identifier{
						Name: "b",
					},
					Value: &ast.IntegerLiteral{
						Value: 12,
					},
				},
			},
		},
		{
			name: "expression statement",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.ObjectExpression{
							Properties: []*ast.Property{
								{
									Key: &ast.Identifier{
										Name: "b",
									},
									Value: &ast.IntegerLiteral{
										Value: 12,
									},
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "a"},
									Property: &ast.Identifier{Name: "b"},
								},
							},
							Property: &ast.Identifier{Name: "c"},
						},
					},
				},
			}}},
			match: &ast.ExpressionStatement{},
			result: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.BinaryExpression{
						Operator: ast.AdditionOperator,
						Left:     &ast.IntegerLiteral{Value: 1},
						Right:    &ast.IntegerLiteral{Value: 2},
					},
				},
				&ast.ExpressionStatement{
					Expression: &ast.ObjectExpression{
						Properties: []*ast.Property{
							{
								Key: &ast.Identifier{
									Name: "b",
								},
								Value: &ast.IntegerLiteral{
									Value: 12,
								},
							},
						},
					},
				},
				&ast.ExpressionStatement{
					Expression: &ast.MemberExpression{
						Object: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "a"},
								Property: &ast.Identifier{Name: "b"},
							},
						},
						Property: &ast.Identifier{Name: "c"},
					},
				},
			},
		},
		{
			name:  "file fuzzy",
			fuzzy: true,
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.VariableAssignment{
						ID: &ast.Identifier{Name: "f"},
						Init: &ast.FunctionExpression{
							Params: []*ast.Property{
								{Key: &ast.Identifier{Name: "a"}},
								{Key: &ast.Identifier{Name: "b"}},
							},
							Body: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.Identifier{Name: "b"},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "f"},
							Arguments: []ast.Expression{&ast.ObjectExpression{
								Properties: []*ast.Property{
									{Key: &ast.Identifier{Name: "a"}, Value: &ast.IntegerLiteral{Value: 2}},
									{Key: &ast.Identifier{Name: "b"}, Value: &ast.IntegerLiteral{Value: 3}},
								},
							}},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Arguments: []ast.Expression{&ast.ObjectExpression{
								Properties: []*ast.Property{
									{Value: &ast.IntegerLiteral{Value: 3}},
								},
							}},
						},
					},
				},
			},
			result: []ast.Node{
				&ast.File{
					Body: []ast.Statement{
						&ast.VariableAssignment{
							ID: &ast.Identifier{Name: "f"},
							Init: &ast.FunctionExpression{
								Params: []*ast.Property{
									{Key: &ast.Identifier{Name: "a"}},
									{Key: &ast.Identifier{Name: "b"}},
								},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "a"},
									Right:    &ast.Identifier{Name: "b"},
								},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "f"},
								Arguments: []ast.Expression{&ast.ObjectExpression{
									Properties: []*ast.Property{
										{Key: &ast.Identifier{Name: "a"}, Value: &ast.IntegerLiteral{Value: 2}},
										{Key: &ast.Identifier{Name: "b"}, Value: &ast.IntegerLiteral{Value: 3}},
									},
								}},
							},
						},
					},
				},
			},
		},
		{
			name:  "file no match found",
			fuzzy: false,
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.VariableAssignment{
						ID: &ast.Identifier{Name: "f"},
						Init: &ast.FunctionExpression{
							Params: []*ast.Property{
								{Key: &ast.Identifier{Name: "a"}},
								{Key: &ast.Identifier{Name: "b"}},
							},
							Body: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.Identifier{Name: "b"},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "f"},
							Arguments: []ast.Expression{&ast.ObjectExpression{
								Properties: []*ast.Property{
									{Key: &ast.Identifier{Name: "a"}, Value: &ast.IntegerLiteral{Value: 2}},
									{Key: &ast.Identifier{Name: "b"}, Value: &ast.IntegerLiteral{Value: 3}},
								},
							}},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Arguments: []ast.Expression{&ast.ObjectExpression{
								Properties: []*ast.Property{
									{Value: &ast.IntegerLiteral{Value: 3}},
								},
							}},
						},
					},
				},
			},
			result: []ast.Node{},
		},
		{
			name:  "fuzzy slice",
			fuzzy: true,
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 4},
							Right:    &ast.IntegerLiteral{Value: 4},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
				},
			},
			result: []ast.Node{
				&ast.File{
					Body: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 1},
								Right:    &ast.IntegerLiteral{Value: 1},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 2},
								Right:    &ast.IntegerLiteral{Value: 2},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 3},
								Right:    &ast.IntegerLiteral{Value: 3},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 4},
								Right:    &ast.IntegerLiteral{Value: 4},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 5},
								Right:    &ast.IntegerLiteral{Value: 5},
							},
						},
					},
				},
			},
		},
		{
			name:  "fuzzy slice repeated no match",
			fuzzy: true,
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 4},
							Right:    &ast.IntegerLiteral{Value: 4},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
				},
			},
			result: []ast.Node{},
		},
		{
			name:  "fuzzy empty slice matches everything",
			fuzzy: true,
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{},
			},
			result: []ast.Node{
				&ast.File{
					Body: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 1},
								Right:    &ast.IntegerLiteral{Value: 1},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 2},
								Right:    &ast.IntegerLiteral{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			name: "exact slice",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 4},
							Right:    &ast.IntegerLiteral{Value: 4},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					nil,
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					nil,
					nil,
				},
			},
			result: []ast.Node{
				&ast.File{
					Body: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 1},
								Right:    &ast.IntegerLiteral{Value: 1},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 2},
								Right:    &ast.IntegerLiteral{Value: 2},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 3},
								Right:    &ast.IntegerLiteral{Value: 3},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 4},
								Right:    &ast.IntegerLiteral{Value: 4},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.IntegerLiteral{Value: 5},
								Right:    &ast.IntegerLiteral{Value: 5},
							},
						},
					},
				},
			},
		},
		{
			name: "exact slice no match wrong position",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 4},
							Right:    &ast.IntegerLiteral{Value: 4},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					nil,
					nil,
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					nil,
				},
			},
			result: []ast.Node{},
		},
		{
			name: "exact slice no match wrong len",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 1},
							Right:    &ast.IntegerLiteral{Value: 1},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 4},
							Right:    &ast.IntegerLiteral{Value: 4},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 5},
							Right:    &ast.IntegerLiteral{Value: 5},
						},
					},
				},
			}}},
			match: &ast.File{
				Body: []ast.Statement{
					nil,
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 2},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &ast.IntegerLiteral{Value: 3},
							Right:    &ast.IntegerLiteral{Value: 3},
						},
					},
					nil,
				},
			},
			result: []ast.Node{},
		},
		{
			name: "exact empty slice matches empty slice",
			in: &ast.Package{Files: []*ast.File{{
				Body: []ast.Statement{}}}},
			match: &ast.File{
				Body: []ast.Statement{},
			},
			result: []ast.Node{
				&ast.File{
					Body: []ast.Statement{},
				},
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ms := edit.Match(tc.in, tc.match, tc.fuzzy)
			if !cmp.Equal(tc.result, ms, asttest.IgnoreBaseNodeOptions...) {
				t.Errorf("unexpected match result: -want/+got:\n%s", cmp.Diff(tc.result, ms, asttest.IgnoreBaseNodeOptions...))
			}
		})
	}
}
