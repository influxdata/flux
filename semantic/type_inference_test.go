package semantic_test

import (
	"errors"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
)

func TestInferTypes(t *testing.T) {
	testCases := []struct {
		name     string
		node     semantic.Node
		solution SolutionVisitor
		wantErr  error
	}{
		{
			name: "var assignment",
			node: &semantic.Program{
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
			name: "redeclaration",
			node: &semantic.Program{
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
					switch node.(type) {
					case *semantic.NativeVariableDeclaration:
						return semantic.Bool
					}
					return nil
				},
			},
		},
		{
			name: "redeclaration error",
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: true},
					},
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.IntegerLiteral{Value: 13},
					},
				},
			},
			// TODO(nathanielc): Get better errors providing context in the type constraints
			wantErr: errors.New(`type error: bool != int`),
		},
		{
			name: "var assignment with binary expression",
			node: &semantic.Program{
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
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "f"},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Parameters: &semantic.FunctionParameters{
									List: []*semantic.FunctionParameter{{
										Key: &semantic.Identifier{Name: "a"},
									}},
								},
								Body: &semantic.BinaryExpression{
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
						*semantic.FunctionBlock,
						*semantic.FunctionParameter,
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
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "two"},
						Init: &semantic.CallExpression{
							Callee: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{{
											Key: &semantic.Identifier{Name: "a"},
										}},
									},
									Body: &semantic.BinaryExpression{
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
						*semantic.FunctionBlock,
						*semantic.FunctionParameter,
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
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "add"},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Parameters: &semantic.FunctionParameters{
									List: []*semantic.FunctionParameter{{
										Key: &semantic.Identifier{Name: "a"},
									}},
								},
								Body: &semantic.BinaryExpression{
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
						*semantic.FunctionBlock,
						*semantic.FunctionParameter:
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
			name: "call polymorphic function",
			node: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "add"},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Parameters: &semantic.FunctionParameters{
									List: []*semantic.FunctionParameter{
										{Key: &semantic.Identifier{Name: "a"}},
										{Key: &semantic.Identifier{Name: "b"}},
									},
								},
								Body: &semantic.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &semantic.IdentifierExpression{Name: "a"},
									Right:    &semantic.IdentifierExpression{Name: "b"},
								},
							},
						},
					},
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "two"},
						Init: &semantic.CallExpression{
							Callee: &semantic.IdentifierExpression{Name: "add"},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key:   &semantic.Identifier{Name: "a"},
										Value: &semantic.IntegerLiteral{Value: 1},
									},
									{
										Key:   &semantic.Identifier{Name: "b"},
										Value: &semantic.IntegerLiteral{Value: 1},
									},
								},
							},
						},
					},
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "hello"},
						Init: &semantic.CallExpression{
							Callee: &semantic.IdentifierExpression{Name: "add"},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key:   &semantic.Identifier{Name: "a"},
										Value: &semantic.StringLiteral{Value: "hello "},
									},
									{
										Key:   &semantic.Identifier{Name: "b"},
										Value: &semantic.StringLiteral{Value: "world!"},
									},
								},
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
						*semantic.FunctionBlock,
						*semantic.FunctionParameter:
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
			name: "extern",
			node: &semantic.Extern{
				Declarations: []*semantic.ExternalVariableDeclaration{{
					Identifier: &semantic.Identifier{Name: "foo"},
					TypeScheme: semantic.Int,
				}},
				Program: &semantic.Program{
					Body: []semantic.Statement{
						&semantic.ExpressionStatement{
							Expression: &semantic.IdentifierExpression{Name: "foo"},
						},
					},
				},
			},
			solution: &visitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.IdentifierExpression,
						*semantic.ExternalVariableDeclaration:
						return semantic.Int
					}
					return nil
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			solution, err := semantic.InferTypes(tc.node)
			if err != nil {
				if tc.wantErr != nil {
					if want, got := tc.wantErr.Error(), err.Error(); want != got {
						t.Fatalf("unexpected error want: %s got: %s", want, got)
					}
					return
				}
				t.Error(err)
			} else if tc.wantErr != nil {
				t.Fatalf("expected error: %s: ", tc.wantErr.Error())
			}
			semantic.Walk(tc.solution, tc.node)
			wantSolution := tc.solution.Solution()
			if want, got := len(wantSolution), len(solution); got != want {
				t.Errorf("unexpected solution length want: %d got: %d", want, got)
			}
			for n, want := range wantSolution {
				got := solution[n]
				if got != want {
					t.Errorf("unexpected type for node %#v, want: %v got: %v", n, want, got)
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

type SolutionVisitor interface {
	semantic.Visitor
	Solution() map[semantic.Node]semantic.Type
}

type visitor struct {
	f        func(node semantic.Node) semantic.Type
	solution map[semantic.Node]semantic.Type
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

func (v *visitor) Solution() map[semantic.Node]semantic.Type {
	return v.solution
}
