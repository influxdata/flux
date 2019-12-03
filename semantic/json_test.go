package semantic_test

import (
	"encoding/json"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
)

func TestJSONMarshal(t *testing.T) {
	testCases := []struct {
		name string
		node semantic.Node
		want string
	}{
		{
			name: "string interpolation",
			node: &semantic.StringExpression{
				Parts: []semantic.StringExpressionPart{
					&semantic.TextPart{
						Value: "a + b = ",
					},
					&semantic.InterpolatedPart{
						Expression: &semantic.BinaryExpression{
							Left: &semantic.IdentifierExpression{
								Name: "a",
							},
							Right: &semantic.IdentifierExpression{
								Name: "b",
							},
							Operator: ast.AdditionOperator,
						},
					},
				},
			},
			want: `{"type":"StringExpression","parts":[{"type":"TextPart","value":"a + b = "},{"type":"InterpolatedPart","expression":{"type":"BinaryExpression","operator":"+","left":{"type":"IdentifierExpression","name":"a"},"right":{"type":"IdentifierExpression","name":"b"}}}]}`,
		},
		{
			name: "simple package",
			node: &semantic.Package{
				Package: "main",
			},
			want: `{"type":"Package","package":"main","files":null}`,
		},
		{
			name: "simple file",
			node: &semantic.File{
				Body: []semantic.Statement{
					&semantic.ExpressionStatement{
						Expression: &semantic.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"File","package":null,"imports":null,"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "file",
			node: &semantic.File{
				Package: &semantic.PackageClause{
					Name: &semantic.Identifier{Name: "foo"},
				},
				Body: []semantic.Statement{
					&semantic.ExpressionStatement{
						Expression: &semantic.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":null,"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "block",
			node: &semantic.Block{
				Body: []semantic.Statement{
					&semantic.ExpressionStatement{
						Expression: &semantic.StringLiteral{Value: "hello"},
					},
				},
			},
			want: `{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "option statement",
			node: &semantic.OptionStatement{
				Assignment: &semantic.NativeVariableAssignment{
					Identifier: &semantic.Identifier{Name: "task"},
					Init: &semantic.ObjectExpression{
						Properties: []*semantic.Property{
							{
								Key:   &semantic.Identifier{Name: "name"},
								Value: &semantic.StringLiteral{Value: "foo"},
							},
							{
								Key: &semantic.Identifier{Name: "every"},
								Value: &semantic.DurationLiteral{
									Values: []ast.Duration{
										{Magnitude: 1, Unit: ast.HourUnit},
									},
								},
							},
							{
								Key: &semantic.Identifier{Name: "delay"},
								Value: &semantic.DurationLiteral{
									Values: []ast.Duration{
										{Magnitude: 10, Unit: ast.MinuteUnit},
									},
								},
							},
							{
								Key:   &semantic.Identifier{Name: "cron"},
								Value: &semantic.StringLiteral{Value: "0 2 * * *"},
							},
							{
								Key:   &semantic.Identifier{Name: "retry"},
								Value: &semantic.IntegerLiteral{Value: 5},
							},
						},
					},
				},
			},
			want: `{"type":"OptionStatement","assignment":{"type":"NativeVariableAssignment","identifier":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}},{"type":"Property","key":{"type":"Identifier","name":"delay"},"value":{"type":"DurationLiteral","values":[{"magnitude":10,"unit":"m"}]}},{"type":"Property","key":{"type":"Identifier","name":"cron"},"value":{"type":"StringLiteral","value":"0 2 * * *"}},{"type":"Property","key":{"type":"Identifier","name":"retry"},"value":{"type":"IntegerLiteral","value":"5"}}]}}}`,
		},
		{
			name: "qualified option statement",
			node: &semantic.OptionStatement{
				Assignment: &semantic.MemberAssignment{
					Member: &semantic.MemberExpression{
						Object: &semantic.IdentifierExpression{
							Name: "alert",
						},
						Property: "state",
					},
					Init: &semantic.StringLiteral{
						Value: "Warning",
					},
				},
			},
			want: `{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"IdentifierExpression","name":"alert"},"property":"state"},"init":{"type":"StringLiteral","value":"Warning"}}}`,
		},
		{
			name: "test statement",
			node: &semantic.TestStatement{
				Assignment: &semantic.NativeVariableAssignment{
					Identifier: &semantic.Identifier{Name: "mean"},
					Init: &semantic.ObjectExpression{
						Properties: []*semantic.Property{
							{
								Key: &semantic.Identifier{
									Name: "want",
								},
								Value: &semantic.IntegerLiteral{
									Value: 0,
								},
							},
							{
								Key: &semantic.Identifier{
									Name: "got",
								},
								Value: &semantic.IntegerLiteral{
									Value: 0,
								},
							},
						},
					},
				},
			},
			want: `{"type":"TestStatement","assignment":{"type":"NativeVariableAssignment","identifier":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}`,
		},
		{
			name: "expression statement",
			node: &semantic.ExpressionStatement{
				Expression: &semantic.StringLiteral{Value: "hello"},
			},
			want: `{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "return statement",
			node: &semantic.ReturnStatement{
				Argument: &semantic.StringLiteral{Value: "hello"},
			},
			want: `{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "variable assignment",
			node: &semantic.NativeVariableAssignment{
				Identifier: &semantic.Identifier{Name: "a"},
				Init:       &semantic.StringLiteral{Value: "hello"},
			},
			want: `{"type":"NativeVariableAssignment","identifier":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "call expression",
			node: &semantic.CallExpression{
				Callee:    &semantic.IdentifierExpression{Name: "a"},
				Arguments: &semantic.ObjectExpression{Properties: []*semantic.Property{{Key: &semantic.Identifier{Name: "s"}, Value: &semantic.StringLiteral{Value: "hello"}}}},
			},
			want: `{"type":"CallExpression","callee":{"type":"IdentifierExpression","name":"a"},"arguments":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"s"},"value":{"type":"StringLiteral","value":"hello"}}]}}`,
		},
		{
			name: "member expression",
			node: &semantic.MemberExpression{
				Object:   &semantic.IdentifierExpression{Name: "a"},
				Property: "hello",
			},
			want: `{"type":"MemberExpression","object":{"type":"IdentifierExpression","name":"a"},"property":"hello"}`,
		},
		{
			name: "index expression",
			node: &semantic.IndexExpression{
				Array: &semantic.IdentifierExpression{Name: "a"},
				Index: &semantic.IntegerLiteral{Value: 3},
			},
			want: `{"type":"IndexExpression","array":{"type":"IdentifierExpression","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}`,
		},
		{
			name: "function expression",
			node: &semantic.FunctionExpression{
				Defaults: &semantic.ObjectExpression{
					Properties: []*semantic.Property{{Key: &semantic.Identifier{Name: "a"}, Value: &semantic.StringLiteral{Value: "hi"}}},
				},
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "a"}}}},
					Body:       &semantic.IdentifierExpression{Name: "a"},
				},
			},
			want: `{"type":"FunctionExpression","defaults":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hi"}}]},"block":{"type":"FunctionBlock","parameters":{"type":"FunctionParameters","list":[{"type":"FunctionParameter","key":{"type":"Identifier","name":"a"}}],"pipe":null},"body":{"type":"IdentifierExpression","name":"a"}}}`,
		},
		{
			name: "binary expression",
			node: &semantic.BinaryExpression{
				Operator: ast.AdditionOperator,
				Left:     &semantic.StringLiteral{Value: "hello"},
				Right:    &semantic.StringLiteral{Value: "world"},
			},
			want: `{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}`,
		},
		{
			name: "unary expression",
			node: &semantic.UnaryExpression{
				Operator: ast.NotOperator,
				Argument: &semantic.BooleanLiteral{Value: true},
			},
			want: `{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}`,
		},
		{
			name: "logical expression",
			node: &semantic.LogicalExpression{
				Operator: ast.OrOperator,
				Left:     &semantic.BooleanLiteral{Value: false},
				Right:    &semantic.BooleanLiteral{Value: true},
			},
			want: `{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}`,
		},
		{
			name: "array expression",
			node: &semantic.ArrayExpression{
				Elements: []semantic.Expression{&semantic.StringLiteral{Value: "hello"}},
			},
			want: `{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}`,
		},
		{
			name: "object expression",
			node: &semantic.ObjectExpression{
				Properties: []*semantic.Property{{
					Key:   &semantic.Identifier{Name: "a"},
					Value: &semantic.StringLiteral{Value: "hello"},
				}},
			},
			want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
		},
		{
			name: "conditional expression",
			node: &semantic.ConditionalExpression{
				Test:       &semantic.BooleanLiteral{Value: true},
				Alternate:  &semantic.StringLiteral{Value: "false"},
				Consequent: &semantic.StringLiteral{Value: "true"},
			},
			want: `{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"alternate":{"type":"StringLiteral","value":"false"},"consequent":{"type":"StringLiteral","value":"true"}}`,
		},
		{
			name: "property",
			node: &semantic.Property{
				Key:   &semantic.Identifier{Name: "a"},
				Value: &semantic.StringLiteral{Value: "hello"},
			},
			want: `{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "string key property",
			node: &semantic.Property{
				Key:   &semantic.StringLiteral{Value: "a"},
				Value: &semantic.StringLiteral{Value: "hello"},
			},
			want: `{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}`,
		},
		{
			name: "identifier",
			node: &semantic.Identifier{
				Name: "a",
			},
			want: `{"type":"Identifier","name":"a"}`,
		},
		{
			name: "string literal",
			node: &semantic.StringLiteral{
				Value: "hello",
			},
			want: `{"type":"StringLiteral","value":"hello"}`,
		},
		{
			name: "boolean literal",
			node: &semantic.BooleanLiteral{
				Value: true,
			},
			want: `{"type":"BooleanLiteral","value":true}`,
		},
		{
			name: "float literal",
			node: &semantic.FloatLiteral{
				Value: 42.1,
			},
			want: `{"type":"FloatLiteral","value":42.1}`,
		},
		{
			name: "integer literal",
			node: &semantic.IntegerLiteral{
				Value: math.MaxInt64,
			},
			want: `{"type":"IntegerLiteral","value":"9223372036854775807"}`,
		},
		{
			name: "unsigned integer literal",
			node: &semantic.UnsignedIntegerLiteral{
				Value: math.MaxUint64,
			},
			want: `{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}`,
		},
		{
			name: "regexp literal",
			node: &semantic.RegexpLiteral{
				Value: regexp.MustCompile(`.*`),
			},
			want: `{"type":"RegexpLiteral","value":".*"}`,
		},
		{
			name: "duration literal",
			node: &semantic.DurationLiteral{
				Values: []ast.Duration{
					{Magnitude: 1, Unit: ast.HourUnit},
					{Magnitude: 1, Unit: ast.MinuteUnit},
				},
			},
			want: `{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"m"}]}`,
		},
		{
			name: "datetime literal",
			node: &semantic.DateTimeLiteral{
				Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
			},
			want: `{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008Z"}`,
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
			node, err := semantic.UnmarshalNode(data)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.node, node, semantictest.CmpOptions...) {
				t.Errorf("unexpected node after unmarshalling: -want/+got:\n%s", cmp.Diff(tc.node, node, semantictest.CmpOptions...))
			}
		})
	}
}
