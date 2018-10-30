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
			wantErr: errors.New(`type error 3:1-3:7: int != bool`),
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
						return semantic.NewFunctionPolyType(
							params,
							required,
							semantic.Int,
						)
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
						return semantic.NewFunctionPolyType(
							params,
							required,
							semantic.Int,
						)
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"b": semantic.Int,
							},
							semantic.EmptyLabelSet(),
							semantic.LabelSet{"b"},
						)
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
						return semantic.NewFunctionPolyType(
							params,
							required,
							semantic.Int,
						)
					case *semantic.ObjectExpression:
						return semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a": semantic.Int,
							},
							semantic.EmptyLabelSet(),
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
					ft := semantic.NewFunctionPolyType(
						params,
						required,
						semantic.Int,
					)
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
								semantic.EmptyLabelSet(),
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
					t4 := semantic.Tvar(4)
					params := map[string]semantic.PolyType{
						"x": t4,
					}
					required := semantic.LabelSet{"x"}
					out := t4
					ft := semantic.NewFunctionPolyType(
						params,
						required,
						out,
					)

					paramsInt := map[string]semantic.PolyType{
						"x": semantic.Int,
					}
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(
						paramsInt,
						required,
						outInt,
					)

					paramsF := map[string]semantic.PolyType{
						"x": ftInt,
					}
					outF := semantic.NewFunctionPolyType(
						paramsInt,
						required,
						outInt,
					)
					ftF := semantic.NewFunctionPolyType(
						paramsF,
						required,
						outF,
					)
					switch n := node.(type) {
					case *semantic.CallExpression:
						switch n.Location().Start.Column {
						case 1:
							return outF
						case 21:
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
									semantic.EmptyLabelSet(),
									required,
								)
							case 22:
								return semantic.NewObjectPolyType(
									paramsInt,
									semantic.EmptyLabelSet(),
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
					case *semantic.IdentifierExpression:
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
					t9 := semantic.Tvar(9)
					params := map[string]semantic.PolyType{
						"a": t9,
						"b": t9,
					}
					requiredAB := semantic.LabelSet{"a", "b"}
					out := t9
					ft := semantic.NewFunctionPolyType(
						params,
						requiredAB,
						out,
					)
					paramsInt := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					outInt := semantic.Int
					ftInt := semantic.NewFunctionPolyType(
						paramsInt,
						requiredAB,
						outInt,
					)
					paramsR := map[string]semantic.PolyType{
						"r": semantic.Int,
					}
					requiredR := semantic.LabelSet{"r"}
					ftR := semantic.NewFunctionPolyType(
						paramsR,
						requiredR,
						semantic.Int,
					)
					switch n := node.(type) {
					case *semantic.IdentifierExpression:
						switch n.Name {
						case "a":
							return t9
						case "b":
							return t9
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
							return t9
						}
					case *semantic.FunctionParameter:
						switch n.Key.Name {
						case "a":
							return t9
						case "b":
							return t9
						case "r":
							return outInt
						}
					case *semantic.Property:
						return outInt
					case *semantic.BinaryExpression:
						return out
					case *semantic.BlockStatement,
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
								semantic.EmptyLabelSet(),
								requiredAB,
							)
						}
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
					t30 := semantic.Tvar(30)
					paramsCall := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					requiredAB := semantic.LabelSet{"a", "b"}
					outCall := t30
					call := semantic.NewFunctionPolyType(
						paramsCall,
						requiredAB,
						outCall,
					)

					paramsFoo := map[string]semantic.PolyType{
						"f": call,
					}
					requiredF := semantic.LabelSet{"f"}
					outFoo := outCall
					foo := semantic.NewFunctionPolyType(
						paramsFoo,
						requiredF,
						outFoo,
					)

					paramsCallInt := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
					}
					outCallInt := semantic.Int

					callInt := semantic.NewFunctionPolyType(
						paramsCallInt,
						requiredAB,
						outCallInt,
					)
					paramsFooInt := map[string]semantic.PolyType{
						"f": callInt,
					}
					outFooInt := outCallInt
					fooInt := semantic.NewFunctionPolyType(
						paramsFooInt,
						requiredF,
						outFooInt,
					)

					paramsAdd := map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Int,
						"c": semantic.Int,
					}
					outAdd := semantic.Int
					add := semantic.NewFunctionPolyType(
						paramsAdd,
						requiredAB,
						outAdd,
					)

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
									semantic.EmptyLabelSet(),
									semantic.EmptyLabelSet(),
								)
							case 16:
								return semantic.NewObjectPolyType(
									paramsCallInt,
									semantic.EmptyLabelSet(),
									requiredAB,
								)
							}
						case 3:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"c": semantic.Int,
								},
								semantic.EmptyLabelSet(),
								semantic.LabelSet{"c"},
							)
						case 4:
							return semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"f": semantic.NewFunctionPolyType(
										map[string]semantic.PolyType{
											"a": semantic.Int,
											"b": semantic.Int,
											"c": semantic.Int,
										},
										requiredAB,
										semantic.Int,
									),
								},
								semantic.EmptyLabelSet(),
								requiredF,
							)
						}
						return semantic.Int
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
device = {name: 42, lat:28.25892, lon: 15.62234}

name = (p) => p.name

name(p:jim)
name(p:jane)
name(p:device)
`,
			solution: &solutionVisitor{
				f: func(node semantic.Node) semantic.PolyType {
					jim := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name":   semantic.String,
							"age":    semantic.Int,
							"weight": semantic.Float,
						},
						semantic.EmptyLabelSet(),
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
						semantic.EmptyLabelSet(),
						semantic.LabelSet{"p"},
					)
					jane := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.String,
							"age":  semantic.Int,
						},
						semantic.EmptyLabelSet(),
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
						semantic.EmptyLabelSet(),
						semantic.LabelSet{"p"},
					)
					device := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": semantic.Int,
							"lat":  semantic.Float,
							"lon":  semantic.Float,
						},
						semantic.EmptyLabelSet(),
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
						semantic.EmptyLabelSet(),
						semantic.LabelSet{"p"},
					)

					tv := semantic.Tvar(41)
					p := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"name": tv,
						},
						semantic.LabelSet{"name"},
						semantic.AllLabels,
					)
					name := semantic.NewFunctionPolyType(
						map[string]semantic.PolyType{
							"p": p,
						},
						semantic.LabelSet{"p"},
						tv,
					)
					nameCallJim := semantic.NewFunctionPolyType(
						map[string]semantic.PolyType{
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
						semantic.LabelSet{"p"},
						semantic.String,
					)
					nameCallJane := semantic.NewFunctionPolyType(
						map[string]semantic.PolyType{
							"p": semantic.NewObjectPolyType(
								map[string]semantic.PolyType{
									"name": semantic.String,
									"age":  semantic.Int,
								},
								semantic.LabelSet{"name"},
								semantic.LabelSet{"name", "age"},
							),
						},
						semantic.LabelSet{"p"},
						semantic.String,
					)
					nameCallDevice := semantic.NewFunctionPolyType(
						map[string]semantic.PolyType{
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
						semantic.LabelSet{"p"},
						semantic.Int,
					)

					nameJim := semantic.String
					nameJane := semantic.String
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
						case l == 4 && c == 11:
							return semantic.Int
						case l == 4 && c == 21:
							return semantic.Float
						case l == 4 && c == 35:
							return semantic.Float
						case l == 8:
							return jimCall
						case l == 9:
							return janeCall
						case l == 10:
							return deviceCall
						}
					case *semantic.ObjectExpression:
						switch n.Location().Start.Line {
						case 2:
							return jim
						case 3:
							return jane
						case 4:
							return device
						case 6:
							return semantic.NewEmptyObjectPolyType()
						case 8:
							return pJim
						case 9:
							return pJane
						case 10:
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
						case 8:
							return nameJim
						case 9:
							return nameJane
						case 10:
							return nameDevice
						}
					case *semantic.IdentifierExpression:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 6:
							return p
						case l == 8 && c == 1:
							return nameCallJim
						case l == 8 && c == 8:
							return jimCall
						case l == 9 && c == 1:
							return nameCallJane
						case l == 9 && c == 8:
							return janeCall
						case l == 10 && c == 1:
							return nameCallDevice
						case l == 10 && c == 8:
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
					t38 := semantic.Tvar(38)
					t39 := semantic.Tvar(39)

					r := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a": t38,
							"b": t39,
						},
						semantic.LabelSet{"a", "b"},
						semantic.AllLabels,
					)
					fooParams := map[string]semantic.PolyType{
						"r": r,
					}
					requiredR := semantic.LabelSet{"r"}
					fooOut := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  t38,
							"a2": t38,
							"b":  t39,
						},
						semantic.EmptyLabelSet(),
						semantic.LabelSet{"a", "a2", "b"},
					)
					foo := semantic.NewFunctionPolyType(
						fooParams,
						requiredR,
						fooOut,
					)

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
					foo1 := semantic.NewFunctionPolyType(
						params1,
						requiredR,
						semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a":  semantic.Int,
								"a2": semantic.Int,
								"b":  semantic.String,
							},
							semantic.EmptyLabelSet(),
							semantic.LabelSet{"a", "a2", "b"},
						),
					)
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
					foo2 := semantic.NewFunctionPolyType(
						params2,
						requiredR,
						semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a":  semantic.Float,
								"a2": semantic.Float,
								"b":  semantic.Float,
							},
							semantic.EmptyLabelSet(),
							semantic.LabelSet{"a", "a2", "b"},
						),
					)

					out1 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  semantic.Int,
							"a2": semantic.Int,
							"b":  semantic.String,
						},
						semantic.EmptyLabelSet(),
						semantic.LabelSet{"a", "a2", "b"},
					)
					out2 := semantic.NewObjectPolyType(
						map[string]semantic.PolyType{
							"a":  semantic.Float,
							"a2": semantic.Float,
							"b":  semantic.Float,
						},
						semantic.EmptyLabelSet(),
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
						case l == 2:
							return semantic.NewEmptyObjectPolyType()
						case l == 3:
							return fooOut
						case l == 7 && c == 5:
							return semantic.NewObjectPolyType(
								params1,
								semantic.EmptyLabelSet(),
								requiredR,
							)
						case l == 7 && c == 8:
							return obj1
						case l == 8 && c == 5:
							return semantic.NewObjectPolyType(
								params2,
								semantic.EmptyLabelSet(),
								requiredR,
							)
						case l == 8 && c == 8:
							return obj2
						}
					case *semantic.Property:
						switch l, c := n.Location().Start.Line, n.Location().Start.Column; {
						case l == 3:
							return t38
						case l == 4:
							return t38
						case l == 5:
							return t39
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
							return t38
						case 5:
							return t39
						}
					case *semantic.CallExpression:
						switch n.Location().Start.Line {
						case 7:
							return out1
						case 8:
							return out2
						}
					case *semantic.BinaryExpression:
						return t38
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

			ts, err := semantic.InferTypes(tc.node)
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

			// Read all types from node
			v := &typeVisitor{
				typeSolution: ts,
			}
			semantic.Walk(v, tc.node)
			gotSolution := v.Solution()

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

func (v *solutionVisitor) Solution() SolutionMap {
	return v.solution
}

type typeVisitor struct {
	typeSolution semantic.TypeSolution
	solution     SolutionMap
}

func (v *typeVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.solution == nil {
		v.solution = make(SolutionMap)
	}
	t, err := v.typeSolution.PolyTypeOf(node)
	if err != nil {
		return nil
	}
	if t != nil {
		v.solution[node] = t
	}
	return v
}

func (v *typeVisitor) Done(semantic.Node) {}

func (v *typeVisitor) Solution() SolutionMap {
	return v.solution
}

func sortNodes(nodes []semantic.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Location().Less(nodes[j].Location())
	})
}
