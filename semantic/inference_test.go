package semantic_test

import (
	"errors"
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
		importer semantic.Importer
		skip     string
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
			node: &semantic.NativeVariableAssignment{
				Identifier: &semantic.Identifier{Name: "b"},
				Init:       &semantic.BooleanLiteral{Value: false},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					return nil
				},
			},
		},
		{
			name: "instantiate array",
			script: `
				a = [0, 1]
				x = a[0]`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.IntegerLiteral:
						return semantic.Int
					case *semantic.ArrayExpression:
						return semantic.NewArrayPolyType(semantic.Int)
					case *semantic.IdentifierExpression:
						return semantic.NewArrayPolyType(semantic.Int)
					case *semantic.IndexExpression:
						return semantic.Int
					}
					return nil
				},
			},
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
					case *semantic.BinaryExpression:
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
					params := map[string]semantic.PolyType{
						"a": semantic.Int,
					}
					required := semantic.LabelSet{"a"}
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.IdentifierExpression,
						*semantic.FunctionBlock,
						*semantic.FunctionParameter:
						return semantic.Int
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: params,
							Required:   required,
							Return:     semantic.Int,
						})
					case *semantic.ObjectExpression:
						return semantic.NewEmptyObjectPolyType()
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
					params := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					required := semantic.LabelSet{"a"}
					switch node.(type) {
					case *semantic.BinaryExpression,
						*semantic.IdentifierExpression,
						*semantic.Property,
						*semantic.FunctionBlock,
						*semantic.FunctionParameter:
						return semantic.Int
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: params,
							Required:   required,
							Return:     semantic.Int,
						})
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"b": semantic.Int,
							},
							nil,
							semantic.LabelSet{"b"},
						)
					}
					return nil
				},
			},
		},
		{
			name: "call function",
			node: &semantic.File{
				Body: []semantic.Statement{
					&semantic.NativeVariableAssignment{
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
					params := map[string]semantic.PolyType{
						"a": semantic.Int,
					}
					required := semantic.LabelSet{"a"}
					switch node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.Property,
						*semantic.FunctionBlock,
						*semantic.FunctionParameter,
						*semantic.IdentifierExpression:
						return semantic.Int
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: params,
							Required:   required,
							Return:     semantic.Int,
						})
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a": semantic.Int,
							},
							nil,
							required,
						)
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
					params := map[string]semantic.PolyType{
						"a": semantic.Int,
					}
					required := semantic.LabelSet{"a"}
					ft := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: params,
						Required:   required,
						Return:     semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.CallExpression,
						*semantic.BinaryExpression,
						*semantic.Property,
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
					case *semantic.FunctionExpression:
						return ft
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return semantic.NewEmptyObjectPolyType()
						case 3:
							return semantic.NewObjectPolyType(
								params,
								nil,
								required,
							)
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
					tv := semantic.Tvar(3)
					params := map[string]semantic.PolyType{
						"x": tv,
					}
					required := semantic.LabelSet{"x"}
					out := tv
					ft := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: params,
						Required:   required,
						Return:     out,
					})

					paramsInt := map[string]semantic.PolyType{
						"x": semantic.Int,
					}
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsInt,
						Required:   required,
						Return:     outInt,
					})

					paramsF := map[string]semantic.PolyType{
						"x": ftInt,
					}
					outF := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsInt,
						Required:   required,
						Return:     outInt,
					})
					ftF := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsF,
						Required:   required,
						Return:     outF,
					})
					switch n := node.(type) {
					case *semantic.CallExpression:
						switch n.Location().End.Column {
						case 21:
							return outF
						case 26:
							return outInt
						}
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "identity":
							switch n.Location().Start.Column {
							case 1:
								return ftF
							case 12:
								return ftInt
							}
						case "x":
							switch n.Location().Start.Column {
							case 2:
								return ftInt
							case 19:
								return out
							}
						}
					case *semantic.FunctionParameter:
						return out
					case *semantic.Property:
						switch n.Location().Start.Column {
						case 10:
							return outF
						case 22:
							return outInt
						}
					case *semantic.FunctionExpression:
						return ft
					case *semantic.FunctionBlock:
						return out
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return semantic.NewEmptyObjectPolyType()
						case 3:
							switch n.Location().Start.Column {
							case 10:
								return semantic.NewObjectPolyType(
									paramsF,
									nil,
									required,
								)
							case 22:
								return semantic.NewObjectPolyType(
									paramsInt,
									nil,
									required,
								)
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "extern",
			node: &semantic.Extern{
				Assignments: []*semantic.ExternalVariableAssignment{{
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
					case *semantic.IdentifierExpression:
						return semantic.Int
					}
					return nil
				},
			},
		},
		{
			name: "extern object",
			node: &semantic.Extern{
				Assignments: []*semantic.ExternalVariableAssignment{{
					Identifier: &semantic.Identifier{Name: "foo"},
					ExternType: semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"x": semantic.Tvar(9),
						},
						nil,
						semantic.LabelSet{"x"},
					),
				}},
				Block: &semantic.ExternBlock{
					Node: &semantic.IdentifierExpression{Name: "foo"},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.IdentifierExpression:
						return semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"x": semantic.Tvar(5),
							},
							nil,
							semantic.LabelSet{"x"},
						)
					}
					return nil
				},
			},
		},
		{
			name: "extern type variables",
			node: &semantic.Extern{
				Assignments: []*semantic.ExternalVariableAssignment{
					{
						Identifier: &semantic.Identifier{Name: "f"},
						ExternType: semantic.NewFunctionPolyType(
							semantic.FunctionPolySignature{
								Return: semantic.Tvar(3),
							},
						),
					},
					{
						Identifier: &semantic.Identifier{Name: "g"},
						ExternType: semantic.NewFunctionPolyType(
							semantic.FunctionPolySignature{
								Return: semantic.Tvar(5),
							},
						),
					},
				},
				Block: &semantic.ExternBlock{
					Node: &semantic.File{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "a"},
								Init:       &semantic.IdentifierExpression{Name: "f"},
							},
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "b"},
								Init:       &semantic.IdentifierExpression{Name: "g"},
							},
						},
					},
				},
			},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					f := semantic.NewFunctionPolyType(
						semantic.FunctionPolySignature{
							Return: semantic.Tvar(7),
						})
					g := semantic.NewFunctionPolyType(
						semantic.FunctionPolySignature{
							Return: semantic.Tvar(8),
						})
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "f":
							return f
						case "g":
							return g
						}
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
					tv := semantic.Tvar(8)
					params := map[string]semantic.PolyType{
						"a": tv,
						"b": tv,
					}
					requiredAB := semantic.LabelSet{"a", "b"}
					out := tv
					ft := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: params,
						Required:   requiredAB,
						Return:     out,
					})
					paramsInt := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsInt,
						Required:   requiredAB,
						Return:     outInt,
					})
					paramsR := map[string]semantic.PolyType{
						"r": semantic.Int,
					}
					requiredR := semantic.LabelSet{"r"}
					ftR := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsR,
						Required:   requiredR,
						Return:     semantic.Int,
					})
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "a":
							return tv
						case "b":
							return tv
						case "r":
							return outInt
						case "f":
							return ftInt
						}
					case *semantic.FunctionExpression:
						switch n.Location().Start.Line {
						case 2:
							return ftR
						case 3:
							return ft
						}
					case *semantic.FunctionBlock:
						switch n.Location().Start.Line {
						case 2:
							return outInt
						case 3:
							return tv
						}
					case *semantic.FunctionParameter:
						switch n.Key.Name {
						case "a":
							return tv
						case "b":
							return tv
						case "r":
							return outInt
						}
					case *semantic.Property:
						return outInt
					case *semantic.BinaryExpression:
						return out
					case *semantic.Block,
						*semantic.ReturnStatement,
						*semantic.CallExpression:
						return outInt
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2, 3:
							return semantic.NewEmptyObjectPolyType()
						case 4:
							return semantic.NewObjectPolyType(
								paramsInt,
								nil,
								requiredAB,
							)
						}
					}
					return nil
				},
			},
		},
		{
			name: "function call with and without defaults",
			script: `
add = (a,b,c=1) => a + b + c
add(a:1,b:2,c:1)
add(a:1,b:2)
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					requiredAB := semantic.LabelSet{"a", "b"}
					requiredABC := semantic.LabelSet{"a", "b", "c"}

					callWith := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
						"c": semantic.Int,
					}
					callWithout := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}

					objWith := semantic.NewObjectPolyType(
						callWith,
						nil,
						requiredABC,
					)
					objWithout := semantic.NewObjectPolyType(
						callWithout,
						nil,
						requiredAB,
					)

					paramsAdd := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
						"c": semantic.Int,
					}
					outAdd := semantic.Int
					add := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsAdd,
						Required:   requiredAB,
						Return:     outAdd,
					})

					switch n := node.(type) {
					case *semantic.FunctionExpression:
						return add
					case *semantic.FunctionBlock:
						return outAdd
					case *semantic.FunctionParameter:
						return semantic.Int
					case *semantic.Property:
						return semantic.Int
					case *semantic.CallExpression:
						return outAdd
					case *semantic.BinaryExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "a", "b", "c":
							return semantic.Int
						case "add":
							return add
						}
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"c": semantic.Int,
								},
								nil,
								semantic.LabelSet{"c"},
							)
						case 3:
							return objWith
						case 4:
							return objWithout
						}
					}
					return nil
				},
			},
		},
		{
			name: "high order function call without defaults",
			script: `
foo = (f) => f(a:1, b:2)
add = (a,b,c=1) => a + b + c
foo(f:add)
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					tv := semantic.Tvar(29)
					paramsCall := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					requiredAB := semantic.LabelSet{"a", "b"}
					outCall := tv
					call := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsCall,
						Required:   requiredAB,
						Return:     outCall,
					})

					paramsFoo := map[string]semantic.PolyType{
						"f": call,
					}
					requiredF := semantic.LabelSet{"f"}
					outFoo := outCall
					foo := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsFoo,
						Required:   requiredF,
						Return:     outFoo,
					})

					paramsCallInt := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					outCallInt := semantic.Int

					callInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsCallInt,
						Required:   requiredAB,
						Return:     outCallInt,
					})
					paramsFooInt := map[string]semantic.PolyType{
						"f": callInt,
					}
					outFooInt := outCallInt
					fooInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsFooInt,
						Required:   requiredF,
						Return:     outFooInt,
					})

					paramsAdd := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
						"c": semantic.Int,
					}
					outAdd := semantic.Int
					add := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsAdd,
						Required:   requiredAB,
						Return:     outAdd,
					})

					out := semantic.Int
					switch n := node.(type) {
					case *semantic.FunctionExpression:
						switch n.Location().Start.Line {
						case 2:
							return foo
						case 3:
							return add
						}
					case *semantic.FunctionBlock:
						switch n.Location().Start.Line {
						case 2:
							return outFoo
						case 3:
							return out
						}
					case *semantic.FunctionParameter:
						switch n.Location().Start.Line {
						case 2:
							return call
						case 3:
							return semantic.Int
						}
					case *semantic.Property:
						switch n.Location().Start.Line {
						case 2, 3:
							return semantic.Int
						case 4:
							return add
						}
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
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
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							switch n.Location().Start.Column {
							case 7:
								return semantic.NewObjectPolyType(
									nil,
									nil,
									nil,
								)
							case 16:
								return semantic.NewObjectPolyType(
									paramsCallInt,
									nil,
									requiredAB,
								)
							}
						case 3:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"c": semantic.Int,
								},
								nil,
								semantic.LabelSet{"c"},
							)
						case 4:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"f": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
										Parameters: map[string]semantic.PolyType{
											"a": semantic.Int,
											"b": semantic.Int,
											"c": semantic.Int,
										},
										Required: requiredAB,
										Return:   semantic.Int,
									}),
								},
								nil,
								requiredF,
							)
						}
					}
					return nil
				},
			},
		},
		{
			name: "high order function call with defaults",
			script: `
foo = (f) => f(a:1, b:2)
add = (a,b=1) => a + b
foo(f:add)
			`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					tv := semantic.Tvar(26)
					paramsCall := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					requiredAB := semantic.LabelSet{"a", "b"}
					outCall := tv
					call := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsCall,
						Required:   requiredAB,
						Return:     outCall,
					})

					paramsFoo := map[string]semantic.PolyType{
						"f": call,
					}
					requiredF := semantic.LabelSet{"f"}
					outFoo := outCall
					foo := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsFoo,
						Required:   requiredF,
						Return:     outFoo,
					})

					paramsCallInt := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					outCallInt := semantic.Int

					callInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsCallInt,
						Required:   requiredAB,
						Return:     outCallInt,
					})
					paramsFooInt := map[string]semantic.PolyType{
						"f": callInt,
					}
					outFooInt := outCallInt
					fooInt := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsFooInt,
						Required:   requiredF,
						Return:     outFooInt,
					})

					paramsAdd := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					requiredA := semantic.LabelSet{"a"}
					outAdd := semantic.Int
					add := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: paramsAdd,
						Required:   requiredA,
						Return:     outAdd,
					})

					out := semantic.Int
					switch n := node.(type) {
					case *semantic.FunctionExpression:
						switch n.Location().Start.Line {
						case 2:
							return foo
						case 3:
							return add
						}
					case *semantic.FunctionBlock:
						switch n.Location().Start.Line {
						case 2:
							return outFoo
						case 3:
							return out
						}
					case *semantic.FunctionParameter:
						switch n.Location().Start.Line {
						case 2:
							return call
						case 3:
							return semantic.Int
						}
					case *semantic.Property:
						switch n.Location().Start.Line {
						case 2, 3:
							return semantic.Int
						case 4:
							return add
						}
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
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
					case *semantic.ObjectExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 7:
							return semantic.NewObjectPolyType(
								nil,
								nil,
								nil,
							)
						case l == 2 && c == 16:
							return semantic.NewObjectPolyType(
								paramsCallInt,
								nil,
								requiredAB,
							)
						case l == 3:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"b": semantic.Int,
								},
								nil,
								semantic.LabelSet{"b"},
							)
						case l == 4:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"f": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
										Parameters: map[string]semantic.PolyType{
											"a": semantic.Int,
											"b": semantic.Int,
										},
										Required: requiredA,
										Return:   semantic.Int,
									}),
								},
								nil,
								requiredF,
							)
						}
					}
					return nil
				},
			},
		},
		{
			name: "structural polymorphism",
			script: `
jim  = {name: "Jim", age: 30, weight: 100.0}
jane = {name: "Jane", age: 31}
jimb = {jim with lastName: "Buttler"}
device = {name: 42, lat:28.25892, lon: 15.62234}

name = (p) => p.name

name(p: jim)
name(p: jane)
name(p: jimb)
name(p: device)
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					jim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
						},
						nil,
						semantic.LabelSet{"name", "age", "weight"},
					)
					jimCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age", "weight"},
					)
					pJim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": jimCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					jane := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.String,
							"age":  semantic.Int,
						},
						nil,
						semantic.LabelSet{"name", "age"},
					)
					janeCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.String,
							"age":  semantic.Int,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age"},
					)
					pJane := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": janeCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					jimb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":     semantic.String,
							"age":      semantic.Int,
							"weight":   semantic.Float,
							"lastName": semantic.String,
						},
						nil,
						semantic.LabelSet{"name", "age", "weight", "lastName"},
					)
					jimbCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":     semantic.String,
							"age":      semantic.Int,
							"weight":   semantic.Float,
							"lastName": semantic.String,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age", "weight", "lastName"},
					)
					pJimb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": jimbCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					device := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.Int,
							"lat":  semantic.Float,
							"lon":  semantic.Float,
						},
						nil,
						semantic.LabelSet{"name", "lat", "lon"},
					)
					deviceCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.Int,
							"lat":  semantic.Float,
							"lon":  semantic.Float,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "lat", "lon"},
					)
					pDevice := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": deviceCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)

					tv := semantic.NullableTvar{semantic.Tvar(27)}
					p := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": tv,
						},
						semantic.LabelSet{"name"},
						semantic.AllLabels(),
					)
					name := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": p,
						},
						Required: semantic.LabelSet{"p"},
						Return:   tv,
					})
					nameCallJim := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name":   semantic.String,
									"age":    semantic.Int,
									"weight": semantic.Float,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age", "weight"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallJane := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name": semantic.String,
									"age":  semantic.Int,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallJimb := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name":     semantic.String,
									"age":      semantic.Int,
									"weight":   semantic.Float,
									"lastName": semantic.String,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age", "weight", "lastName"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallDevice := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name": semantic.Int,
									"lat":  semantic.Float,
									"lon":  semantic.Float,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "lat", "lon"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.Int,
					})

					nameJim := semantic.String
					nameJane := semantic.String
					nameJimb := semantic.String
					nameDevice := semantic.Int

					switch n := node.(type) {
					case *semantic.Property:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 9:
							return semantic.String
						case l == 2 && c == 22:
							return semantic.Int
						case l == 2 && c == 31:
							return semantic.Float
						case l == 3 && c == 9:
							return semantic.String
						case l == 3 && c == 23:
							return semantic.Int
						case l == 4 && c == 18:
							return semantic.String
						case l == 5 && c == 11:
							return semantic.Int
						case l == 5 && c == 21:
							return semantic.Float
						case l == 5 && c == 35:
							return semantic.Float
						case l == 9:
							return jimCall
						case l == 10:
							return janeCall
						case l == 11:
							return jimbCall
						case l == 12:
							return deviceCall
						}
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return jim
						case 3:
							return jane
						case 4:
							return jimb
						case 5:
							return device
						case 9:
							return pJim
						case 10:
							return pJane
						case 11:
							return pJimb
						case 12:
							return pDevice
						}
					case *semantic.FunctionExpression:
						return name
					case *semantic.FunctionParameter:
						return p
					case *semantic.FunctionBlock:
						return tv
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
						case 9:
							return nameJim
						case 10:
							return nameJane
						case 11:
							return nameJimb
						case 12:
							return nameDevice
						}
					case *semantic.IdentifierExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 4:
							return jim
						case l == 7:
							return p
						case l == 9 && c == 1:
							return nameCallJim
						case l == 9 && c == 9:
							return jimCall
						case l == 10 && c == 1:
							return nameCallJane
						case l == 10 && c == 9:
							return janeCall
						case l == 11 && c == 1:
							return nameCallJimb
						case l == 11 && c == 9:
							return jimbCall
						case l == 12 && c == 1:
							return nameCallDevice
						case l == 12 && c == 9:
							return deviceCall
						}
					case *semantic.MemberExpression:
						return tv
					}
					return nil
				},
			},
		},
		{
			skip: "skipping until type checking using with is fixed",
			name: "structural polymorphism with object updates and indirection",
			script: `
getJim  = () => ({name: "Jim", age: 30, weight: 100.0})
j = getJim()
john = {j with lastName: "asdf"}

name = (p) => p.name

name(p: john)
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					jim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
						},
						nil,
						semantic.LabelSet{"name", "age", "weight"},
					)
					jimCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age", "weight"},
					)
					pJim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": jimCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					jane := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.String,
							"age":  semantic.Int,
						},
						nil,
						semantic.LabelSet{"name", "age"},
					)
					janeCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.String,
							"age":  semantic.Int,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age"},
					)
					pJane := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": janeCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					jimb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":     semantic.String,
							"age":      semantic.Int,
							"weight":   semantic.Float,
							"lastName": semantic.String,
						},
						nil,
						semantic.LabelSet{"name", "age", "weight", "lastName"},
					)
					jimbCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":     semantic.String,
							"age":      semantic.Int,
							"weight":   semantic.Float,
							"lastName": semantic.String,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age", "weight", "lastName"},
					)
					pJimb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": jimbCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)
					device := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.Int,
							"lat":  semantic.Float,
							"lon":  semantic.Float,
						},
						nil,
						semantic.LabelSet{"name", "lat", "lon"},
					)
					deviceCall := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.Int,
							"lat":  semantic.Float,
							"lon":  semantic.Float,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "lat", "lon"},
					)
					pDevice := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"p": deviceCall,
						},
						nil,
						semantic.LabelSet{"p"},
					)

					tv := semantic.Tvar(50)
					p := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": tv,
						},
						semantic.LabelSet{"name"},
						semantic.AllLabels(),
					)
					name := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": p,
						},
						Required: semantic.LabelSet{"p"},
						Return:   tv,
					})
					nameCallJim := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name":   semantic.String,
									"age":    semantic.Int,
									"weight": semantic.Float,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age", "weight"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallJane := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name": semantic.String,
									"age":  semantic.Int,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallJimb := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name":     semantic.String,
									"age":      semantic.Int,
									"weight":   semantic.Float,
									"lastName": semantic.String,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age", "weight", "lastName"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.String,
					})
					nameCallDevice := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name": semantic.Int,
									"lat":  semantic.Float,
									"lon":  semantic.Float,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "lat", "lon"},
							),
						},
						Required: semantic.LabelSet{"p"},
						Return:   semantic.Int,
					})

					nameJim := semantic.String
					nameJane := semantic.String
					nameJimb := semantic.String
					nameDevice := semantic.Int

					switch n := node.(type) {
					case *semantic.Property:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 9:
							return semantic.String
						case l == 2 && c == 22:
							return semantic.Int
						case l == 2 && c == 31:
							return semantic.Float
						case l == 3 && c == 9:
							return semantic.String
						case l == 3 && c == 23:
							return semantic.Int
						case l == 4 && c == 18:
							return semantic.String
						case l == 5 && c == 11:
							return semantic.Int
						case l == 5 && c == 21:
							return semantic.Float
						case l == 5 && c == 35:
							return semantic.Float
						case l == 9:
							return jimCall
						case l == 10:
							return janeCall
						case l == 11:
							return jimbCall
						case l == 12:
							return deviceCall
						}
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return jim
						case 3:
							return jane
						case 4:
							return jimb
						case 5:
							return device
						case 9:
							return pJim
						case 10:
							return pJane
						case 11:
							return pJimb
						case 12:
							return pDevice
						}
					case *semantic.FunctionExpression:
						return name
					case *semantic.FunctionParameter:
						return p
					case *semantic.FunctionBlock:
						return tv
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
						case 9:
							return nameJim
						case 10:
							return nameJane
						case 11:
							return nameJimb
						case 12:
							return nameDevice
						}
					case *semantic.IdentifierExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 4:
							return jim
						case l == 7:
							return p
						case l == 9 && c == 1:
							return nameCallJim
						case l == 9 && c == 9:
							return jimCall
						case l == 10 && c == 1:
							return nameCallJane
						case l == 10 && c == 9:
							return janeCall
						case l == 11 && c == 1:
							return nameCallJimb
						case l == 11 && c == 9:
							return jimbCall
						case l == 12 && c == 1:
							return nameCallDevice
						case l == 12 && c == 9:
							return deviceCall
						}
					case *semantic.MemberExpression:
						return tv
					}
					return nil
				},
			},
		},
		{
			skip: "skipping until inference using with is fixed",
			name: "structural polymorphism indirection",
			script: `
getJim  = (x) => ({name: "Jim", age: 30, weight: 100.0, x:x})

name = (p) => p.name

j = getJim(x: 5)

name(p: j)
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					x := semantic.Tvar(3)
					jim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
							"x":      x,
						},
						nil,
						semantic.LabelSet{"name", "age", "weight", "x"},
					)
					jimInt := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"weight": semantic.Float,
							"age":    semantic.Int,
							"x":      semantic.Int,
						},
						semantic.LabelSet{"name"},
						semantic.LabelSet{"name", "age", "weight", "x"},
					)
					getJim := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: map[string]semantic.PolyType{
							"x": x,
						},
						Required: semantic.LabelSet{"x"},
						Return:   jim,
					})

					// tv for name function
					tvNameProp := semantic.Tvar(29)

					p := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": tvNameProp,
						},
						semantic.LabelSet{"name"},
						semantic.AllLabels(),
					)

					switch n := node.(type) {
					case *semantic.FunctionBlock:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 11:
							return jim
						case l == 4 && c == 8:
							return tvNameProp
						}
					case *semantic.FunctionExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 11:
							return getJim
						case l == 4 && c == 8:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"p": p,
								},
								Required: semantic.LabelSet{"p"},
								Return:   tvNameProp,
							})
						}
					case *semantic.FunctionParameter:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 12:
							return x
						case l == 4 && c == 9:
							return p
						}
					case *semantic.ObjectExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 19:
							return jim
						case l == 6 && c == 12:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								nil,
								semantic.LabelSet{"x"},
							)
						case l == 8 && c == 10:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"p": jimInt,
								},
								nil,
								semantic.LabelSet{"p"},
							)
						}
					case *semantic.Property:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 20:
							return semantic.String
						case l == 2 && c == 33:
							return semantic.Int
						case l == 2 && c == 42:
							return semantic.Float
						case l == 2 && c == 57:
							return x
						case l == 6 && c == 12:
							return semantic.Int
						case l == 8 && c == 10:
							return jimInt
						}
					case *semantic.IdentifierExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 59:
							return x
						case l == 4 && c == 15:
							return p
						case l == 6 && c == 5:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								Required: semantic.LabelSet{"x"},
								Return:   jimInt,
							})
						case l == 8 && c == 5:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"p": jimInt,
								},
								Required: semantic.LabelSet{"p"},
								Return:   semantic.String,
							})
						case l == 8 && c == 13:
							return jimInt
						case l == 9 && c == 1:
							return semantic.String
						}
					case *semantic.MemberExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 4 && c == 15:
							return tvNameProp
						}
					case *semantic.CallExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 6 && c == 5:
							return jimInt
						case l == 8 && c == 5:
							return semantic.String
						}
					}
					return nil
				},
			},
		},
		{
			name: "function with polymorphic object parameter",
			script: `
foo = (r) => ({
    a: r.a,
    a2: r.a*r.a,
    b: r.b,
})
foo(r:{a:1,b:"hi"})
foo(r:{a:1.1,b:42.0})
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					tvA := semantic.NullableTvar{semantic.Tvar(12)}
					tvB := semantic.NullableTvar{semantic.Tvar(15)}

					r := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": tvA,
							"b": tvB,
						},
						semantic.LabelSet{"a", "b"},
						semantic.AllLabels(),
					)
					fooParams := map[string]semantic.PolyType{
						"r": r,
					}
					requiredR := semantic.LabelSet{"r"}
					fooOut := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  tvA,
							"a2": tvA,
							"b":  tvB,
						},
						nil,
						semantic.LabelSet{"a", "a2", "b"},
					)
					foo := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: fooParams,
						Required:   requiredR,
						Return:     fooOut,
					})

					obj1 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
							"b": semantic.String,
						},
						semantic.LabelSet{"a", "b"},
						semantic.LabelSet{"a", "b"},
					)
					params1 := map[string]semantic.PolyType{
						"r": obj1,
					}
					foo1 := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: params1,
						Required:   requiredR,
						Return: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a":  semantic.Int,
								"a2": semantic.Int,
								"b":  semantic.String,
							},
							nil,
							semantic.LabelSet{"a", "a2", "b"},
						),
					})
					obj2 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Float,
							"b": semantic.Float,
						},
						semantic.LabelSet{"a", "b"},
						semantic.LabelSet{"a", "b"},
					)
					params2 := map[string]semantic.PolyType{
						"r": obj2,
					}
					foo2 := semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
						Parameters: params2,
						Required:   requiredR,
						Return: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a":  semantic.Float,
								"a2": semantic.Float,
								"b":  semantic.Float,
							},
							nil,
							semantic.LabelSet{"a", "a2", "b"},
						),
					})

					out1 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  semantic.Int,
							"a2": semantic.Int,
							"b":  semantic.String,
						},
						nil,
						semantic.LabelSet{"a", "a2", "b"},
					)
					out2 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  semantic.Float,
							"a2": semantic.Float,
							"b":  semantic.Float,
						},
						nil,
						semantic.LabelSet{"a", "a2", "b"},
					)

					switch n := node.(type) {
					case *semantic.FunctionExpression:
						return foo
					case *semantic.FunctionParameter:
						return r
					case *semantic.FunctionBlock:
						return fooOut
					case *semantic.ObjectExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 2 && c == 8:
							return semantic.NewEmptyObjectPolyType()
						case l == 2 && c == 15:
							return fooOut
						case l == 7 && c == 5:
							return semantic.NewObjectPolyType(
								params1,
								nil,
								requiredR,
							)
						case l == 7 && c == 7:
							return obj1
						case l == 8 && c == 5:
							return semantic.NewObjectPolyType(
								params2,
								nil,
								requiredR,
							)
						case l == 8 && c == 7:
							return obj2
						}
					case *semantic.Property:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 3:
							return tvA
						case l == 4:
							return tvA
						case l == 5:
							return tvB
						case l == 7 && c == 5:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.Int,
									"b": semantic.String,
								},
								semantic.LabelSet{"a", "b"},
								semantic.LabelSet{"a", "b"},
							)
						case l == 7 && c == 8:
							return semantic.Int
						case l == 7 && c == 12:
							return semantic.String
						case l == 8 && c == 5:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.Float,
									"b": semantic.Float,
								},
								semantic.LabelSet{"a", "b"},
								semantic.LabelSet{"a", "b"},
							)
						case l == 8 && c == 8:
							return semantic.Float
						case l == 8 && c == 14:
							return semantic.Float
						}
					case *semantic.MemberExpression:
						switch n.Location().Start.Line {
						case 3, 4:
							return tvA
						case 5:
							return tvB
						}
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
						case 7:
							return out1
						case 8:
							return out2
						}
					case *semantic.BinaryExpression:
						return tvA
					case *semantic.IdentifierExpression:
						switch n.Location().Start.Line {
						case 3, 4, 5:
							return r
						case 7:
							return foo1
						case 8:
							return foo2
						}
					}
					return nil
				},
			},
		},
		{
			name: "object kind unification error",
			script: `
plus1 = (r={_value:1}) => r._value + 1
plus1(r:{_value: 2.0})
`,
			wantErr: errors.New(`type error 3:1-3:23: invalid record access "_value": int != float`),
		},
		{
			name: "generalize types",
			script: `
(x) => {
	y = x
	return y
}
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					tv := semantic.Tvar(3)
					switch node.(type) {
					case *semantic.FunctionBlock,
						*semantic.FunctionParameter,
						*semantic.Block,
						*semantic.IdentifierExpression,
						*semantic.ReturnStatement:
						return tv
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: map[string]semantic.PolyType{
								"x": tv,
							},
							Required: semantic.LabelSet{"x"},
							Return:   tv,
						})
					}
					return nil
				},
			},
		},
		{
			name: "occurs check",
			script: `
(f) => { return f(a:f) }
`,
			wantErr: errors.New(`type error 2:17-2:23: type var t3 occurs in (^a: t3) -> t11 creating a cycle`),
		},
		{
			name: "imports",
			script: `
import "foo"

foo.a
foo.b
`,
			importer: importer{packages: map[string]semantic.PackageType{
				"foo": semantic.PackageType{
					Name: "foo",
					Type: semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
							"b": semantic.Int,
							"c": semantic.String,
						},
						nil,
						semantic.LabelSet{"a", "b", "c"},
					),
				},
			}},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Location().Start.Line {
						case 4:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.Int,
									"b": semantic.Int,
									"c": semantic.String,
								},
								semantic.LabelSet{"a"},
								semantic.LabelSet{"a", "b", "c"},
							)
						case 5:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.Int,
									"b": semantic.Int,
									"c": semantic.String,
								},
								semantic.LabelSet{"b"},
								semantic.LabelSet{"a", "b", "c"},
							)
						}
					case *semantic.MemberExpression:
						return semantic.Int
					}
					return nil
				},
			},
		},
		{
			name: "imports pipe expression",
			script: `
import "foo"

foo.b
    |> foo.a()
`,
			importer: importer{packages: map[string]semantic.PackageType{
				"foo": semantic.PackageType{
					Name: "foo",
					Type: semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								Required:     semantic.LabelSet{"x"},
								Return:       semantic.Int,
								PipeArgument: "x",
							}),
							"b": semantic.Int,
						},
						semantic.LabelSet{"a", "b"},
						semantic.LabelSet{"a", "b"},
					),
				},
			}},
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Location().Start.Line {
						case 4, 5:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
										Parameters: map[string]semantic.PolyType{
											"x": semantic.Int,
										},
										Required:     semantic.LabelSet{"x"},
										Return:       semantic.Int,
										PipeArgument: "x",
									}),
									"b": semantic.Int,
								},
								semantic.LabelSet{"a", "b"},
								semantic.LabelSet{"a", "b"},
							)
						}
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(nil, nil, nil)
					case *semantic.CallExpression:
						return semantic.Int
					case *semantic.MemberExpression:
						switch n.Location().Start.Line {
						case 4:
							return semantic.Int
						case 5:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								Required:     semantic.LabelSet{"x"},
								Return:       semantic.Int,
								PipeArgument: "x",
							})
						}
					}
					return nil
				},
			},
		},
		{
			name:   "conditional expression",
			script: `if true then 3 else 30`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.ConditionalExpression:
						return semantic.Int
					case *semantic.IdentifierExpression:
						return semantic.Bool
					}
					return nil
				},
			},
		},
		{
			name:   "conditional infer branches",
			script: `(t, c, a) => if t then c else a`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// Type inference is able to deduce that the branches of the conditional
					// must have the same type, so parameters c and a must also have the same type.
					tv := semantic.Tvar(5)
					switch n := node.(type) {
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: map[string]semantic.PolyType{
								"t": semantic.Bool,
								"c": tv,
								"a": tv,
							},
							Required: semantic.LabelSet{"t", "c", "a"},
							Return:   tv,
						})
					case *semantic.FunctionBlock:
						return tv
					case *semantic.FunctionParameter:
						if n.Key.Name == "t" {
							return semantic.Bool
						} else {
							return tv
						}
					case *semantic.ConditionalExpression:
						return tv
					case *semantic.IdentifierExpression:
						if n.Name == "t" {
							return semantic.Bool
						} else {
							return tv
						}
					}
					return nil
				},
			},
		},
		{
			name:    "conditional branches must agree",
			script:  `if true then 0 else "foo"`,
			wantErr: errors.New(`type error 1:1-1:26: int != string`),
		},
		{
			name:    "conditional test must be bool",
			script:  `if 1 then 0.1 else 0.0`,
			wantErr: errors.New(`type error 1:4-1:5: int != bool`),
		},
		{
			name: "exists",
			script: `b = 1
a = exists b`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						if n.Name == "a" {
							return semantic.Bool
						}
						return semantic.Int
					case *semantic.UnaryExpression:
						return semantic.Bool
					}
					return nil
				},
			},
		},
		{
			name:   "conditional expression with exists",
			script: `if exists 1 then 3 else 30`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					switch node.(type) {
					case *semantic.ConditionalExpression:
						return semantic.Int
					case *semantic.UnaryExpression:
						return semantic.Bool
					}
					return nil
				},
			},
		},
		{
			name:   "function with exists",
			script: `fillNull = (v) => if exists v then v else "NULL"`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// Type inference is able to deduce that the branches of the conditional
					// must have the same type. The alternate is a StringLiteral, so v must be a string, and
					// the result of the conditional too.
					switch node.(type) {
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: map[string]semantic.PolyType{
								"v": semantic.String,
							},
							Required: semantic.LabelSet{"v"},
							Return:   semantic.String,
						})
					case *semantic.FunctionBlock:
						return semantic.String
					case *semantic.FunctionParameter:
						return semantic.String
					case *semantic.ConditionalExpression:
						return semantic.String
					case *semantic.IdentifierExpression:
						return semantic.String
					case *semantic.UnaryExpression:
						return semantic.Bool
					}
					return nil
				},
			},
		},
		{
			name: "retrieve nonexistant value",
			script: `r = {}
r.a`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// This will eventually resolve as null,
					// but it stays a NullableTvar until the real type
					// is retrieved.
					tv := semantic.NullableTvar{semantic.Tvar(2)}
					r := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": tv,
						},
						semantic.LabelSet{"a"},
						nil,
					)
					switch n := node.(type) {
					case *semantic.MemberExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return tv
							}
						}
					case *semantic.ObjectExpression:
						return semantic.NewEmptyObjectPolyType()
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return r
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "addition with nonexistant value",
			script: `r = {a: 1}
r.a + r.b`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					r := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
						},
						nil,
						semantic.LabelSet{"a"},
					)
					ra := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
						},
						semantic.LabelSet{"a"},
						semantic.LabelSet{"a"},
					)
					rb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
							"b": semantic.Int,
						},
						semantic.LabelSet{"b"},
						semantic.LabelSet{"a"},
					)
					switch n := node.(type) {
					case *semantic.BinaryExpression:
						return semantic.Int
					case *semantic.MemberExpression:
						return semantic.Int
					case *semantic.ObjectExpression:
						return r
					case *semantic.Property:
						switch n.Start.Line {
						case 1:
							switch n.Start.Column {
							case 6:
								return semantic.Int
							}
						}
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return ra
							case 7:
								return rb
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "comparison with nonexistant value",
			script: `r = {a: 1}
r.a < r.b`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// This will eventually resolve as null,
					// but it stays a NullableTvar until the real type
					// is retrieved.
					tv := semantic.NullableTvar{semantic.Tvar(7)}
					r := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
						},
						nil,
						semantic.LabelSet{"a"},
					)
					ra := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
						},
						semantic.LabelSet{"a"},
						semantic.LabelSet{"a"},
					)
					rb := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": semantic.Int,
							"b": tv,
						},
						semantic.LabelSet{"b"},
						semantic.LabelSet{"a"},
					)
					switch n := node.(type) {
					case *semantic.BinaryExpression:
						return semantic.Bool
					case *semantic.MemberExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return semantic.Int
							case 7:
								return tv
							}
						}
					case *semantic.ObjectExpression:
						return r
					case *semantic.Property:
						switch n.Start.Line {
						case 1:
							switch n.Start.Column {
							case 6:
								return semantic.Int
							}
						}
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return ra
							case 7:
								return rb
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "addition with only null values",
			script: `r = {}
r.a + r.b`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// This will eventually resolve as null,
					// but it stays a NullableTvar until the real type
					// is retrieved.
					tv := semantic.NullableTvar{semantic.Tvar(5)}
					r1 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": tv,
						},
						semantic.LabelSet{"a"},
						nil,
					)
					r2 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"b": tv,
						},
						semantic.LabelSet{"b"},
						nil,
					)
					switch n := node.(type) {
					case *semantic.BinaryExpression:
						return tv
					case *semantic.MemberExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return tv
							case 7:
								return tv
							}
						}
					case *semantic.ObjectExpression:
						return semantic.NewEmptyObjectPolyType()
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return r1
							case 7:
								return r2
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "comparison with only null values",
			script: `r = {}
r.a < r.b`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					// This will eventually resolve as null,
					// but it stays a NullableTvar until the real type
					// is retrieved.
					tv1 := semantic.NullableTvar{semantic.Tvar(3)}
					tv2 := semantic.NullableTvar{semantic.Tvar(5)}
					r1 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": tv1,
						},
						semantic.LabelSet{"a"},
						nil,
					)
					r2 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"b": tv2,
						},
						semantic.LabelSet{"b"},
						nil,
					)
					switch n := node.(type) {
					case *semantic.BinaryExpression:
						return semantic.Bool
					case *semantic.MemberExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return tv1
							case 7:
								return tv2
							}
						}
					case *semantic.ObjectExpression:
						return semantic.NewEmptyObjectPolyType()
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							switch n.Start.Column {
							case 1:
								return r1
							case 7:
								return r2
							}
						}
					}
					return nil
				},
			},
		},
		{
			name: "access nonexistant object",
			script: `r = {a: 1}
r.b.c + r.a
`,
			wantErr: errors.New("missing object properties (b)"),
		},
		{
			name: "nonexistant regex in object",
			script: `r = {a: "abc"}
		r.a =~ r.b
		`,
			wantErr: errors.New("missing object properties (b)"),
			skip:    "object cannot recognize that b is a regex and not nullable",
		},
		{
			name: "use variable returned from function",
			script: `
v = ((x) => x)(x:1)
v
`,
			skip: "variable loses return type information from the function",
		},
		{
			name: "same function can return different types",
			script: `
identity = (x) => x
a = identity(x:1)
b = identity(x:2.0)
{a:a, b:b}
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					x := semantic.Tvar(3)
					switch n := node.(type) {
					case *semantic.FunctionExpression:
						return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: map[string]semantic.PolyType{
								"x": x,
							},
							Required: semantic.LabelSet{"x"},
							Return:   x,
						})
					case *semantic.FunctionBlock:
						return x
					case *semantic.FunctionParameter:
						return x
					case *semantic.IdentifierExpression:
						switch n.Start.Line {
						case 2:
							return x
						case 3:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								Required: semantic.LabelSet{"x"},
								Return:   semantic.Int,
							})
						case 4:
							return semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
								Parameters: map[string]semantic.PolyType{
									"x": semantic.Float,
								},
								Required: semantic.LabelSet{"x"},
								Return:   semantic.Float,
							})
						case 5:
							switch n.Start.Column {
							case 4:
								return semantic.Tvar(24)
							case 9:
								return semantic.Tvar(25)
							}
						}
					case *semantic.CallExpression:
						switch n.Start.Line {
						case 3:
							return semantic.Int
						case 4:
							return semantic.Float
						}
					case *semantic.ObjectExpression:
						switch n.Start.Line {
						case 3:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"x": semantic.Int,
								},
								nil,
								semantic.LabelSet{"x"},
							)
						case 4:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"x": semantic.Float,
								},
								nil,
								semantic.LabelSet{"x"},
							)
						case 5:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"a": semantic.Tvar(24),
									"b": semantic.Tvar(25),
								},
								nil,
								semantic.LabelSet{"a", "b"},
							)
						}
					case *semantic.Property:
						switch n.Start.Line {
						case 3:
							return semantic.Int
						case 4:
							return semantic.Float
						case 5:
							switch n.Start.Column {
							case 2:
								return semantic.Tvar(24)
							case 7:
								return semantic.Tvar(25)
							}
						}
					}
					return nil
				},
			},
		},
		{
			name:   "return type conflicts with usage",
			script: `((x) => x)(x:1).a`,
			skip:   "type system doesn't identify this as an error",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			if tc.script != "" {
				pkg := parser.ParseSource(tc.script)
				if ast.Check(pkg) > 0 {
					t.Fatal(ast.GetError(pkg))
				}
				node, err := semantic.New(pkg)
				if err != nil {
					t.Fatal(err)
				}
				tc.node = node
			}

			// Add the true and false identifiers.
			tc.node = &semantic.Extern{
				Assignments: []*semantic.ExternalVariableAssignment{
					{
						Identifier: &semantic.Identifier{Name: "true"},
						ExternType: semantic.Bool,
					},
					{
						Identifier: &semantic.Identifier{Name: "false"},
						ExternType: semantic.Bool,
					},
				},
				Block: &semantic.ExternBlock{
					Node: tc.node,
				},
			}

			var wantSolution semantic.SolutionMap
			if tc.solution != nil {
				semantic.Walk(tc.solution, tc.node)
				wantSolution = tc.solution.Solution()
			}

			ts, err := semantic.InferTypes(tc.node, tc.importer)
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

			gotSolution := semantic.CreateSolutionMap(tc.node, ts)

			if want, got := len(wantSolution), len(gotSolution); got != want {
				t.Errorf("unexpected solution length want: %d got: %d", want, got)
			}
			wantNodes := make([]semantic.Node, 0, len(wantSolution))
			for n := range wantSolution {
				wantNodes = append(wantNodes, n)
			}
			semantic.SortNodes(wantNodes)
			for _, n := range wantNodes {
				want := wantSolution[n]
				got := gotSolution[n]
				if want == nil && got != nil {
					t.Errorf("unexpected type for node %T@%v, want: %v got: %v", n, n.Location(), want, got)
				} else if !want.Equal(got) {
					t.Errorf("unexpected type for node %T@%v, want: %v got: %v", n, n.Location(), want, got)
				}
			}
			gotNodes := make([]semantic.Node, 0, len(gotSolution))
			for n := range gotSolution {
				gotNodes = append(gotNodes, n)
			}
			semantic.SortNodes(gotNodes)
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

func TestInferTypes_TypeOf_Null(t *testing.T) {
	expr := &semantic.MemberExpression{
		Object:   &semantic.IdentifierExpression{Name: "r"},
		Property: "a",
	}
	node := &semantic.File{
		Body: []semantic.Statement{
			&semantic.NativeVariableAssignment{
				Identifier: &semantic.Identifier{Name: "r"},
				Init: &semantic.ObjectExpression{
					Properties: []*semantic.Property{},
				},
			},
			&semantic.ExpressionStatement{
				Expression: expr,
			},
		},
	}
	ts, err := semantic.InferTypes(node, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typ, err := ts.TypeOf(expr)
	if err != nil {
		t.Errorf("could not resolve type of statement: %v", err)
	} else if got, want := typ, semantic.Nil; got != want {
		t.Errorf("unexpected type for statement, want: %v got: %v", want, got)
	}
}

type SolutionVisitor interface {
	semantic.Visitor
	Solution() semantic.SolutionMap
}

type solutionVisitor struct {
	f        func(node semantic.Node) semantic.PolyType
	solution semantic.SolutionMap
}

func (v *solutionVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.solution == nil {
		v.solution = make(semantic.SolutionMap)
	}
	// Handle literals here
	if l, ok := node.(semantic.Literal); ok {
		var typ semantic.PolyType
		switch l.(type) {
		case *semantic.StringLiteral:
			typ = semantic.String
		case *semantic.IntegerLiteral:
			typ = semantic.Int
		case *semantic.UnsignedIntegerLiteral:
			typ = semantic.UInt
		case *semantic.FloatLiteral:
			typ = semantic.Float
		case *semantic.BooleanLiteral:
			typ = semantic.Bool
		case *semantic.DateTimeLiteral:
			typ = semantic.Time
		case *semantic.DurationLiteral:
			typ = semantic.Duration
		case *semantic.RegexpLiteral:
			typ = semantic.Regexp
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

func (v *solutionVisitor) Solution() semantic.SolutionMap {
	return v.solution
}

type importer struct {
	packages map[string]semantic.PackageType
}

func (imp importer) Import(path string) (semantic.PackageType, bool) {
	pkg, ok := imp.packages[path]
	return pkg, ok
}
