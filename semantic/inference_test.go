package semantic_test

import (
	"errors"
	"fmt"
	"strings"
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
				f: func(node semantic.Node) semantic.PolyType {
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
				f: func(node semantic.Node) semantic.PolyType {
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
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
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
					})
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.IdentifierExpression,
						*semantic.FunctionParameter:
						return semantic.Int
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(in, semantic.Int)
					case *semantic.ObjectExpression:
						return in
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
				f: func(node semantic.Node) semantic.PolyType {
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
					})
					switch node.(type) {
					case *semantic.CallExpression,
						*semantic.NativeVariableDeclaration,
						*semantic.BinaryExpression,
						*semantic.Property,
						*semantic.FunctionParameter,
						*semantic.IdentifierExpression:
						return semantic.Int
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(in, semantic.Int)
					case *semantic.ObjectExpression:
						return in
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
				f: func(node semantic.Node) semantic.PolyType {
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
					})
					ft := semantic.NewFunctionPolyType(in, semantic.Int)
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.Property,
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
						return in
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
				f: func(node semantic.Node) semantic.PolyType {
					f := semantic.NewFresher()
					tv := f.Fresh()
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"x": tv,
					})
					ft := semantic.NewFunctionPolyType(in, tv)
					inMono := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"x": semantic.Int,
					})
					ftMono := semantic.NewFunctionPolyType(inMono, semantic.Int)
					switch n := node.(type) {
					case *semantic.ExpressionStatement,
						*semantic.CallExpression,
						*semantic.Property,
						*semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "identity":
							return ftMono
						case "x":
							return tv
						}
					case *semantic.FunctionParameter:
						return tv
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return ft
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(map[string]semantic.PolyType{
							"x": semantic.Int,
						})
					}
					return nil
				},
			},
		},
		{
			name: "call polymorphic identity",
			// identity = (x) => x
			// identity(x:identity)(x:2)
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
			solution: newIdentityCaseVisitor(),
		},
		{
			name: "extern",
			node: &semantic.Extern{
				Declarations: []*semantic.ExternalVariableDeclaration{{
					Identifier: &semantic.Identifier{Name: "foo"},
					ExternType: semantic.Int,
				}},
				Block: &semantic.ExternBlock{
					Node: &semantic.IdentifierExpression{Name: "foo"},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.Extern,
						*semantic.ExternBlock,
						*semantic.ExternalVariableDeclaration,
						*semantic.IdentifierExpression:
						return semantic.Int
					}
					return nil
				},
			},
		},
		{
			name: "nested functions",
			node: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BlockStatement{
						Body: []semantic.Statement{
							&semantic.NativeVariableDeclaration{
								Identifier: &semantic.Identifier{Name: "f"},
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
							&semantic.ReturnStatement{
								Argument: &semantic.CallExpression{
									Callee: &semantic.IdentifierExpression{Name: "f"},
									Arguments: &semantic.ObjectExpression{
										Properties: []*semantic.Property{
											{Key: &semantic.Identifier{Name: "a"}, Value: &semantic.IntegerLiteral{Value: 1}},
											{Key: &semantic.Identifier{Name: "b"}, Value: &semantic.IdentifierExpression{Name: "r"}},
										},
									},
								},
							},
						},
					},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					f := semantic.NewFresher()
					_ = f.Fresh()
					tv1 := f.Fresh()
					tv2 := f.Fresh()
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": tv1,
						"b": tv2,
					})
					out := tv1
					ft := semantic.NewFunctionPolyType(in, out)
					inInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					})
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(inInt, outInt)
					ftR := semantic.NewFunctionPolyType(
						semantic.NewObjectPolyType(map[string]semantic.PolyType{
							"r": semantic.Int,
						}),
						semantic.Int,
					)
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "a":
							return tv1
						case "b":
							return tv2
						case "r":
							return outInt
						case "f":
							return ftInt
						}
					case *semantic.FunctionExpression:
						switch n.Block.Body.(type) {
						case semantic.Statement:
							return ftR
						case semantic.Expression:
							return ft
						}
					case *semantic.FunctionParameter:
						switch n.Key.Name {
						case "a":
							return tv1
						case "b":
							return tv2
						case "r":
							return outInt
						}
					case *semantic.ObjectExpression:
						return inInt
					case *semantic.Property:
						return outInt
					case *semantic.NativeVariableDeclaration:
						return ft
					case *semantic.BinaryExpression:
						return out
					case *semantic.BlockStatement,
						*semantic.ReturnStatement,
						*semantic.CallExpression:
						return outInt
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

			ts := semantic.Infer(tc.node)

			// Read all types from node
			v := &typeVisitor{
				typeSolution: ts,
			}
			semantic.Walk(v, tc.node)
			gotSolution, err := v.Solution()
			if err != nil {
				if tc.wantErr != nil {
					if got, want := err.Error(), tc.wantErr.Error(); got != want {
						t.Fatalf("unexpected error want: %s got: %s", want, got)
					}
					return
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
				if !got.Equal(want) {
					t.Errorf("unexpected type for node %#v, want: %v got: %v", n, want, got)
				}
			}
			for n := range gotSolution {
				_, ok := wantSolution[n]
				if !ok {
					t.Errorf("unexpected extra nodes in solution node %#v", n)
				}
			}
			t.Log(gotSolution)
		})
	}
}

type SolutionMap map[semantic.Node]semantic.PolyType

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	for n, t := range s {
		fmt.Fprintf(&builder, "%T: %v\n", n, t)
	}
	builder.WriteString("}")
	return builder.String()
}

type SolutionVisitor interface {
	semantic.Visitor
	Solution() SolutionMap
}

type solutionVisitor struct {
	f        func(node semantic.Node) semantic.PolyType
	solution SolutionMap
}

func (v *solutionVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.solution == nil {
		v.solution = make(SolutionMap)
	}
	// Handle literals here
	if l, ok := node.(semantic.Literal); ok {
		var typ semantic.PolyType
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

func (v *solutionVisitor) Done(semantic.Node) {}

func (v *solutionVisitor) Solution() SolutionMap {
	return v.solution
}

type typeVisitor struct {
	typeSolution semantic.TypeSolution
	solution     SolutionMap
	err          error
}

func (v *typeVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.err != nil {
		return nil
	}
	if v.solution == nil {
		v.solution = make(SolutionMap)
	}
	t, err := v.typeSolution.PolyTypeOf(node)
	if err != nil {
		v.err = err
		return nil
	}
	if t != nil {
		v.solution[node] = t
	}
	return v
}

func (v *typeVisitor) Done(semantic.Node) {}

func (v *typeVisitor) Solution() (SolutionMap, error) {
	return v.solution, v.err
}

// identityCaseVisitor implements the SolutionVisitor interface for the "call polymorphic identity" test case
// This type is neccessary to track where we are within the semantic graph.
type identityCaseVisitor struct {
	solution SolutionMap

	propCount,
	identICount,
	identXCount,
	callCount,
	objCount int

	in  semantic.PolyType
	out semantic.PolyType
	ft  semantic.PolyType

	inF  semantic.PolyType
	outF semantic.PolyType
	ftF  semantic.PolyType

	inInt  semantic.PolyType
	outInt semantic.PolyType
	ftInt  semantic.PolyType
}

func newIdentityCaseVisitor() *identityCaseVisitor {
	f := semantic.NewFresher()
	tv0 := f.Fresh()
	in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
		"x": tv0,
	})
	out := tv0
	ft := semantic.NewFunctionPolyType(in, out)

	inInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
		"x": semantic.Int,
	})
	outInt := semantic.Int
	ftInt := semantic.NewFunctionPolyType(inInt, outInt)

	inF := semantic.NewObjectPolyType(map[string]semantic.PolyType{
		"x": ftInt,
	})
	outF := semantic.NewFunctionPolyType(inInt, outInt)
	ftF := semantic.NewFunctionPolyType(inF, outF)

	return &identityCaseVisitor{
		solution: make(SolutionMap),
		in:       in,
		out:      out,
		ft:       ft,
		inF:      inF,
		outF:     outF,
		ftF:      ftF,
		inInt:    inInt,
		outInt:   outInt,
		ftInt:    ftInt,
	}
}

func (v *identityCaseVisitor) Visit(node semantic.Node) semantic.Visitor {
	switch n := node.(type) {
	case *semantic.CallExpression:
		v.callCount++
		if v.callCount == 1 {
			v.solution[n] = v.outInt
		} else {
			v.solution[n] = v.outF
		}
	case *semantic.IdentifierExpression:
		switch n.Name {
		case "identity":
			v.identICount++
			if v.identICount == 1 {
				v.solution[n] = v.ftF
			} else {
				v.solution[n] = v.ftInt
			}
		case "x":
			v.identXCount++
			if v.identXCount == 1 {
				v.solution[n] = v.out
			} else {
				v.solution[n] = v.ftInt
			}
		}
	case *semantic.ExpressionStatement:
		v.solution[n] = v.outInt
	case *semantic.FunctionParameter:
		v.solution[n] = v.out
	case *semantic.Property:
		v.propCount++
		if v.propCount == 1 {
			v.solution[n] = v.outF
		} else {
			v.solution[n] = v.outInt
		}
	case *semantic.NativeVariableDeclaration,
		*semantic.FunctionExpression:
		v.solution[n] = v.ft
	case *semantic.ObjectExpression:
		v.objCount++
		if v.objCount == 1 {
			v.solution[n] = v.inF
		} else {
			v.solution[n] = v.inInt
		}

	case *semantic.IntegerLiteral:
		v.solution[n] = semantic.Int
	}
	return v
}

func (v *identityCaseVisitor) Done(semantic.Node) {
}
func (v *identityCaseVisitor) Solution() SolutionMap {
	return v.solution
}
