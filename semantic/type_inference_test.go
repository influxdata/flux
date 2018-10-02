package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
)

func TestInferTypes(t *testing.T) {
	testCases := []struct {
		name    string
		program *semantic.Program
		want    semantic.SolutionMap
	}{
		{
			name: "var assignment",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: true},
					},
				},
			},
		},
		{
			name: "var assignment with binary expression",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init: &semantic.BinaryExpression{
							Operator: ast.AdditionOperator,
							Left:     &semantic.IntegerLiteral{Value: 1},
							Right:    &semantic.IntegerLiteral{Value: 1},
						},
					},
				},
			},
		},
		{
			name: "var assignment with function",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{{
								Key: &semantic.Identifier{Name: "a"},
							}},
							Body: &semantic.FunctionBody{
								Argument: &semantic.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &semantic.IntegerLiteral{Value: 1},
									Right:    &semantic.IdentifierExpression{Name: "a"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			solution, err := semantic.InferTypes(tc.program)
			t.Log(solution)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

//func TestSolveVariableDeclaration(t *testing.T) {
//	program := &semantic.Program{
//		Body: []semantic.Statement{
//			&semantic.NativeVariableDeclaration{
//				Identifier: &semantic.Identifier{Name: "a"},
//				Init:       &semantic.BooleanLiteral{Value: true},
//			},
//		},
//	}
//}
//
//func TestSolveVariableReDeclaration(t *testing.T) {
//	program := &semantic.Program{
//		Body: []semantic.Statement{
//			&semantic.NativeVariableDeclaration{
//				Identifier: &semantic.Identifier{Name: "a"},
//				Init:       &semantic.BooleanLiteral{Value: true},
//			},
//			&semantic.NativeVariableDeclaration{
//				Identifier: &semantic.Identifier{Name: "a"},
//				Init:       &semantic.BooleanLiteral{Value: false},
//			},
//			&semantic.NativeVariableDeclaration{
//				Identifier: &semantic.Identifier{Name: "a"},
//				Init:       &semantic.BooleanLiteral{Value: false},
//			},
//		},
//	}
//}
//
//func TestSolveAdditionOperator(t *testing.T) {
//	program := &semantic.Program{
//		Body: []semantic.Statement{
//			&semantic.NativeVariableDeclaration{
//				Identifier: &semantic.Identifier{Name: "a"},
//				Init: &semantic.BinaryExpression{
//					Operator: ast.AdditionOperator,
//					Left:     &semantic.IntegerLiteral{Value: 2},
//					Right:    &semantic.IntegerLiteral{Value: 2},
//				},
//			},
//		},
//	}
//}
