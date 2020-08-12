package ast_test

import (
	"encoding/json"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/asttest"
)

func TestJSONMarshal(t *testing.T) {
	testCases := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			name: "string interpolation",
			node: &ast.StringExpression{
				Parts: []ast.StringExpressionPart{
					&ast.TextPart{
						Value: "a = ",
					},
					&ast.InterpolatedPart{
						Expression: &ast.Identifier{
							Name: "a",
						},
					},
				},
			},
			want: `{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}`,
		},
		{
			name: "paren expression",
			node: &ast.ParenExpression{
				Expression: &ast.StringExpression{
					Parts: []ast.StringExpressionPart{
						&ast.TextPart{
							Value: "a = ",
						},
						&ast.InterpolatedPart{
							Expression: &ast.Identifier{
								Name: "a",
							},
						},
					},
				},
			},
			want: `{"type":"ParenExpression","expression":{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}}`,
		},
		{
			name: "simple package",
			node: &ast.Package{
				Package: "foo",
			},
			want: `{"type":"Package","package":"foo","files":null}`,
		},
		{
			name: "package path",
			node: &ast.Package{
				Path:    "bar/foo",
				Package: "foo",
			},
			want: `{"type":"Package","path":"bar/foo","package":"foo","files":null}`,
		},
		{
			name: "simple file",
			node: &ast.File{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"File","package":null,"imports":null,"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "file",
			node: &ast.File{
				Metadata: "parser-type=none",
				Package: &ast.PackageClause{
					Name: &ast.Identifier{Name: "foo"},
				},
				Imports: []*ast.ImportDeclaration{{
					As:   &ast.Identifier{Name: "b"},
					Path: &ast.StringLiteral{Value: "path/bar"},
				}},
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"File","metadata":"parser-type=none","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "block",
			node: &ast.Block{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "expression statement",
			node: &ast.ExpressionStatement{
				Expression: &ast.StringLiteral{Value: "hello"},
			},
			want: `{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "return statement",
			node: &ast.ReturnStatement{
				Argument: &ast.StringLiteral{Value: "hello"},
			},
			want: `{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "option statement",
			node: &ast.OptionStatement{
				Assignment: &ast.VariableAssignment{
					ID: &ast.Identifier{Name: "task"},
					Init: &ast.ObjectExpression{
						Properties: []*ast.Property{
							{
								Key:   &ast.Identifier{Name: "name"},
								Value: &ast.StringLiteral{Value: "foo"},
							},
							{
								Key: &ast.Identifier{Name: "every"},
								Value: &ast.DurationLiteral{
									Values: []ast.Duration{
										{
											Magnitude: 1,
											Unit:      "h",
										},
									},
								},
							},
						},
					},
				},
			},
			want: `{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}`,
		},
		{
			name: "builtin statement",
			node: &ast.BuiltinStatement{
				ID: &ast.Identifier{Name: "task"},
			},
			want: `{"type":"BuiltinStatement","id":{"name":"task"}}`,
		},
		{
			name: "NamedType",
			node: &ast.NamedType{
				BaseNode: ast.BaseNode{},
				ID: &ast.Identifier{
					BaseNode: ast.BaseNode{},
					Name:     "int",
				},
			},
			want: `{"type":"NamedType","id":{"type":"Identifier","name":"int"}}`,
		},
		{
			name: "TvarType",
			node: &ast.TvarType{
				BaseNode: ast.BaseNode{},
				ID: &ast.Identifier{
					BaseNode: ast.BaseNode{},
					Name:     "A",
				},
			},
			want: `{"type":"TvarType","id":{"type":"Identifier","name":"A"}}`,
		},
		{
			name: "ArrayType",
			node: &ast.ArrayType{
				BaseNode: ast.BaseNode{},
				ElementType: &ast.RecordType{
					BaseNode: ast.BaseNode{},
					Properties: []*ast.PropertyType{
						{
							BaseNode: ast.BaseNode{},
							Name: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "A",
							},
							Ty: &ast.NamedType{
								BaseNode: ast.BaseNode{},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{},
									Name:     "int",
								},
							},
						},
					},
					Tvar: &ast.Identifier{
						BaseNode: ast.BaseNode{},
						Name:     "A",
					},
				},
			},
			want: `{"type":"ArrayType","elementType":{"type":"RecordType","properties":[{"type":"PropertyType","id":{"type":"Identifier","name":"A"},"ty":{"type":"NamedType","id":{"type":"Identifier","name":"int"}}}],"tvar":{"type":"Identifier","name":"A"}}}`,
		},
		{
			name: "RecordType",
			node: &ast.RecordType{
				BaseNode: ast.BaseNode{},
				Properties: []*ast.PropertyType{
					{
						BaseNode: ast.BaseNode{},
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "A",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "int",
							},
						},
					},
				},
				Tvar: &ast.Identifier{
					BaseNode: ast.BaseNode{},
					Name:     "A",
				},
			},
			want: `{"type":"RecordType","properties":[{"type":"PropertyType","id":{"type":"Identifier","name":"A"},"ty":{"type":"NamedType","id":{"type":"Identifier","name":"int"}}}],"tvar":{"type":"Identifier","name":"A"}}`,
		},
		{
			name: "FunctionType",
			node: &ast.FunctionType{
				BaseNode: ast.BaseNode{},
				Parameters: []*ast.ParameterType{
					{
						BaseNode: ast.BaseNode{},
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "A",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "int",
							},
						},
						Kind: 0},
					{
						BaseNode: ast.BaseNode{},
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "B",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "uint",
							},
						},
						Kind: 1},
					{
						BaseNode: ast.BaseNode{},
						Name: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "C",
						},
						Ty: &ast.NamedType{
							BaseNode: ast.BaseNode{},
							ID: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "string",
							},
						},
						Kind: 2},
				},
				Return: &ast.TvarType{
					BaseNode: ast.BaseNode{},
					ID: &ast.Identifier{
						BaseNode: ast.BaseNode{},
						Name:     "A",
					},
				},
			},
			want: `{"type":"FunctionType","parameters":[{"type":"ParameterType","name":{"type":"Identifier","name":"A"},"ty":{"type":"NamedType","id":{"type":"Identifier","name":"int"}},"kind":0},{"type":"ParameterType","name":{"type":"Identifier","name":"B"},"ty":{"type":"NamedType","id":{"type":"Identifier","name":"uint"}},"kind":1},{"type":"ParameterType","name":{"type":"Identifier","name":"C"},"ty":{"type":"NamedType","id":{"type":"Identifier","name":"string"}},"kind":2}],"return":{"type":"TvarType","id":{"type":"Identifier","name":"A"}}}`,
		},
		{
			name: "TypeExpression Test",
			node: &ast.TypeExpression{
				BaseNode: ast.BaseNode{},
				Ty: &ast.FunctionType{
					BaseNode: ast.BaseNode{},
					Parameters: []*ast.ParameterType{
						{
							BaseNode: ast.BaseNode{},
							Name: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "a",
							},
							Ty: &ast.TvarType{
								BaseNode: ast.BaseNode{},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{},
									Name:     "T",
								},
							},
							Kind: 0,
						},
						{
							BaseNode: ast.BaseNode{},
							Name: &ast.Identifier{
								BaseNode: ast.BaseNode{},
								Name:     "b",
							},
							Ty: &ast.TvarType{
								BaseNode: ast.BaseNode{},
								ID: &ast.Identifier{
									BaseNode: ast.BaseNode{},
									Name:     "T",
								},
							},
							Kind: 0,
						},
					},
					Return: &ast.TvarType{
						BaseNode: ast.BaseNode{},
						ID: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "T",
						},
					},
				},
				Constraints: []*ast.TypeConstraint{
					{
						BaseNode: ast.BaseNode{},
						Tvar: &ast.Identifier{
							BaseNode: ast.BaseNode{},
							Name:     "T",
						},
						Kinds: []*ast.Identifier{
							{
								BaseNode: ast.BaseNode{},
								Name:     "Addable",
							},
							{
								BaseNode: ast.BaseNode{},
								Name:     "Divisible",
							},
						},
					},
				},
			},
			want: `{"type":"TypeExpression","ty":{"type":"FunctionType","parameters":[{"type":"ParameterType","name":{"type":"Identifier","name":"a"},"ty":{"type":"TvarType","id":{"type":"Identifier","name":"T"}},"kind":0},{"type":"ParameterType","name":{"type":"Identifier","name":"b"},"ty":{"type":"TvarType","id":{"type":"Identifier","name":"T"}},"kind":0}],"return":{"type":"TvarType","id":{"type":"Identifier","name":"T"}}},"constraints":[{"type":"TypeConstraint","tvar":{"type":"Identifier","name":"T"},"kinds":[{"type":"Identifier","name":"Addable"},{"type":"Identifier","name":"Divisible"}]}]}`,
		},
		{
			name: "test statement",
			node: &ast.TestStatement{
				Assignment: &ast.VariableAssignment{
					ID: &ast.Identifier{Name: "mean"},
					Init: &ast.ObjectExpression{
						Properties: []*ast.Property{
							{
								Key: &ast.Identifier{
									Name: "want",
								},
								Value: &ast.IntegerLiteral{
									Value: 0,
								},
							},
							{
								Key: &ast.Identifier{
									Name: "got",
								},
								Value: &ast.IntegerLiteral{
									Value: 0,
								},
							},
						},
					},
				},
			},
			want: `{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}`,
		},
		{
			name: "qualified option statement",
			node: &ast.OptionStatement{
				Assignment: &ast.MemberAssignment{
					Member: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "alert",
						},
						Property: &ast.Identifier{
							Name: "state",
						},
					},
					Init: &ast.StringLiteral{
						Value: "Warning",
					},
				},
			},
			want: `{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}`,
		},
		{
			name: "variable assignment",
			node: &ast.VariableAssignment{
				ID:   &ast.Identifier{Name: "a"},
				Init: &ast.StringLiteral{Value: "hello"},
			},
			want: `{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "call expression",
			node: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "a"},
				Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
			},
			want: `{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}`,
		},
		{
			name: "pipe expression",
			node: &ast.PipeExpression{
				Argument: &ast.Identifier{Name: "a"},
				Call: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "a"},
					Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
				},
			},
			want: `{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}`,
		},
		{
			name: "member expression with identifier",
			node: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "a"},
				Property: &ast.Identifier{Name: "b"},
			},
			want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}`,
		},
		{
			name: "member expression with string literal",
			node: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "a"},
				Property: &ast.StringLiteral{Value: "b"},
			},
			want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}`,
		},
		{
			name: "index expression",
			node: &ast.IndexExpression{
				Array: &ast.Identifier{Name: "a"},
				Index: &ast.IntegerLiteral{Value: 3},
			},
			want: `{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}`,
		},
		{
			name: "arrow function expression",
			node: &ast.FunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "a"}}},
				Body:   &ast.StringLiteral{Value: "hello"},
			},
			want: `{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "binary expression",
			node: &ast.BinaryExpression{
				Operator: ast.AdditionOperator,
				Left:     &ast.StringLiteral{Value: "hello"},
				Right:    &ast.StringLiteral{Value: "world"},
			},
			want: `{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}`,
		},
		{
			name: "unary expression",
			node: &ast.UnaryExpression{
				Operator: ast.NotOperator,
				Argument: &ast.BooleanLiteral{Value: true},
			},
			want: `{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}`,
		},
		{
			name: "logical expression",
			node: &ast.LogicalExpression{
				Operator: ast.OrOperator,
				Left:     &ast.BooleanLiteral{Value: false},
				Right:    &ast.BooleanLiteral{Value: true},
			},
			want: `{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}`,
		},
		{
			name: "array expression",
			node: &ast.ArrayExpression{
				Elements: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
			},
			want: `{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}`,
		},
		{
			name: "object expression",
			node: &ast.ObjectExpression{
				Properties: []*ast.Property{{
					Key:   &ast.Identifier{Name: "a"},
					Value: &ast.StringLiteral{Value: "hello"},
				}},
			},
			want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "object expression with string literal key",
			node: &ast.ObjectExpression{
				Properties: []*ast.Property{{
					Key:   &ast.StringLiteral{Value: "a"},
					Value: &ast.StringLiteral{Value: "hello"},
				}},
			},
			want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "object expression implicit keys",
			node: &ast.ObjectExpression{
				Properties: []*ast.Property{{
					Key: &ast.Identifier{Name: "a"},
				}},
			},
			want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}`,
		},
		{
			name: "conditional expression",
			node: &ast.ConditionalExpression{
				Test:       &ast.BooleanLiteral{Value: true},
				Alternate:  &ast.StringLiteral{Value: "false"},
				Consequent: &ast.StringLiteral{Value: "true"},
			},
			want: `{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}`,
		},
		{
			name: "property",
			node: &ast.Property{
				Key:   &ast.Identifier{Name: "a"},
				Value: &ast.StringLiteral{Value: "hello"},
			},
			want: `{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "identifier",
			node: &ast.Identifier{
				Name: "a",
			},
			want: `{"type":"Identifier","name":"a"}`,
		},
		{
			name: "string literal",
			node: &ast.StringLiteral{
				Value: "hello",
			},
			want: `{"type":"StringLiteral","value":"hello"}`,
		},
		{
			name: "boolean literal",
			node: &ast.BooleanLiteral{
				Value: true,
			},
			want: `{"type":"BooleanLiteral","value":true}`,
		},
		{
			name: "float literal",
			node: &ast.FloatLiteral{
				Value: 42.1,
			},
			want: `{"type":"FloatLiteral","value":42.1}`,
		},
		{
			name: "integer literal",
			node: &ast.IntegerLiteral{
				Value: math.MaxInt64,
			},
			want: `{"type":"IntegerLiteral","value":"9223372036854775807"}`,
		},
		{
			name: "unsigned integer literal",
			node: &ast.UnsignedIntegerLiteral{
				Value: math.MaxUint64,
			},
			want: `{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}`,
		},
		{
			name: "regexp literal",
			node: &ast.RegexpLiteral{
				Value: regexp.MustCompile(`.*`),
			},
			want: `{"type":"RegexpLiteral","value":".*"}`,
		},
		{
			name: "duration literal",
			node: &ast.DurationLiteral{
				Values: []ast.Duration{
					{
						Magnitude: 1,
						Unit:      "h",
					},
					{
						Magnitude: 1,
						Unit:      "h",
					},
				},
			},
			want: `{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}`,
		},
		{
			name: "datetime literal",
			node: &ast.DateTimeLiteral{
				Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
			},
			want: `{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008Z"}`,
		},
		{
			name: "object expression with source locations and errors",
			node: &ast.ObjectExpression{
				BaseNode: ast.BaseNode{
					Loc: &ast.SourceLocation{
						File:   "foo.flux",
						Start:  ast.Position{Line: 1, Column: 1},
						End:    ast.Position{Line: 1, Column: 13},
						Source: "{a: \"hello\"}",
					},
				},
				Properties: []*ast.Property{{
					BaseNode: ast.BaseNode{
						Loc: &ast.SourceLocation{
							File:   "foo.flux",
							Start:  ast.Position{Line: 1, Column: 2},
							End:    ast.Position{Line: 1, Column: 12},
							Source: "a: \"hello\"",
						},
						Errors: []ast.Error{{Msg: "an error"}},
					},
					Key: &ast.Identifier{BaseNode: ast.BaseNode{
						Loc: &ast.SourceLocation{
							File:   "foo.flux",
							Start:  ast.Position{Line: 1, Column: 2},
							End:    ast.Position{Line: 1, Column: 3},
							Source: "a",
						},
					},
						Name: "a",
					},
					Value: &ast.StringLiteral{BaseNode: ast.BaseNode{
						Loc: &ast.SourceLocation{
							File:   "foo.flux",
							Start:  ast.Position{Line: 1, Column: 5},
							End:    ast.Position{Line: 1, Column: 12},
							Source: "\"hello\"",
						},
						Errors: []ast.Error{{Msg: "an error"}, {Msg: "another error"}},
					},
						Value: "hello",
					},
				}},
			},
			want: `{"type":"ObjectExpression","location":{"file":"foo.flux","start":{"line":1,"column":1},"end":{"line":1,"column":13},"source":"{a: \"hello\"}"},"properties":[{"type":"Property","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":12},"source":"a: \"hello\""},"errors":[{"msg":"an error"}],"key":{"type":"Identifier","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":3},"source":"a"},"name":"a"},"value":{"type":"StringLiteral","location":{"file":"foo.flux","start":{"line":1,"column":5},"end":{"line":1,"column":12},"source":"\"hello\""},"errors":[{"msg":"an error"},{"msg":"another error"}],"value":"hello"}}]}`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.node)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(data); got != tc.want {
				t.Errorf("unexpected json data:\nwant:%s\ngot: %s\n", tc.want, got)
			}
			node, err := ast.UnmarshalNode(data)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.node, node, asttest.CompareOptions...) {
				t.Errorf("unexpected node after unmarshalling: -want/+got:\n%s", cmp.Diff(tc.node, node, asttest.CompareOptions...))
			}
		})
	}
}
