package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/semantic"
)

func TestW(t *testing.T) {
	testCases := []struct {
		name string
		node semantic.Node
		want semantic.Type
	}{
		{
			name: "bool",
			node: &semantic.BooleanLiteral{Value: false},
			want: semantic.Bool,
		},
		{
			name: "bool decl",
			node: &semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "b"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
			want: semantic.Bool,
		},
		{
			name: "identity",
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "identity"},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Parameters: &semantic.FunctionParameters{
									List: []*semantic.FunctionParameter{{
										Key: &semantic.Identifier{Name: "x"},
									}},
								},
								Body: &semantic.IdentifierExpression{Name: "x"},
							},
						},
					},
					&semantic.ExpressionStatement{
						Expression: &semantic.CallExpression{
							Callee: &semantic.IdentifierExpression{Name: "identity"},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{{
									Key:   &semantic.Identifier{Name: "x"},
									Value: &semantic.IntegerLiteral{Value: 1},
								}},
							},
						},
					},
				},
			},
			want: semantic.Int,
		},
		{
			name: "identity2",
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "identity"},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Parameters: &semantic.FunctionParameters{
									List: []*semantic.FunctionParameter{{
										Key: &semantic.Identifier{Name: "x"},
									}},
								},
								Body: &semantic.IdentifierExpression{Name: "x"},
							},
						},
					},
					&semantic.ExpressionStatement{
						Expression: &semantic.CallExpression{
							Callee: &semantic.CallExpression{
								Callee: &semantic.IdentifierExpression{Name: "identity"},
								Arguments: &semantic.ObjectExpression{
									Properties: []*semantic.Property{{
										Key:   &semantic.Identifier{Name: "x"},
										Value: &semantic.IdentifierExpression{Name: "identity"},
									}},
								},
							},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{{
									Key:   &semantic.Identifier{Name: "x"},
									Value: &semantic.StringLiteral{Value: "2"},
								}},
							},
						},
					},
				},
			},
			want: semantic.String,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := semantic.Infer(tc.node)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("unexpected types want: %v got %v", tc.want, got)
			}
		})
	}
}
