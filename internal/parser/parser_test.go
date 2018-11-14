package parser_test

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/scanner"
)

var CompareOptions = []cmp.Option{
	cmp.Transformer("", func(re *regexp.Regexp) string {
		return re.String()
	}),
}

func TestParser(t *testing.T) {
	for _, tt := range []struct {
		name string
		raw  string
		want *ast.Program
		skip bool
	}{
		{
			name: "optional query metadata",
			raw: `option task = {
				name: "foo",
				every: 1h,
				delay: 10m,
				cron: "0 2 * * *",
				retry: 5,
			  }`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Declaration: &ast.VariableDeclarator{
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
									{
										Key: &ast.Identifier{Name: "delay"},
										Value: &ast.DurationLiteral{
											Values: []ast.Duration{
												{
													Magnitude: 10,
													Unit:      "m",
												},
											},
										},
									},
									{
										Key:   &ast.Identifier{Name: "cron"},
										Value: &ast.StringLiteral{Value: "0 2 * * *"},
									},
									{
										Key:   &ast.Identifier{Name: "retry"},
										Value: &ast.IntegerLiteral{Value: 5},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "optional query metadata preceding query text",
			raw: `option task = {
					name: "foo",  // Name of task
					every: 1h,    // Execution frequency of task
				}

				// Task will execute the following query
				from() |> count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Declaration: &ast.VariableDeclarator{
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
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "from"},
								Arguments: nil,
							},
							Call: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "count"},
								Arguments: nil,
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from",
			raw:  `from()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "comment",
			raw: `// Comment
			from()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "identifier with number",
			raw:  `tan2()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "tan2",
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "regex literal",
			raw:  `/.*/`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.RegexpLiteral{
							Value: regexp.MustCompile(".*"),
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "regex literal with escape sequence",
			raw:  `/a\/b\\c\d/`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.RegexpLiteral{
							Value: regexp.MustCompile(`a/b\\c\d`),
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "regex match operators",
			raw:  `"a" =~ /.*/ and "b" !~ /c$/`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.AndOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.RegexpMatchOperator,
								Left:     &ast.StringLiteral{Value: "a"},
								Right:    &ast.RegexpLiteral{Value: regexp.MustCompile(".*")},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.NotRegexpMatchOperator,
								Left:     &ast.StringLiteral{Value: "b"},
								Right:    &ast.RegexpLiteral{Value: regexp.MustCompile("c$")},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "declare variable as an int",
			raw:  `howdy = 1`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "declare variable as a float",
			raw:  `howdy = 1.1`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.FloatLiteral{Value: 1.1},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "declare variable as an array",
			raw:  `howdy = [1, 2, 3, 4]`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{Name: "howdy"},
							Init: &ast.ArrayExpression{
								Elements: []ast.Expression{
									&ast.IntegerLiteral{Value: 1},
									&ast.IntegerLiteral{Value: 2},
									&ast.IntegerLiteral{Value: 3},
									&ast.IntegerLiteral{Value: 4},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "use variable to declare something",
			raw: `howdy = 1
			from()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "variable is from statement",
			raw: `howdy = from()
			howdy.count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "howdy",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "howdy",
								},
								Property: &ast.Identifier{
									Name: "count",
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "pipe expression",
			raw:  `from() |> count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "from"},
								Arguments: nil,
							},
							Call: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "count"},
								Arguments: nil,
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "literal pipe expression",
			raw:  `5 |> pow2()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.IntegerLiteral{Value: 5},
							Call: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "pow2"},
								Arguments: nil,
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "member expression pipe expression",
			raw:  `foo.bar |> baz()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "foo"},
								Property: &ast.Identifier{Name: "bar"},
							},
							Call: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "baz"},
								Arguments: nil,
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "multiple pipe expressions",
			raw:  `from() |> range() |> filter() |> count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.PipeExpression{
								Argument: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "range"},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "filter"},
								},
							},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "count"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "two variables for two froms",
			raw: `howdy = from()
			doody = from()
			howdy|>count()
			doody|>sum()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "howdy",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "doody",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.Identifier{Name: "howdy"},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "count",
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.Identifier{Name: "doody"},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "sum",
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with database",
			raw:  `from(bucket:"telegraf/autogen")`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{
												Name: "bucket",
											},
											Value: &ast.StringLiteral{
												Value: "telegraf/autogen",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "map member expressions",
			raw: `m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "m",
							},
							Init: &ast.ObjectExpression{
								Properties: []*ast.Property{
									{
										Key:   &ast.Identifier{Name: "key1"},
										Value: &ast.IntegerLiteral{Value: 1},
									},
									{
										Key:   &ast.Identifier{Name: "key2"},
										Value: &ast.StringLiteral{Value: "value2"},
									},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "m"},
							Property: &ast.Identifier{Name: "key1"},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "m"},
							Property: &ast.StringLiteral{Value: "key2"},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "var as binary expression of other vars",
			raw: `a = 1
            b = 2
            c = a + b
            d = a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.IntegerLiteral{Value: 2},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.Identifier{Name: "b"},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "d",
							},
							Init: &ast.Identifier{Name: "a"},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "var as unary expression of other vars",
			raw: `a = 5
            c = -a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 5},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.UnaryExpression{
								Operator: ast.SubtractionOperator,
								Argument: &ast.Identifier{Name: "a"},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "var as both binary and unary expressions",
			raw: `a = 5
            c = 10 * -a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 5},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.MultiplicationOperator,
								Left:     &ast.IntegerLiteral{Value: 10},
								Right: &ast.UnaryExpression{
									Operator: ast.SubtractionOperator,
									Argument: &ast.Identifier{Name: "a"},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "unary expressions within logical expression",
			raw: `a = 5.0
            10.0 * -a == -0.5 or a == 6.0`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.FloatLiteral{Value: 5},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.BinaryExpression{
									Operator: ast.MultiplicationOperator,
									Left:     &ast.FloatLiteral{Value: 10},
									Right: &ast.UnaryExpression{
										Operator: ast.SubtractionOperator,
										Argument: &ast.Identifier{Name: "a"},
									},
								},
								Right: &ast.UnaryExpression{
									Operator: ast.SubtractionOperator,
									Argument: &ast.FloatLiteral{Value: 0.5},
								},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.FloatLiteral{Value: 6},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "unary expressions with too many comments",
			raw: `// define a
a = 5.0
// eval this
10.0 * -a == -0.5
	// or this
	or a == 6.0`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.FloatLiteral{Value: 5},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.BinaryExpression{
									Operator: ast.MultiplicationOperator,
									Left:     &ast.FloatLiteral{Value: 10},
									Right: &ast.UnaryExpression{
										Operator: ast.SubtractionOperator,
										Argument: &ast.Identifier{Name: "a"},
									},
								},
								Right: &ast.UnaryExpression{
									Operator: ast.SubtractionOperator,
									Argument: &ast.FloatLiteral{Value: 0.5},
								},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.FloatLiteral{Value: 6},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "expressions with function calls",
			raw:  `a = foo() == 10`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "foo"},
								},
								Right: &ast.IntegerLiteral{Value: 10},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "mix unary logical and binary expressions",
			raw: `
            not (f() == 6.0 * x) or fail()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.UnaryExpression{
								Operator: ast.NotOperator,
								Argument: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "f"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.MultiplicationOperator,
										Left:     &ast.FloatLiteral{Value: 6},
										Right:    &ast.Identifier{Name: "x"},
									},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "fail"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "mix unary logical and binary expressions with extra parens",
			raw: `
            (not (f() == 6.0 * x) or fail())`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.UnaryExpression{
								Operator: ast.NotOperator,
								Argument: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "f"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.MultiplicationOperator,
										Left:     &ast.FloatLiteral{Value: 6},
										Right:    &ast.Identifier{Name: "x"},
									},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "fail"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function called",
			raw: `plusOne = (r) => r + 1
			plusOne(r:5)
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "plusOne",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "r"},
									Right:    &ast.IntegerLiteral{Value: 1},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plusOne"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{
												Name: "r",
											},
											Value: &ast.IntegerLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function return map",
			raw:  `toMap = (r) =>({r:r})`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "toMap",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.ObjectExpression{
									Properties: []*ast.Property{{
										Key:   &ast.Identifier{Name: "r"},
										Value: &ast.Identifier{Name: "r"},
									}},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function with default arg",
			raw:  `addN = (r, n=5) => r + n`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "addN",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{
									{Key: &ast.Identifier{Name: "r"}},
									{Key: &ast.Identifier{Name: "n"}, Value: &ast.IntegerLiteral{Value: 5}},
								},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "r"},
									Right:    &ast.Identifier{Name: "n"},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function called in binary expression",
			raw: `
            plusOne = (r) => r + 1
            plusOne(r:5) == 6 or die()
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "plusOne",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "r"},
									Right:    &ast.IntegerLiteral{Value: 1},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "plusOne"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{
														Name: "r",
													},
													Value: &ast.IntegerLiteral{
														Value: 5,
													},
												},
											},
										},
									},
								},
								Right: &ast.IntegerLiteral{Value: 6},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "die"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function as single expression",
			raw:  `f = (r) => r["_measurement"] == "cpu"`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "f",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "r"},
										Property: &ast.StringLiteral{Value: "_measurement"},
									},
									Right: &ast.StringLiteral{Value: "cpu"},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "arrow function as block",
			raw: `f = (r) => { 
                m = r["_measurement"]
                return m == "cpu"
            }`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "f",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.BlockStatement{
									Body: []ast.Statement{
										&ast.VariableDeclaration{
											Declarations: []*ast.VariableDeclarator{{
												ID: &ast.Identifier{
													Name: "m",
												},
												Init: &ast.MemberExpression{
													Object:   &ast.Identifier{Name: "r"},
													Property: &ast.StringLiteral{Value: "_measurement"},
												},
											}},
										},
										&ast.ReturnStatement{
											Argument: &ast.BinaryExpression{
												Operator: ast.EqualOperator,
												Left:     &ast.Identifier{Name: "m"},
												Right:    &ast.StringLiteral{Value: "cpu"},
											},
										},
									},
								},
							},
						}},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with filter with no parens",
			raw:  `from(bucket:"telegraf/autogen").filter(fn: (r) => r["other"]=="mem" and r["this"]=="that" or r["these"]!="those")`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Property: &ast.Identifier{Name: "filter"},
								Object: &ast.CallExpression{
									Callee: &ast.Identifier{
										Name: "from",
									},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "bucket"},
													Value: &ast.StringLiteral{Value: "telegraf/autogen"},
												},
											},
										},
									},
								},
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
												Body: &ast.LogicalExpression{
													Operator: ast.OrOperator,
													Left: &ast.LogicalExpression{
														Operator: ast.AndOperator,
														Left: &ast.BinaryExpression{
															Operator: ast.EqualOperator,
															Left: &ast.MemberExpression{
																Object:   &ast.Identifier{Name: "r"},
																Property: &ast.StringLiteral{Value: "other"},
															},
															Right: &ast.StringLiteral{Value: "mem"},
														},
														Right: &ast.BinaryExpression{
															Operator: ast.EqualOperator,
															Left: &ast.MemberExpression{
																Object:   &ast.Identifier{Name: "r"},
																Property: &ast.StringLiteral{Value: "this"},
															},
															Right: &ast.StringLiteral{Value: "that"},
														},
													},
													Right: &ast.BinaryExpression{
														Operator: ast.NotEqualOperator,
														Left: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "r"},
															Property: &ast.StringLiteral{Value: "these"},
														},
														Right: &ast.StringLiteral{Value: "those"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with range",
			raw:  `from(bucket:"telegraf/autogen")|>range(start:-1h, end:10m)`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "from"},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key:   &ast.Identifier{Name: "bucket"},
												Value: &ast.StringLiteral{Value: "telegraf/autogen"},
											},
										},
									},
								},
							},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "range"},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key: &ast.Identifier{Name: "start"},
												Value: &ast.UnaryExpression{
													Operator: ast.SubtractionOperator,
													Argument: &ast.DurationLiteral{
														Values: []ast.Duration{
															{
																Magnitude: 1,
																Unit:      "h",
															},
														},
													},
												},
											},
											{
												Key: &ast.Identifier{Name: "end"},
												Value: &ast.DurationLiteral{
													Values: []ast.Duration{
														{
															Magnitude: 10,
															Unit:      "m",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with limit",
			raw:  `from(bucket:"telegraf/autogen")|>limit(limit:100, offset:10)`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "from"},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key:   &ast.Identifier{Name: "bucket"},
												Value: &ast.StringLiteral{Value: "telegraf/autogen"},
											},
										},
									},
								},
							},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "limit"},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key:   &ast.Identifier{Name: "limit"},
												Value: &ast.IntegerLiteral{Value: 100},
											},
											{
												Key:   &ast.Identifier{Name: "offset"},
												Value: &ast.IntegerLiteral{Value: 10},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with range and count",
			raw: `from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.PipeExpression{
								Argument: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "from"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "bucket"},
													Value: &ast.StringLiteral{Value: "mydb/autogen"},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "range"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
															Values: []ast.Duration{
																{
																	Magnitude: 4,
																	Unit:      "h",
																},
															},
														},
													},
												},
												{
													Key: &ast.Identifier{Name: "stop"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
															Values: []ast.Duration{
																{
																	Magnitude: 2,
																	Unit:      "h",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "count"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with range, limit and count",
			raw: `from(bucket:"mydb/autogen")
						|> range(start:-4h, stop:-2h)
						|> limit(n:10)
						|> count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.PipeExpression{
							Argument: &ast.PipeExpression{
								Argument: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucket"},
														Value: &ast.StringLiteral{Value: "mydb/autogen"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "range"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "start"},
														Value: &ast.UnaryExpression{
															Operator: ast.SubtractionOperator,
															Argument: &ast.DurationLiteral{
																Values: []ast.Duration{
																	{
																		Magnitude: 4,
																		Unit:      "h",
																	},
																},
															},
														},
													},
													{
														Key: &ast.Identifier{Name: "stop"},
														Value: &ast.UnaryExpression{
															Operator: ast.SubtractionOperator,
															Argument: &ast.DurationLiteral{
																Values: []ast.Duration{
																	{
																		Magnitude: 2,
																		Unit:      "h",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "limit"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{{
												Key:   &ast.Identifier{Name: "n"},
												Value: &ast.IntegerLiteral{Value: 10},
											}},
										},
									},
								},
							},
							Call: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "count"},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with join",
			raw: `
a = from(bucket:"dbA/autogen") |> range(start:-1h)
b = from(bucket:"dbB/autogen") |> range(start:-1h)
join(tables:[a,b], on:["host"], fn: (a,b) => a["_field"] + b["_field"])`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.PipeExpression{
								Argument: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "from"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "bucket"},
													Value: &ast.StringLiteral{Value: "dbA/autogen"},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "range"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
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
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.PipeExpression{
								Argument: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "from"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "bucket"},
													Value: &ast.StringLiteral{Value: "dbB/autogen"},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "range"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
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
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "join"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "tables"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.Identifier{Name: "a"},
													&ast.Identifier{Name: "b"},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "on"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{&ast.StringLiteral{Value: "host"}},
											},
										},
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Property{
													{Key: &ast.Identifier{Name: "a"}},
													{Key: &ast.Identifier{Name: "b"}},
												},
												Body: &ast.BinaryExpression{
													Operator: ast.AdditionOperator,
													Left: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "a"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
													Right: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "b"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "from with join with complex expression",
			raw: `
a = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "a")
	|> range(start:-1h)

b = from(bucket:"Flux/autogen")
	|> filter(fn: (r) => r["_measurement"] == "b")
	|> range(start:-1h)

join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])
`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.PipeExpression{
								Argument: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucket"},
														Value: &ast.StringLiteral{Value: "Flux/autogen"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "filter"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.ArrowFunctionExpression{
															Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
															Body: &ast.BinaryExpression{
																Operator: ast.EqualOperator,
																Left: &ast.MemberExpression{
																	Object:   &ast.Identifier{Name: "r"},
																	Property: &ast.StringLiteral{Value: "_measurement"},
																},
																Right: &ast.StringLiteral{Value: "a"},
															},
														},
													},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "range"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
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
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.PipeExpression{
								Argument: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucket"},
														Value: &ast.StringLiteral{Value: "Flux/autogen"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "filter"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.ArrowFunctionExpression{
															Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
															Body: &ast.BinaryExpression{
																Operator: ast.EqualOperator,
																Left: &ast.MemberExpression{
																	Object:   &ast.Identifier{Name: "r"},
																	Property: &ast.StringLiteral{Value: "_measurement"},
																},
																Right: &ast.StringLiteral{Value: "b"},
															},
														},
													},
												},
											},
										},
									},
								},
								Call: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "range"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{
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
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "join"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "tables"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.Identifier{Name: "a"},
													&ast.Identifier{Name: "b"},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "on"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.StringLiteral{
														Value: "t1",
													},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Property{
													{Key: &ast.Identifier{Name: "a"}},
													{Key: &ast.Identifier{Name: "b"}},
												},
												Body: &ast.BinaryExpression{
													Operator: ast.DivisionOperator,
													Left: &ast.BinaryExpression{
														Operator: ast.SubtractionOperator,
														Left: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "a"},
															Property: &ast.StringLiteral{Value: "_field"},
														},
														Right: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "b"},
															Property: &ast.StringLiteral{Value: "_field"},
														},
													},
													Right: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "b"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			skip: true,
		},
		{
			name: "duration literal, all units",
			raw:  `dur = 1y3mo2w1d4h1m30s1ms2µs70ns`,
			want: &ast.Program{
				Body: []ast.Statement{&ast.VariableDeclaration{
					Declarations: []*ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "dur"},
						Init: &ast.DurationLiteral{
							Values: []ast.Duration{
								{Magnitude: 1, Unit: "y"},
								{Magnitude: 3, Unit: "mo"},
								{Magnitude: 2, Unit: "w"},
								{Magnitude: 1, Unit: "d"},
								{Magnitude: 4, Unit: "h"},
								{Magnitude: 1, Unit: "m"},
								{Magnitude: 30, Unit: "s"},
								{Magnitude: 1, Unit: "ms"},
								{Magnitude: 2, Unit: "us"},
								{Magnitude: 70, Unit: "ns"},
							},
						},
					}},
				}},
			},
			skip: true,
		},
		{
			name: "duration literal, months",
			raw:  `dur = 6mo`,
			want: &ast.Program{
				Body: []ast.Statement{&ast.VariableDeclaration{
					Declarations: []*ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "dur"},
						Init: &ast.DurationLiteral{
							Values: []ast.Duration{
								{Magnitude: 6, Unit: "mo"},
							},
						},
					}},
				}},
			},
			skip: true,
		},
		{
			name: "duration literal, milliseconds",
			raw:  `dur = 500ms`,
			want: &ast.Program{
				Body: []ast.Statement{&ast.VariableDeclaration{
					Declarations: []*ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "dur"},
						Init: &ast.DurationLiteral{
							Values: []ast.Duration{
								{Magnitude: 500, Unit: "ms"},
							},
						},
					}},
				}},
			},
			skip: true,
		},
		{
			name: "duration literal, months, minutes, milliseconds",
			raw:  `dur = 6mo30m500ms`,
			want: &ast.Program{
				Body: []ast.Statement{&ast.VariableDeclaration{
					Declarations: []*ast.VariableDeclarator{{
						ID: &ast.Identifier{Name: "dur"},
						Init: &ast.DurationLiteral{
							Values: []ast.Duration{
								{Magnitude: 6, Unit: "mo"},
								{Magnitude: 30, Unit: "m"},
								{Magnitude: 500, Unit: "ms"},
							},
						},
					}},
				}},
			},
			skip: true,
		},
		{
			name: "invalid token",
			raw:  `@ ident`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.Error{
						Message: "illegal token: @",
					},
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "ident"},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fatalf := t.Fatalf
			if tt.skip {
				fatalf = t.Skipf
			}

			s := scanner.New([]byte(tt.raw))
			result := parser.NewAST(s)
			if got, want := result, tt.want; !cmp.Equal(want, got, CompareOptions...) {
				fatalf("unexpected statement -want/+got\n%s", cmp.Diff(want, got, CompareOptions...))
			}
		})
	}
}
