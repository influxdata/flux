package semantic_test

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
)

func TestInferTypes(t *testing.T) {
	testCases := []struct {
		name     string
		node     semantic.Node
		script   string
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
			script: `
a = true
a = 13
			`,
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
			wantErr: errors.New(`type error 3:1-3:7: bool != int`),
		},
		{
			name: "array expression",
			node: &semantic.ArrayExpression{
				Elements: []semantic.Expression{
					&semantic.IntegerLiteral{Value: 0},
					&semantic.IntegerLiteral{Value: 1},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.ArrayExpression:
						return semantic.NewArrayPolyType(semantic.Int)
					}
					return nil
				},
			},
		},
		{
			name: "var assignment with binary expression",
			script: `
a = 1 + 1
`,
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
			script: `
f = (a) => 1 + a
`,
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
					case *semantic.FunctionParameters:
						return in
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
							In:           in,
							Defaults:     nil,
							Out:          semantic.Int,
							PipeArgument: "",
						})
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(nil)
					}
					return nil
				},
			},
		},
		{
			name: "var assignment with function with defaults",
			script: `
f = (a,b=0) => a + b
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					})
					defaults := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"b": semantic.Int,
					})
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.IdentifierExpression,
						*semantic.Property,
						*semantic.FunctionParameter:
						return semantic.Int
					case *semantic.FunctionParameters:
						return in
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
							In:           in,
							Defaults:     defaults,
							Out:          semantic.Int,
							PipeArgument: "",
						})
					case *semantic.ObjectExpression:
						return defaults
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
					case *semantic.FunctionParameters:
						return in
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
							In:           in,
							Defaults:     nil,
							Out:          semantic.Int,
							PipeArgument: "",
						})
					case *semantic.ObjectExpression:
						return in
					}
					return nil
				},
			},
		},
		{
			name: "call function identifier",
			script: `
			add = (a) => 1 + a
			two = add(a:1)
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
					})
					ft := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           in,
						Defaults:     nil,
						Out:          semantic.Int,
						PipeArgument: "",
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.Property,
						*semantic.FunctionParameter:
						return semantic.Int
					case *semantic.FunctionParameters:
						return in
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
						switch n.Location().Start.Line {
						case 2:
							return semantic.NewObjectPolyType(nil)
						case 3:
							return in
						}
					}
					return nil
				},
			},
		},
		{
			name: "call polymorphic identity",
			script: `
identity = (x) => x
identity(x:identity)(x:2)
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					f := semantic.NewFresher()
					tv0 := f.Fresh()
					in := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"x": tv0,
					})
					out := tv0
					ft := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           in,
						Defaults:     nil,
						Out:          out,
						PipeArgument: "",
					})

					inInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"x": semantic.Int,
					})
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           inInt,
						Defaults:     nil,
						Out:          outInt,
						PipeArgument: "",
					})

					inF := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"x": ftInt,
					})
					outF := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           inInt,
						Defaults:     nil,
						Out:          outInt,
						PipeArgument: "",
					})
					ftF := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           inF,
						Defaults:     nil,
						Out:          outF,
						PipeArgument: "",
					})
					switch n := node.(type) {
					case *semantic.CallExpression:
						switch l := n.Location().Start.Column; l {
						case 1:
							return outF
						case 21:
							return outInt
						}
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "identity":
							switch l := n.Location().Start.Column; l {
							case 1:
								return ftF
							case 12:
								return ftInt
							}
						case "x":
							switch l := n.Location().Start.Column; l {
							case 2:
								return ftInt
							case 19:
								return out
							}
						}
					case *semantic.ExpressionStatement:
						return outInt
					case *semantic.FunctionParameter:
						return out
					case *semantic.Property:
						switch l := n.Location().Start.Column; l {
						case 10:
							return outF
						case 22:
							return outInt
						}
					case *semantic.NativeVariableDeclaration,
						*semantic.FunctionExpression:
						return ft
					case *semantic.ObjectExpression:
						switch l := n.Location().Start.Line; l {
						case 2:
							return semantic.NewObjectPolyType(nil)
						case 3:
							switch c := n.Location().Start.Column; c {
							case 10:
								return inF
							case 22:
								return inInt
							}
						}
					case *semantic.FunctionParameters:
						return in
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
			script: `
(r) => {
	f = (a,b) => a + b
	return f(a:1, b:r)
}`,
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
					ft := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           in,
						Defaults:     nil,
						Out:          out,
						PipeArgument: "",
					})
					inInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					})
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           inInt,
						Defaults:     nil,
						Out:          outInt,
						PipeArgument: "",
					})
					inR := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"r": semantic.Int,
					})
					ftR := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:           inR,
						Defaults:     nil,
						Out:          semantic.Int,
						PipeArgument: "",
					})
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
					case *semantic.FunctionParameters:
						switch l := n.Location().Start.Line; l {
						case 2:
							return inR
						case 3:
							return in
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
						switch l := n.Location().Start.Line; l {
						case 2, 3:
							return semantic.NewObjectPolyType(nil)
						case 4:
							return inInt
						}
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
					case *semantic.ExpressionStatement:
						return semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
							In:  inR,
							Out: semantic.Int,
						})
					}
					return nil
				},
			},
		},
		{
			name: "function call with defaults",
			script: `
foo = (f) => f(a:1, b:2)
add = (a,b,c=1) => a + b + c
foo(f:add)
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					f := semantic.NewFresher()
					_ = f.Fresh()
					tv1 := f.Fresh()
					inCall := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					})
					outCall := tv1
					call := semantic.NewCallFunctionPolyType(inCall, outCall)

					inFoo := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"f": call,
					})
					outFoo := outCall
					foo := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:  inFoo,
						Out: outFoo,
					})

					inCallInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					})
					outCallInt := semantic.Int
					callInt := semantic.NewCallFunctionPolyType(inCallInt, outCallInt)

					inFooInt := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"f": callInt,
					})
					outFooInt := outCallInt
					fooInt := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:  inFooInt,
						Out: outFooInt,
					})

					inAdd := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
						"c": semantic.Int,
					})
					defaultsAdd := semantic.NewObjectPolyType(map[string]semantic.PolyType{
						"c": semantic.Int,
					})
					outAdd := semantic.Int
					add := semantic.NewFunctionPolyType(semantic.PolyFunctionSignature{
						In:       inAdd,
						Defaults: defaultsAdd,
						Out:      outAdd,
					})

					out := semantic.Int
					switch n := node.(type) {
					case *semantic.ExpressionStatement:
						return out
					case *semantic.NativeVariableDeclaration:
						switch l := n.Location().Start.Line; l {
						case 2:
							return foo
						case 3:
							return add
						}
					case *semantic.FunctionExpression:
						switch l := n.Location().Start.Line; l {
						case 2:
							return foo
						case 3:
							return add
						}
					case *semantic.FunctionParameters:
						switch l := n.Location().Start.Line; l {
						case 2:
							return inFoo
						case 3:
							return inAdd
						}
					case *semantic.FunctionParameter:
						switch l := n.Location().Start.Line; l {
						case 2:
							return call
						case 3:
							return semantic.Int
						}
					case *semantic.ObjectExpression:
						switch l := n.Location().Start.Line; l {
						case 2:
							switch l := n.Location().Start.Column; l {
							case 7:
								return semantic.NewObjectPolyType(nil)
							case 16:
								return inCall
							}
						case 3:
							return defaultsAdd
						case 4:
							return semantic.NewObjectPolyType(map[string]semantic.PolyType{
								"f": add,
							})
						}
					case *semantic.Property:
						switch l := n.Location().Start.Line; l {
						case 2, 3:
							return semantic.Int
						case 4:
							return add
						}
					case *semantic.CallExpression:
						switch l := n.Location().Start.Line; l {
						case 2:
							return outCall
						case 4:
							return out
						}
					case *semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "a", "b", "c":
							return semantic.Int
						case "foo":
							return fooInt
						case "add":
							return add
						case "f":
							return call
						}
					}
					return nil
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.script != "" {
				program, err := parser.NewAST(tc.script)
				if err != nil {
					t.Fatal(err)
				}
				node, err := semantic.New(program)
				if err != nil {
					t.Fatal(err)
				}
				tc.node = node
			}
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
			wantNodes := make([]semantic.Node, 0, len(wantSolution))
			for n := range wantSolution {
				wantNodes = append(wantNodes, n)
			}
			sortNodes(wantNodes)
			for _, n := range wantNodes {
				want := wantSolution[n]
				got := gotSolution[n]
				if !got.Equal(want) {
					t.Errorf("unexpected type for node %T@%v, want: %v got: %v", n, n.Location(), want, got)
				}
			}
			gotNodes := make([]semantic.Node, 0, len(gotSolution))
			for n := range gotSolution {
				gotNodes = append(gotNodes, n)
			}
			sortNodes(gotNodes)
			for _, n := range gotNodes {
				_, ok := wantSolution[n]
				if !ok {
					t.Errorf("unexpected extra nodes in solution node %T@%v", n, n.Location())
				}
			}
			t.Log("got solution:", gotSolution)
		})
	}
}

type SolutionMap map[semantic.Node]semantic.PolyType

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	nodes := make([]semantic.Node, 0, len(s))
	for n := range s {
		nodes = append(nodes, n)
	}
	sortNodes(nodes)
	for _, n := range nodes {
		t := s[n]
		fmt.Fprintf(&builder, "%T@%v: %v\n", n, n.Location(), t)
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

func sortNodes(nodes []semantic.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Location().Less(nodes[j].Location())
	})
}
