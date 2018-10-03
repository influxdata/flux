package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
)

type SolutionVisitor interface {
	semantic.Visitor
	Solution() semantic.SolutionMap
}

type visitor struct {
	f        func(node semantic.Node) semantic.Type
	solution semantic.SolutionMap
}

func (v *visitor) Visit(node semantic.Node) semantic.Visitor {
	if v.solution == nil {
		v.solution = make(semantic.SolutionMap)
	}
	// Handle literals here
	if l, ok := node.(semantic.Literal); ok {
		var typ semantic.Type
		switch l.(type) {
		case *semantic.BooleanLiteral:
			typ = semantic.Bool
		case *semantic.IntegerLiteral:
			typ = semantic.Int
		}
		v.solution[node] = typ
		return v
	}

	typ := v.f(node)
	if typ != nil {
		v.solution[node] = typ
	}
	return v
}

func (v *visitor) Done() {}

func (v *visitor) Solution() semantic.SolutionMap {
	return v.solution
}

func TestInferTypes(t *testing.T) {
	testCases := []struct {
		name     string
		program  *semantic.Program
		solution SolutionVisitor
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
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.NativeVariableDeclaration:
						return semantic.Bool
					}
					return nil
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
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.NativeVariableDeclaration,
						*semantic.BinaryExpression:
						return semantic.Int
					}
					return nil
				},
			},
		},
		{
			name: "var assignment with function",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "f"},
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
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.FunctionBody,
						*semantic.FunctionParam,
						*semantic.IdentifierExpression:
						return semantic.Int
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return semantic.NewFunctionType(semantic.FunctionSignature{
							Params: map[string]semantic.Type{
								"a": semantic.Int,
							},
							ReturnType: semantic.Int,
						})
					}
					return nil
				},
			},
		},
		{
			name: "call function",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "two"},
						Init: &semantic.CallExpression{
							Callee: &semantic.FunctionExpression{
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
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{{
									Key:   &semantic.Identifier{Name: "a"},
									Value: &semantic.IntegerLiteral{Value: 1},
								}},
							},
						},
					},
				},
			},
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.NativeVariableDeclaration,
						*semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.FunctionBody,
						*semantic.FunctionParam,
						*semantic.IdentifierExpression:
						return semantic.Int
					case *semantic.FunctionExpression:
						return semantic.NewFunctionType(semantic.FunctionSignature{
							Params: map[string]semantic.Type{
								"a": semantic.Int,
							},
							ReturnType: semantic.Int,
						})
					case *semantic.ObjectExpression:
						return semantic.NewObjectType(map[string]semantic.Type{
							"a": semantic.Int,
						})
					}
					return nil
				},
			},
		},
		{
			name: "call function identifier",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "add"},
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
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "two"},
						Init: &semantic.CallExpression{
							Callee: &semantic.IdentifierExpression{Name: "add"},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{{
									Key:   &semantic.Identifier{Name: "a"},
									Value: &semantic.IntegerLiteral{Value: 1},
								}},
							},
						},
					},
				},
			},
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.FunctionBody,
						*semantic.FunctionParam:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
							return semantic.Int
						}
					case *semantic.NativeVariableDeclaration:
						switch n.Identifier.Name {
						case "add":
							return ft
						case "two":
							return semantic.Int
						}
					case *semantic.FunctionExpression:
						return ft
					case *semantic.ObjectExpression:
						return semantic.NewObjectType(map[string]semantic.Type{
							"a": semantic.Int,
						})
					}
					return nil
				},
			},
		},
		{
			name: "redeclaration",
			program: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: true},
					},
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: false},
					},
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: false},
					},
				},
			},
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.FunctionBody,
						*semantic.FunctionParam:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
							return semantic.Int
						}
					case *semantic.NativeVariableDeclaration:
						switch n.Identifier.Name {
						case "add":
							return ft
						case "two":
							return semantic.Int
						}
					case *semantic.FunctionExpression:
						return ft
					case *semantic.ObjectExpression:
						return semantic.NewObjectType(map[string]semantic.Type{
							"a": semantic.Int,
						})
					}
					return nil
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			semantic.Walk(tc.solution, tc.program)
			wantSolution := tc.solution.Solution()
			solution, err := semantic.InferTypes(tc.program)
			if err != nil {
				t.Error(err)
			}
			if want, got := len(wantSolution), len(solution); got != want {
				t.Errorf("unexpected solution length want %d got %d", want, got)
			}
			for n, want := range wantSolution {
				got := solution[n]
				if got != want {
					t.Errorf("unexpected type for node %#v, want %v got %v", n, want, got)
				}
			}
			for n := range solution {
				_, ok := wantSolution[n]
				if !ok {
					t.Errorf("unexpected extra nodes in solution node %#v", n)
				}
			}
			//t.Log(solution)
		})
	}
}
