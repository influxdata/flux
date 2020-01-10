package interpreter_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/interpreter/interptest"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var testScope = values.NewNestedScope(nil, values.NewObjectWithValues(
	map[string]values.Value{
		"true":  values.NewBool(true),
		"false": values.NewBool(false),
	}))

var optionsObject = values.NewObject(semantic.NewObjectType([]semantic.PropertyType{
	{
		Key:   []byte("name"),
		Value: semantic.BasicString,
	},
	{
		Key:   []byte("repeat"),
		Value: semantic.BasicInt,
	},
}))

func addFunc(f *function) {
	testScope.Set(f.name, f)
}

func addOption(name string, opt values.Value) {
	testScope.Set(name, opt)
}

func addValue(name string, v values.Value) {
	addOption(name, v)
}

func init() {
	optionsObject.Set("name", values.NewString("foo"))
	optionsObject.Set("repeat", values.NewInt(100))

	addOption("task", optionsObject)
	addValue("NULL", values.NewNull(semantic.BasicInt))

	addFunc(&function{
		name: "fortyTwo",
		t:    semantic.NewFunctionType(semantic.BasicFloat, nil),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewFloat(42.0), nil
		},
		hasSideEffect: false,
	})
	addFunc(&function{
		name: "six",
		t:    semantic.NewFunctionType(semantic.BasicFloat, nil),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewFloat(6.0), nil
		},
		hasSideEffect: false,
	})
	addFunc(&function{
		name: "nine",
		t:    semantic.NewFunctionType(semantic.BasicFloat, nil),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewFloat(9.0), nil
		},
		hasSideEffect: false,
	})
	addFunc(&function{
		name: "fail",
		t:    semantic.NewFunctionType(semantic.BasicBool, nil),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			return nil, errors.New("fail")
		},
		hasSideEffect: false,
	})
	addFunc(&function{
		name: "plusOne",
		t: semantic.NewFunctionType(semantic.BasicFloat, []semantic.ArgumentType{{
			Name: []byte("x"),
			Type: semantic.BasicFloat,
			Pipe: true,
		}}),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			v, ok := args.Get("x")
			if !ok {
				return nil, errors.New("missing argument x")
			}
			return values.NewFloat(v.Float() + 1), nil
		},
		hasSideEffect: false,
	})
	addFunc(&function{
		name: "sideEffect",
		t:    semantic.NewFunctionType(semantic.BasicInt, nil),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			return values.NewInt(0), nil
		},
		hasSideEffect: true,
	})
}

// TestEval tests whether a program can run to completion or not
func TestEval(t *testing.T) {
	testCases := []struct {
		name    string
		query   string
		wantErr bool
		want    []values.Value
	}{
		{
			name: "string interpolation",
			query: `
				str = "str"
				ing = "ing"
				"str + ing = ${str+ing}"`,
			want: []values.Value{
				values.NewString("str + ing = string"),
			},
		},
		{
			name: "string interpolation error",
			query: `
				a = 1
				b = 2
				"a + b = ${a + b}"`,
			wantErr: true,
		},
		{
			name:  "call builtin function",
			query: "six()",
			want: []values.Value{
				values.NewFloat(6.0),
			},
		},
		{
			name:    "call function with fail",
			query:   "fail()",
			wantErr: true,
		},
		{
			name:    "call function with duplicate args",
			query:   "plusOne(x:1.0, x:2.0)",
			wantErr: true,
		},
		{
			name:    "call function with missing args",
			query:   "plusOne()",
			wantErr: true,
		},
		{
			name: "binary expressions",
			query: `
			six = six()
			nine = nine()

			fortyTwo() == six * nine
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "logical expressions short circuit",
			query: `
            six = six()
            nine = nine()

            not (fortyTwo() == six * nine) or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "function",
			query: `
            plusSix = (r) => r + six()
            plusSix(r:1.0) == 7.0 or fail()
			`,
		},
		{
			name: "function block",
			query: `
            f = (r) => {
                r1 = 1.0 + r
                return (r + r1) / r
            }
            f(r:1.0) == 3.0 or fail()
			`,
		},
		{
			name: "function block polymorphic",
			query: `
            f = (r) => {
                r2 = r * r
                return r2 / r
            }
            f(r:2.0) == 2.0 or fail()
            f(r:2) == 2 or fail()
			`,
		},
		{
			name: "function with default param",
			query: `
            addN = (r,n=4) => r + n
            addN(r:2) == 6 or fail()
            addN(r:3,n:1) == 4 or fail()
			`,
		},
		{
			name: "scope closing",
			query: `
			x = 5
            plusX = (r) => r + x
            plusX(r:2) == 7 or fail()
			`,
		},
		{
			name: "nested scope mutations not visible outside",
			query: `
			x = 5
            xinc = () => {
                x = x + 1
                return x
            }
            xinc() == 6 or fail()
            x == 5 or fail()
			`,
		},
		// TODO(jsternberg): This test seems to not
		// infer the type constraints correctly for m.a,
		// but it doesn't fail.
		{
			name: "return map from func",
			query: `
            toMap = (a,b) => ({
                a: a,
                b: b,
            })
            m = toMap(a:1, b:false)
            m.a == 1 or fail()
            not m.b or fail()
			`,
		},
		{
			name: "pipe expression",
			query: `
			add = (a=<-,b) => a + b
			one = 1
			one |> add(b:2) == 3 or fail()
			`,
		},
		{
			name: "ignore pipe default",
			query: `
			add = (a=<-,b) => a + b
			add(a:1, b:2) == 3 or fail()
			`,
		},
		{
			name: "missing pipe",
			query: `
			add = (a=<-,b) => a + b
			add(b:2) == 3 or fail()
			`,
			wantErr: true,
		},
		{
			name: "pipe expression function",
			query: `
			add = (a=<-,b) => a + b
			six() |> add(b:2.0) == 8.0 or fail()
			`,
		},
		{
			name: "pipe builtin function",
			query: `
			six() |> plusOne() == 7.0 or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "regex match",
			query: `
			"abba" =~ /^a.*a$/ or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "regex not match",
			query: `
			"abc" =~ /^a.*a$/ and fail()
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "not regex match",
			query: `
			"abc" !~ /^a.*a$/ or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "not regex not match",
			query: `
			"abba" !~ /^a.*a$/ and fail()
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "options metadata",
			query: `
			option task = {
				name: "foo",
				repeat: 100,
			}
			task.name == "foo" or fail()
			task.repeat == 100 or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
				values.NewBool(true),
			},
		},
		{
			name:  "query with side effects",
			query: `sideEffect() == 0 or fail()`,
			want: []values.Value{
				values.NewInt(0),
				values.NewBool(true),
			},
		},
		{
			name: "array index expression",
			query: `
				a = [1, 2, 3]
				x = a[1]
				x == 2 or fail()
			`,
		},
		{
			name: "array with complex index expression",
			query: `
				f = () => ({l: 0, m: 1, n: 2})
				a = [1, 2, 3]
				x = a[f().l]
				y = a[f().m]
				z = a[f().n]
				x == 1 or fail()
				y == 2 or fail()
				z == 3 or fail()
			`,
		},
		{
			name: "invalid array index expression 1",
			query: `
				a = [1, 2, 3]
				a["b"]
			`,
			wantErr: true,
		},
		{
			name: "invalid array index expression 2",
			query: `
				a = [1, 2, 3]
				f = () => "1"
				a[f()]
			`,
			wantErr: true,
		},
		{
			name: "short circuit logical and",
			query: `
                false and fail()
            `,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "short circuit logical or",
			query: `
                true or fail()
            `,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "no short circuit logical and",
			query: `
                true and fail()
            `,
			wantErr: true,
		},
		{
			name: "no short circuit logical or",
			query: `
                false or fail()
            `,
			wantErr: true,
		},
		{
			name: "conditional true",
			query: `
				if 1 != 0 then 10 else 100
			`,
			want: []values.Value{
				values.NewInt(10),
			},
		},
		{
			name: "conditional false",
			query: `
				if 1 == 0 then 10 else 100
			`,
			want: []values.Value{
				values.NewInt(100),
			},
		},
		{
			name: "conditional in function",
			query: `
				f = (t, c, a) => if t then c else a
				{
					v1: f(t: false, c: 30, a: 300),
					v2: f(t: true, c: "cats", a: "dogs"),
				}
			`,
			want: []values.Value{
				values.NewObjectWithValues(map[string]values.Value{
					"v1": values.NewInt(300),
					"v2": values.NewString("cats"),
				}),
			},
		},
		{
			name: "exists",
			query: `
				exists 1
				exists NULL`,
			want: []values.Value{
				values.NewBool(true),
				values.NewBool(false),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pkg := parser.ParseSource(tc.query)
			if ast.Check(pkg) > 0 {
				t.Fatal(ast.GetError(pkg))
			}
			graph, err := semantic.New(pkg)
			if err != nil {
				t.Fatal(err)
			}

			// Create new interpreter for each test case
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))

			sideEffects, err := itrp.Eval(dependenciestest.Default().Inject(context.Background()), graph, testScope.Copy(), nil)
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}

			vs := getSideEffectsValues(sideEffects)
			if tc.want != nil && !cmp.Equal(tc.want, vs, semantictest.CmpOptions...) {
				t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(tc.want, vs, semantictest.CmpOptions...))
			}
		})
	}
}

// TestEval_Parallel ensures that function values returned from the interpreter can be used in parallel in multiple Eval calls.
func TestEval_Parallel(t *testing.T) {
	var scope = values.NewScope()

	{
		var ident = "ident = (x) => x"
		pkg := parser.ParseSource(ident)
		if ast.Check(pkg) > 0 {
			t.Fatal(ast.GetError(pkg))
		}
		graph, err := semantic.New(pkg)
		if err != nil {
			t.Fatal(err)
		}
		ctx := dependenciestest.Default().Inject(context.Background())
		itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
		if _, err := itrp.Eval(ctx, graph, scope, nil); err != nil {
			t.Fatal(err)
		}
	}

	script := "ident(x:1)"
	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		t.Fatal(ast.GetError(pkg))
	}
	graph, err := semantic.New(pkg)
	if err != nil {
		t.Fatal(err)
	}

	// Spin up multiple interpreters all using the same parent scope
	n := 100
	errC := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
			ctx := dependenciestest.Default().Inject(context.Background())
			_, err := itrp.Eval(ctx, graph, scope.Nest(nil), nil)
			errC <- err
		}()
	}
	for i := 0; i < n; i++ {
		err := <-errC
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNestedExternBlocks(t *testing.T) {
	testcases := []struct {
		packageNode *semantic.Package
		externScope values.Scope
		wantError   error
	}{
		{
			packageNode: &semantic.Package{
				Files: []*semantic.File{
					{
						Body: []semantic.Statement{
							&semantic.OptionStatement{
								Assignment: &semantic.NativeVariableAssignment{
									Identifier: &semantic.Identifier{Name: "b"},
									Init:       &semantic.IdentifierExpression{Name: "a"},
								},
							},
							&semantic.OptionStatement{
								Assignment: &semantic.NativeVariableAssignment{
									Identifier: &semantic.Identifier{Name: "b"},
									Init:       &semantic.StringLiteral{Value: "-----"},
								},
							},
							&semantic.NativeVariableAssignment{
								Identifier: &semantic.Identifier{Name: "a"},
								Init:       &semantic.FloatLiteral{Value: 0.055},
							},
						},
					},
				},
			},
			externScope: values.NewNestedScope(nil, values.NewObjectWithValues(
				map[string]values.Value{
					// initial 'a' value with type int
					"a": values.NewInt(0),
				})).Nest(values.NewObjectWithValues(
				map[string]values.Value{
					// 'a' shadowed, given new type string
					"a": values.NewString("0"),
				})).Nest(nil),
		},
		{
			packageNode: &semantic.Package{
				Files: []*semantic.File{
					{
						Body: []semantic.Statement{
							&semantic.OptionStatement{
								// 'b' should be of type int
								Assignment: &semantic.NativeVariableAssignment{
									Identifier: &semantic.Identifier{Name: "b"},
									Init:       &semantic.IdentifierExpression{Name: "a"},
								},
							},
							&semantic.OptionStatement{
								// Assigning 'b' to value of type string should cause type error
								Assignment: &semantic.NativeVariableAssignment{
									Identifier: &semantic.Identifier{Name: "b"},
									Init:       &semantic.StringLiteral{Value: "-----"},
								},
							},
						},
					},
				},
			},
			externScope: values.NewNestedScope(nil, values.NewObjectWithValues(
				map[string]values.Value{
					// initial 'a' value with type string
					"a": values.NewString("0"),
				})).Nest(values.NewObjectWithValues(
				map[string]values.Value{
					// 'a' shadowed, given new type int
					"a": values.NewInt(0),
				})).Nest(nil),
			wantError: errors.New("type error 0:0-0:0: string != int"),
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run("", func(t *testing.T) {
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
			_, err := itrp.Eval(dependenciestest.Default().Inject(context.Background()), tc.packageNode, tc.externScope, nil)
			if tc.wantError != nil {
				if err == nil {
					t.Errorf("expected error=(%v) but got nothing", tc.wantError)
				} else if tc.wantError.Error() != err.Error() {
					t.Errorf("expected error=(%v) but got error=(%v)", tc.wantError, err)
				}
			} else if tc.wantError == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInterpreter_TypeErrors(t *testing.T) {
	testCases := []struct {
		name    string
		program string
		err     string
	}{
		{
			name: "no pipe arg",
			program: `
				f = () => 0
				g = () => 1 |> f()
				`,
			err: `function does not take a pipe argument`,
		},
		{
			name: "called without pipe args",
			program: `
				f = (x=<-) => x
				g = () => f()
			`,
			err: `function requires a pipe argument`,
		},
		{
			name: "unify with different pipe args 1",
			program: `
				f = (x) => 0 |> x()
				f(x: (v=<-) => v)
				f(x: (w=<-) => w)
			`,
		},
		{
			// This program should type check.
			// arg is any function that takes a pipe argument.
			// arg's pipe parameter can be named anything.
			name: "unify with different pipe args 2",
			program: `
				f = (arg=(x=<-) => x, w) => w |> arg()
				f(arg: (v=<-) => v, w: 0)
			`,
		},
		{
			// This program should not type check.
			// A function that requires a parameter named "arg" cannot unify
			// with a function whose "arg" parameter is also a pipe parameter.
			name: "unify pipe and non-pipe args with same name",
			program: `
				f = (x, y) => x(arg: y)
				f(x: (arg=<-) => arg, y: 0)
			`,
			err: "function does not take a pipe argument",
		},
		{
			// This program should not type check.
			// arg is a function that must take a pipe argument. Even
			// though arg defaults to a function that takes an input
			// param x, if x is not a pipe param then it cannot type check.
			name: "pipe and non-pipe parameters with the same name",
			program: `
				f = (arg=(x=<-) => x) => 0 |> arg()
				g = () => f(arg: (x) => 5 + x)
			`,
			err: `function does not take a parameter "x"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkg := parser.ParseSource(tc.program)
			if ast.Check(pkg) > 0 {
				t.Fatal(ast.GetError(pkg))
			}
			graph, err := semantic.New(pkg)
			if err != nil {
				t.Fatal(err)
			}
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
			if _, err := itrp.Eval(dependenciestest.Default().Inject(context.Background()), graph, values.NewScope(), nil); err == nil {
				if tc.err != "" {
					t.Error("expected type error, but program executed successfully")
				}
			} else {
				if tc.err == "" {
					t.Errorf("expected zero errors, but got %v", err)
				} else if !strings.Contains(err.Error(), "type error") {
					t.Errorf("expected type error, but got the following: %v", err)
				} else if !strings.Contains(err.Error(), tc.err) {
					t.Errorf("wrong error message\n expected error message to contain: %q\n actual error message: %q\n", tc.err, err.Error())
				}
			}
		})
	}
}

func TestInterpreter_MultiPhaseInterpretation(t *testing.T) {
	testCases := []struct {
		name     string
		builtins []string
		program  string
		wantErr  bool
		want     []values.Value
	}{
		{
			// Evaluate two builtin functions in a single phase
			name: "2-phase interpretation",
			builtins: []string{
				`
					_highestOrLowest = (table=<-, reducer) => table |> reducer()
					highestCurrent = (table=<-) => table |> _highestOrLowest(reducer: (table=<-) => table)
				`,
			},
			program: `5 |> highestCurrent()`,
		},
		{
			// Evaluate two builtin functions each in a separate phase
			name: "3-phase interpretation",
			builtins: []string{
				`_highestOrLowest = (table=<-, reducer) => table |> reducer()`,
				`highestCurrent = (table=<-) => table |> _highestOrLowest(reducer: (table=<-) => table)`,
			},
			program: `5 |> highestCurrent()`,
		},
		{
			// Type-check function expression even though it is not called
			// Program is correctly typed so it should not throw any type errors
			name:     "builtin not called - no type error",
			builtins: []string{`_highestOrLowest = (table=<-, reducer) => table |> reducer()`},
			program:  `f = () => 5 |> _highestOrLowest(reducer: (table=<-) => table)`,
		},
		{
			// Type-check function expression even though it is not called
			// Program should not type check due to missing pipe parameter
			name:     "builtin not called - type error",
			builtins: []string{`_highestOrLowest = (table=<-) => table`},
			program:  `f = () => _highestOrLowest()`,
			wantErr:  true,
		},
		{
			name:     "query function with side effects",
			builtins: []string{`foo = () => {sideEffect() return 1}`},
			program:  `foo()`,
			want: []values.Value{
				values.NewInt(0),
				values.NewInt(1),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
			scope := testScope.Copy()

			for _, builtin := range tc.builtins {
				if _, err := interptest.Eval(ctx, itrp, scope, nil, builtin); err != nil {
					t.Fatal("evaluation of builtin failed: ", err)
				}
			}

			sideEffects, err := interptest.Eval(ctx, itrp, scope, nil, tc.program)
			if err != nil && !tc.wantErr {
				t.Fatal("program evaluation failed: ", err)
			} else if err == nil && tc.wantErr {
				t.Fatal("expected to error during program evaluation")
			}

			if tc.want != nil {
				if want, got := tc.want, getSideEffectsValues(sideEffects); !cmp.Equal(want, got, semantictest.CmpOptions...) {
					t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
				}
			}
		})
	}
}

// TestInterpreter_MultipleEval tests that multiple calls to `Eval` to the same interpreter behave as expected.
func TestInterpreter_MultipleEval(t *testing.T) {
	type scriptWithSideEffects struct {
		script      string
		sideEffects []interpreter.SideEffect
	}

	testCases := []struct {
		name  string
		lines []scriptWithSideEffects
	}{
		{
			name: "1 expression statement",
			lines: []scriptWithSideEffects{
				{
					script: `1+1`,
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(2),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.BinaryExpression{
									Left:     &semantic.IntegerLiteral{Value: 1},
									Operator: ast.AdditionOperator,
									Right:    &semantic.IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "more expression statements",
			lines: []scriptWithSideEffects{
				{
					script: `1+1`,
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(2),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.BinaryExpression{
									Left:     &semantic.IntegerLiteral{Value: 1},
									Operator: ast.AdditionOperator,
									Right:    &semantic.IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
				{
					script:      `foo = () => {sideEffect() return 1}`,
					sideEffects: []interpreter.SideEffect{}, // no side effect expected.
				},
				{
					script: `foo()`, // 2 side effects: the function call and the statement expression.
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(0),
							Node: &semantic.CallExpression{
								Callee:    &semantic.IdentifierExpression{Name: "sideEffect"},
								Arguments: &semantic.ObjectExpression{},
							},
						},
						{
							Value: values.NewInt(1),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.CallExpression{
									Callee:    &semantic.IdentifierExpression{Name: "foo"},
									Arguments: &semantic.ObjectExpression{},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
			scope := testScope.Copy()

			for _, line := range tc.lines {
				if ses, err := interptest.Eval(ctx, itrp, scope, nil, line.script); err != nil {
					t.Fatal("evaluation of builtin failed: ", err)
				} else {
					if !cmp.Equal(line.sideEffects, ses, semantictest.CmpOptions...) {
						t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(line.sideEffects, ses, semantictest.CmpOptions...))
					}
				}
			}
		})
	}
}

func TestResolver(t *testing.T) {
	var got semantic.Expression
	f := &function{
		name: "resolver",
		t: semantic.NewFunctionType(semantic.BasicInt, []semantic.ArgumentType{{
			Name: []byte("f"),
			Type: semantic.NewFunctionType(semantic.BasicInt, []semantic.ArgumentType{{
				Name: []byte("r"),
				Type: semantic.BasicInt,
			}}),
		}}),
		call: func(ctx context.Context, args values.Object) (values.Value, error) {
			f, ok := args.Get("f")
			if !ok {
				return nil, errors.New("missing argument f")
			}
			resolver, ok := f.Function().(interpreter.Resolver)
			if !ok {
				return nil, errors.New("function cannot be resolved")
			}
			g, err := resolver.Resolve()
			if err != nil {
				return nil, err
			}
			got = g.(semantic.Expression)
			return nil, nil
		},
		hasSideEffect: false,
	}
	s := make(map[string]values.Value)
	s[f.name] = f

	pkg := parser.ParseSource(`
	x = 42
	resolver(f: (r) => r + x)
`)
	if ast.Check(pkg) > 0 {
		t.Fatal(ast.GetError(pkg))
	}

	graph, err := semantic.New(pkg)
	if err != nil {
		t.Fatal(err)
	}

	ctx := dependenciestest.Default().Inject(context.Background())
	itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))

	ns := values.NewNestedScope(nil, values.NewObjectWithValues(s))

	if _, err := itrp.Eval(ctx, graph, ns, nil); err != nil {
		t.Fatal(err)
	}

	want := &semantic.FunctionExpression{
		Block: &semantic.FunctionBlock{
			Parameters: &semantic.FunctionParameters{
				List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
			},
			Body: &semantic.BinaryExpression{
				Operator: ast.AdditionOperator,
				Left:     &semantic.IdentifierExpression{Name: "r"},
				Right:    &semantic.IntegerLiteral{Value: 42},
			},
		},
	}
	if !cmp.Equal(want, got, semantictest.CmpOptions...) {
		t.Errorf("unexpected resoved function: -want/+got\n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
	}
}

func getSideEffectsValues(ses []interpreter.SideEffect) []values.Value {
	vs := make([]values.Value, len(ses))
	for i, se := range ses {
		vs[i] = se.Value
	}
	return vs
}

type function struct {
	name          string
	t             semantic.MonoType
	call          func(ctx context.Context, args values.Object) (values.Value, error)
	hasSideEffect bool
}

func (f *function) Type() semantic.MonoType {
	return f.t
}
func (f *function) IsNull() bool {
	return false
}
func (f *function) Str() string {
	panic(values.UnexpectedKind(semantic.Function, semantic.String))
}
func (f *function) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
}
func (f *function) Int() int64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Int))
}
func (f *function) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.UInt))
}
func (f *function) Float() float64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Float))
}
func (f *function) Bool() bool {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bool))
}
func (f *function) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Function, semantic.Time))
}
func (f *function) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Function, semantic.Duration))
}
func (f *function) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Function, semantic.Regexp))
}
func (f *function) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Function, semantic.Array))
}
func (f *function) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Function, semantic.Object))
}
func (f *function) Function() values.Function {
	return f
}
func (f *function) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(*function)
	return ok && (f == v)
}
func (f *function) HasSideEffect() bool {
	return f.hasSideEffect
}

func (f *function) Call(ctx context.Context, args values.Object) (values.Value, error) {
	return f.call(ctx, args)
}
