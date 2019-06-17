package compiler_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var CmpOptions []cmp.Option

func init() {
	CmpOptions = append(semantictest.CmpOptions, cmp.Comparer(ValueEqual))
}

func ValueEqual(x, y values.Value) bool {
	if x == values.Null && y == values.Null {
		return true
	}

	switch k := x.Type().Nature(); k {
	case semantic.Object:
		if x.Type() != y.Type() {
			return false
		}
		return cmp.Equal(x.Object(), y.Object(), CmpOptions...)
	default:
		return x.Equal(y)
	}
}

func TestCompilationCache(t *testing.T) {
	add := &semantic.FunctionExpression{
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
	}
	testCases := []struct {
		name   string
		inType semantic.Type
		input  values.Object
		want   values.Value
	}{
		{
			name: "floats",
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Float,
				"b": semantic.Float,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewFloat(5),
				"b": values.NewFloat(4),
			}),
			want: values.NewFloat(9),
		},
		{
			name: "ints",
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Int,
				"b": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewInt(5),
				"b": values.NewInt(4),
			}),
			want: values.NewInt(9),
		},
		{
			name: "uints",
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.UInt,
				"b": semantic.UInt,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewUInt(5),
				"b": values.NewUInt(4),
			}),
			want: values.NewUInt(9),
		},
	}

	//Reuse the same cache for all test cases
	cache := compiler.NewCompilationCache(add, nil)
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f0, err := cache.Compile(tc.inType)
			if err != nil {
				t.Fatal(err)
			}
			f1, err := cache.Compile(tc.inType)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(f0, f1) {
				t.Errorf("unexpected new compilation result")
			}

			got0, err := f0.Eval(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			got1, err := f1.Eval(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got0, tc.want, CmpOptions...) {
				t.Errorf("unexpected eval result -want/+got\n%s", cmp.Diff(tc.want, got0, CmpOptions...))
			}
			if !cmp.Equal(got0, got1, CmpOptions...) {
				t.Errorf("unexpected differing results -got0/+got1\n%s", cmp.Diff(got0, got1, CmpOptions...))
			}
		})
	}
}

func TestCompileAndEval(t *testing.T) {
	testCases := []struct {
		name    string
		fn      *semantic.FunctionExpression
		inType  semantic.Type
		input   values.Object
		want    values.Value
		wantErr bool
	}{
		{
			name: "simple ident return",
			// f = (r) => r
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.IdentifierExpression{Name: "r"},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(4),
			wantErr: false,
		},
		{
			name: "call function",
			// f = (r) => ((a,b) => a + b)(a:1, b:r)
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.CallExpression{
						Callee: &semantic.FunctionExpression{
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
						Arguments: &semantic.ObjectExpression{
							Properties: []*semantic.Property{
								{Key: &semantic.Identifier{Name: "a"}, Value: &semantic.IntegerLiteral{Value: 1}},
								{Key: &semantic.Identifier{Name: "b"}, Value: &semantic.IdentifierExpression{Name: "r"}},
							},
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call function with defaults",
			// f = (r) => ((a=0,b) => a + b)(b:r)
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.CallExpression{
						Callee: &semantic.FunctionExpression{
							Defaults: &semantic.ObjectExpression{
								Properties: []*semantic.Property{{
									Key:   &semantic.Identifier{Name: "a"},
									Value: &semantic.IntegerLiteral{Value: 0},
								}},
							},
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
						Arguments: &semantic.ObjectExpression{
							Properties: []*semantic.Property{
								{Key: &semantic.Identifier{Name: "b"}, Value: &semantic.IdentifierExpression{Name: "r"}},
							},
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(4),
			wantErr: false,
		},
		{
			name: "call function via identifier",
			// f = (r) => {f = (a,b) => a + b return f(a:1, b:r)}
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.Block{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
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
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call function via identifier with different types",
			// f = (r) => {i = (x) => x return i(x:i)(x:r+1)}
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.Block{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "i"},
								Init: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "x"}}},
										},
										Body: &semantic.IdentifierExpression{Name: "x"},
									},
								},
							},
							&semantic.ReturnStatement{
								Argument: &semantic.CallExpression{
									Callee: &semantic.CallExpression{
										Callee: &semantic.IdentifierExpression{Name: "i"},
										Arguments: &semantic.ObjectExpression{
											Properties: []*semantic.Property{
												{Key: &semantic.Identifier{Name: "x"}, Value: &semantic.IdentifierExpression{Name: "i"}},
											},
										},
									},
									Arguments: &semantic.ObjectExpression{
										Properties: []*semantic.Property{
											{
												Key: &semantic.Identifier{Name: "x"},
												Value: &semantic.BinaryExpression{
													Operator: ast.AdditionOperator,
													Left:     &semantic.IdentifierExpression{Name: "r"},
													Right:    &semantic.IntegerLiteral{Value: 1},
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
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call filter function with index expression",
			// f = (r) => r[2] == 3
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.EqualOperator,
						Left: &semantic.IndexExpression{
							Array: &semantic.IdentifierExpression{Name: "r"},
							Index: &semantic.IntegerLiteral{Value: 2},
						},
						Right: &semantic.IntegerLiteral{Value: 3},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewArrayType(semantic.Int),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewArrayWithBacking(semantic.Int, []values.Value{
					values.NewInt(5),
					values.NewInt(6),
					values.NewInt(3),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "call filter function with complex index expression",
			// f = (r) => r[((x) => 2)(x: "anything")] == 3
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.EqualOperator,
						Left: &semantic.IndexExpression{
							Array: &semantic.IdentifierExpression{Name: "r"},
							Index: &semantic.CallExpression{
								Callee: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{
												{Key: &semantic.Identifier{Name: "x"}},
											},
										},
										Body: &semantic.IntegerLiteral{Value: 2},
									},
								},
								Arguments: &semantic.ObjectExpression{
									Properties: []*semantic.Property{
										{
											Key:   &semantic.Identifier{Name: "x"},
											Value: &semantic.StringLiteral{Value: "anything"},
										},
									},
								},
							},
						},
						Right: &semantic.IntegerLiteral{Value: 3},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewArrayType(semantic.Int),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewArrayWithBacking(semantic.Int, []values.Value{
					values.NewInt(5),
					values.NewInt(6),
					values.NewInt(3),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "conditional",
			// f = (t, c, a) => if t then c else a
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "t"}},
							{Key: &semantic.Identifier{Name: "c"}},
							{Key: &semantic.Identifier{Name: "a"}},
						},
					},
					Body: &semantic.ConditionalExpression{
						Test: &semantic.IdentifierExpression{
							Name: "t",
						},
						Consequent: &semantic.IdentifierExpression{
							Name: "c",
						},
						Alternate: &semantic.IdentifierExpression{
							Name: "a",
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"t": semantic.Bool,
				"c": semantic.String,
				"a": semantic.String,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewBool(true),
				"c": values.NewString("cats"),
				"a": values.NewString("dogs"),
			}),
			want: values.NewString("cats"),
		},
		{
			name: "unary logical operator - not",
			// f = (a, b) => not a or b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "a"}},
							{Key: &semantic.Identifier{Name: "b"}},
						},
					},
					Body: &semantic.LogicalExpression{
						Operator: ast.OrOperator,
						Left: &semantic.UnaryExpression{
							Operator: ast.NotOperator,
							Argument: &semantic.IdentifierExpression{
								Name: "a",
							},
						},
						Right: &semantic.IdentifierExpression{
							Name: "b",
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Bool,
				"b": semantic.Bool,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewBool(true),
			}),
			want: values.NewBool(true),
		},
		{
			name: "unary logical operator - exists with null",
			// f = (a) => exists a
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "a"}},
						},
					},
					Body: &semantic.UnaryExpression{
						Operator: ast.ExistsOperator,
						Argument: &semantic.IdentifierExpression{
							Name: "a",
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.String,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewNull(semantic.String),
			}),
			want: values.NewBool(false),
		},
		{
			name: "unary logical operator - exists without null",
			// f = (a, b) => not a and exists b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "a"}},
							{Key: &semantic.Identifier{Name: "b"}},
						},
					},
					Body: &semantic.LogicalExpression{
						Operator: ast.AndOperator,
						Left: &semantic.UnaryExpression{
							Operator: ast.NotOperator,
							Argument: &semantic.IdentifierExpression{
								Name: "a",
							},
						},
						Right: &semantic.UnaryExpression{
							Operator: ast.ExistsOperator,
							Argument: &semantic.IdentifierExpression{
								Name: "b",
							},
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Bool,
				"b": semantic.String,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewString("I exist"),
			}),
			want: values.NewBool(false),
		},
		{
			name: "unary operator",
			// f = (a) => if a < 0 then -a else +a
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "a"}},
						},
					},
					Body: &semantic.ConditionalExpression{
						Test: &semantic.BinaryExpression{
							Operator: ast.LessThanOperator,
							Left: &semantic.IdentifierExpression{
								Name: "a",
							},
							Right: &semantic.IntegerLiteral{
								Value: 0,
							},
						},
						Consequent: &semantic.UnaryExpression{
							Operator: ast.SubtractionOperator,
							Argument: &semantic.IdentifierExpression{
								Name: "a",
							},
						},
						Alternate: &semantic.UnaryExpression{
							Operator: ast.AdditionOperator,
							Argument: &semantic.IdentifierExpression{
								Name: "a",
							},
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewInt(5),
			}),
			want: values.NewInt(5),
		},
		{
			name: "filter with member expression",
			// f = (r) => r.m == "cpu"
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.EqualOperator,
						Left: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "m",
						},
						Right: &semantic.StringLiteral{Value: "cpu"},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"m": semantic.String,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"m": values.NewString("cpu"),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "regex literal filter",
			// f = (r) => r =~ /^(c|g)pu$/
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.RegexpMatchOperator,
						Left:     &semantic.IdentifierExpression{Name: "r"},
						Right: &semantic.RegexpLiteral{
							Value: regexp.MustCompile(`^(c|g)pu$`),
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.String,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewString("cpu"),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "block statement with conditional",
			// f = (r) => {
			//   v = if r < 0 then -r else r
			//   return v * v
			// }
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.Block{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "v"},
								Init: &semantic.ConditionalExpression{
									Test: &semantic.BinaryExpression{
										Operator: ast.LessThanOperator,
										Left:     &semantic.IdentifierExpression{Name: "r"},
										Right:    &semantic.IntegerLiteral{Value: 0},
									},
									Consequent: &semantic.UnaryExpression{
										Operator: ast.SubtractionOperator,
										Argument: &semantic.IdentifierExpression{Name: "r"},
									},
									Alternate: &semantic.IdentifierExpression{Name: "r"},
								},
							},
							&semantic.ReturnStatement{
								Argument: &semantic.BinaryExpression{
									Operator: ast.MultiplicationOperator,
									Left:     &semantic.IdentifierExpression{Name: "v"},
									Right:    &semantic.IdentifierExpression{Name: "v"},
								},
							},
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.Int,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(-3),
			}),
			want:    values.NewInt(9),
			wantErr: false,
		},
		{
			name: "array literal",
			// f = () => [1.0, 2.0, 3.0]
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{},
					Body: &semantic.ArrayExpression{
						Elements: []semantic.Expression{
							&semantic.FloatLiteral{Value: 1},
							&semantic.FloatLiteral{Value: 2},
							&semantic.FloatLiteral{Value: 3},
						},
					},
				},
			},
			inType: semantic.NewObjectType(nil),
			input:  values.NewObjectWithValues(nil),
			want: values.NewArrayWithBacking(
				semantic.Float,
				[]values.Value{
					values.NewFloat(1),
					values.NewFloat(2),
					values.NewFloat(3),
				},
			),
			wantErr: false,
		},
		{
			name: "logical expression",
			// f = (a, b) => a or b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "a"}},
							{Key: &semantic.Identifier{Name: "b"}},
						},
					},
					Body: &semantic.LogicalExpression{
						Operator: ast.OrOperator,
						Left:     &semantic.IdentifierExpression{Name: "a"},
						Right:    &semantic.IdentifierExpression{Name: "b"},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"a": semantic.Bool,
				"b": semantic.Bool,
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewBool(false),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "call with nonexistant value",
			// f = (r) => r.a + r.b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.AdditionOperator,
						Left: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "a",
						},
						Right: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "b",
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"a": semantic.Int,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
				}),
			}),
			want: values.Null,
		},
		{
			name: "call with null value",
			// f = (r) => r.a + r.b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.AdditionOperator,
						Left: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "a",
						},
						Right: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "b",
						},
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"a": semantic.Int,
					"b": semantic.Int,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// The object is typed as an integer,
					// but it doesn't have an actual value
					// because it is null.
					"b": values.Null,
				}),
			}),
			want: values.Null,
		},
		{
			name: "call with null parameter",
			// f = (r) => {
			//   eval = (a, b) => a + b
			//   return eval(a: r.a, b: r.b)
			// }
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.Block{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "eval"},
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
									Callee: &semantic.IdentifierExpression{Name: "eval"},
									Arguments: &semantic.ObjectExpression{
										Properties: []*semantic.Property{
											{
												Key: &semantic.Identifier{Name: "a"},
												Value: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "a",
												},
											},
											{
												Key: &semantic.Identifier{Name: "b"},
												Value: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "b",
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
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"a": semantic.Int,
					// "b": semantic.Nil,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// "b": values.Null,
				}),
			}),
			want: values.Null,
		},
		{
			name: "return nonexistant value",
			// f = (r) => r.b
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.MemberExpression{
						Object:   &semantic.IdentifierExpression{Name: "r"},
						Property: "b",
					},
				},
			},
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"a": semantic.Int,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
				}),
			}),
			want: values.Null,
		},
		{
			name: "return nonexistant and used parameter",
			// f = (r) => {
			//     b = (r) => r.b
			//     return r.a + b(r: r)
			// }
			fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{
							{Key: &semantic.Identifier{Name: "r"}},
						},
					},
					Body: &semantic.Block{
						Body: []semantic.Statement{
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "b"},
								Init: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{
												{Key: &semantic.Identifier{Name: "r"}},
											},
										},
										Body: &semantic.MemberExpression{
											Object:   &semantic.IdentifierExpression{Name: "r"},
											Property: "b",
										},
									},
								},
							},
							&semantic.ReturnStatement{
								Argument: &semantic.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left: &semantic.MemberExpression{
										Object:   &semantic.IdentifierExpression{Name: "r"},
										Property: "a",
									},
									Right: &semantic.CallExpression{
										Callee: &semantic.IdentifierExpression{Name: "b"},
										Arguments: &semantic.ObjectExpression{
											Properties: []*semantic.Property{
												{
													Key:   &semantic.Identifier{Name: "r"},
													Value: &semantic.IdentifierExpression{Name: "r"},
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
			inType: semantic.NewObjectType(map[string]semantic.Type{
				"r": semantic.NewObjectType(map[string]semantic.Type{
					"a": semantic.Int,
					// "b": semantic.Nil,
				}),
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// "b": values.Null,
				}),
			}),
			want: values.Null,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f, err := compiler.Compile(tc.fn, tc.inType, nil)
			if tc.wantErr != (err != nil) {
				t.Fatalf("unexpected error: %s", err)
			}

			got, err := f.Eval(tc.input)
			if tc.wantErr != (err != nil) {
				t.Errorf("unexpected error: %s", err)
			}

			if !cmp.Equal(tc.want, got, CmpOptions...) {
				t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(tc.want, got, CmpOptions...))
			}
		})
	}
}
