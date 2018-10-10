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
			name: "bool",
			node: &semantic.BooleanLiteral{Value: false},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					return nil
				},
			},
		},
		{
			name: "bool decl",
			node: &semantic.NativeVariableDeclaration{
				Identifier: &semantic.Identifier{Name: "b"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					return nil
				},
			},
		},
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.BinaryExpression:
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.IdentifierExpression:
						return semantic.Int
					case *semantic.FunctionExpression:
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
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
			name: "call polymorphic function twice",
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
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
			// identity = (x) => x
			// identity(identity)(2)
			name: "call polymorphic identity",
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
									Value: &semantic.IntegerLiteral{Value: 2},
								}},
							},
						},
					},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					ft := semantic.NewFunctionType(semantic.FunctionSignature{
						Params: map[string]semantic.Type{
							"a": semantic.Int,
						},
						ReturnType: semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "add":
							return ft
						case "a":
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.Type {
					switch node.(type) {
					case *semantic.IdentifierExpression:
						return semantic.Int
					}
					return nil
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var wantSolution SolutionMap
			if tc.solution != nil {
				semantic.Walk(tc.solution, tc.node)
				wantSolution = tc.solution.Solution()
			}

			semantic.Infer(tc.node)

			// Read all types from node
			v := new(typeVisitor)
			semantic.Walk(v, tc.node)
			gotSolution, err := v.Solution()
			if err != nil {
				if tc.wantErr != nil {
					if got, want := err.Error(), tc.wantErr.Error(); got != want {
						t.Fatalf("unexpected error want: %s got: %s", want, got)
					}
				}
				t.Fatal(err)
			} else if tc.wantErr != nil {
				t.Fatalf("expected error: %v", tc.wantErr)
			}

			if want, got := len(wantSolution), len(gotSolution); got != want {
				t.Errorf("unexpected solution length want: %d got: %d", want, got)
			}
			for n, want := range wantSolution {
				got := gotSolution[n]
				if got != want {
					t.Errorf("unexpected type for node %#v, want: %v got: %v", n, want, got)
				}
			}
			for n := range gotSolution {
				_, ok := wantSolution[n]
				if !ok {
					t.Errorf("unexpected extra nodes in solution node %#v", n)
				}
			}
			//t.Log(solution)
		})
	}
}

type SolutionMap map[semantic.Node]semantic.Type

type SolutionVisitor interface {
	semantic.Visitor
	Solution() SolutionMap
}

type solutionVisitor struct {
	f        func(node semantic.Node) semantic.Type
	solution SolutionMap
}

func (v *solutionVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.solution == nil {
		v.solution = make(SolutionMap)
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

func (v *solutionVisitor) Done() {}

func (v *solutionVisitor) Solution() SolutionMap {
	return v.solution
}

type typeVisitor struct {
	solution SolutionMap
	err      error
}

func (v *typeVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.err != nil {
		return nil
	}
	if v.solution == nil {
		v.solution = make(SolutionMap)
	}
	if e, ok := node.(semantic.Expression); ok {
		t, err := e.Type()
		if err != nil {
			v.err = err
			return nil
		}
		v.solution[node] = t
	}
	return v
}

func (v *typeVisitor) Done() {}

func (v *typeVisitor) Solution() (SolutionMap, error) {
	return v.solution, v.err
}
