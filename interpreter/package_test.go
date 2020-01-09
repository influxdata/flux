package interpreter_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/interpreter/interptest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Implementation of interpreter.Importer
type importer struct {
	packages map[string]*interpreter.Package
}

func (imp *importer) Import(path string) (semantic.MonoType, bool) {
	pkg, ok := imp.packages[path]
	if !ok {
		return semantic.MonoType{}, false
	}
	return pkg.Type(), true
}

func (imp *importer) ImportPackageObject(path string) (*interpreter.Package, bool) {
	pkg, ok := imp.packages[path]
	return pkg, ok
}

func TestAccessNestedImport(t *testing.T) {
	// package a
	// x = 0
	packageA := interpreter.NewPackageWithValues("a", values.NewObjectWithValues(map[string]values.Value{
		"x": values.NewInt(0),
	}))

	// package b
	// import "a"
	packageB := interpreter.NewPackageWithValues("b", values.NewObjectWithValues(map[string]values.Value{
		"a": packageA,
	}))

	// package c
	// import "b"
	// e = b.a.x
	node := &semantic.Package{
		Package: "c",
		Files: []*semantic.File{
			{
				Package: &semantic.PackageClause{
					Name: &semantic.Identifier{Name: "c"},
				},
				Imports: []*semantic.ImportDeclaration{
					{
						Path: &semantic.StringLiteral{Value: "b"},
					},
				},
				Body: []semantic.Statement{
					&semantic.NativeVariableAssignment{
						Identifier: &semantic.Identifier{Name: "e"},
						Init: &semantic.MemberExpression{
							Object: &semantic.MemberExpression{
								Object:   &semantic.IdentifierExpression{Name: "b"},
								Property: "a",
							},
							Property: "x",
						},
					},
				},
			},
		},
	}

	importer := importer{
		packages: map[string]*interpreter.Package{
			"b": packageB,
		},
	}

	expectedError := fmt.Errorf(`cannot access imported package "a" of imported package "b"`)
	ctx := dependenciestest.Default().Inject(context.Background())
	_, err := interpreter.NewInterpreter(interpreter.NewPackage("")).Eval(ctx, node, values.NewScope(), &importer)

	if err == nil {
		t.Errorf("expected error")
	} else if err.Error() != expectedError.Error() {
		t.Errorf("unexpected result; want err=%v, got err=%v", expectedError, err)
	}
}

// TODO(jlapacik): re-work these tests
/*
func TestInterpreter_EvalPackage(t *testing.T) {
	testcases := []struct {
		name        string
		imports     [](map[string]string)
		pkg         string
		want        values.Object
		sideEffects []values.Value
	}{
		{
			name: "simple",
			pkg: `
				package foo
				a = 1
				b = 2.0
				1 + 1
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(1),
					"b": values.NewFloat(2.0),
				}),
		},
		{
			name: "import",
			imports: []map[string]string{
				{
					"path/to/bar": `
						package bar
						x = 10
`,
				},
			},
			pkg: `
				package foo
				import baz "path/to/bar"
				a = baz.x
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(10),
				}),
		},
		{
			name: "nested variables",
			imports: []map[string]string{
				{
					"path/to/bar": `
						package bar
						f = () => {
							a = 2
							b = 3
							return a + b
						}
`,
				},
			},
			pkg: `
				package foo
				import "path/to/bar"
				a = bar.f()
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(5),
				}),
		},
		{
			name: "polymorphic function",
			imports: []map[string]string{
				{
					"path/to/bar": `
						package bar
						f = (x) => x
`,
				},
			},
			pkg: `
				package foo
				import baz "path/to/bar"
				a = baz.f(x: 10)
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(10),
				}),
		},
		{
			name: "multiple imports",
			imports: []map[string]string{
				{
					"path/to/a": `
						package a
						f = (x) => x
`,
				},
				{
					"path/to/b": `
						package b
						f = (x) => x + "ing"
`,
				},
			},
			pkg: `
				package foo
				import "path/to/a"
				import "path/to/b"

				x = a.f(x: 10)
				y = b.f(x: "str")
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
					"y": values.NewString("string"),
				}),
		},
		{
			name: "nested imports",
			imports: []map[string]string{
				{
					"path/to/a": `
						package a
						f = (x) => x
`,
				},
				{
					"path/to/b": `
						package b
						f = (x) => x + "ing"
`,
				},
				{
					"path/to/c": `
						package c
						import "path/to/a"
						import "path/to/b"
						x = a.f(x: 10)
						y = b.f(x: "str")
`,
				},
			},
			pkg: `
				package foo
				import "path/to/c"

				x = c.x + 10
				y = c.y + "s"
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(20),
					"y": values.NewString("strings"),
				}),
		},
		{
			name: "main package",
			pkg: `
				package main
				x = 10
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				}),
		},
		{
			name: "side effect",
			pkg: `
				package foo
				sideEffect()
`,
			sideEffects: []values.Value{
				values.NewInt(0),
			},
		},
		{
			name: "implicit main",
			pkg:  `x = 10`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				}),
		},
		{
			name: "explicit side effect",
			pkg: `
				package foo
				sideEffect()
`,
			sideEffects: []values.Value{
				values.NewInt(0),
			},
		},
		{
			name: "import side effect",
			imports: []map[string]string{
				{
					"path/to/foo": `
						package foo
						sideEffect()
`,
				},
			},
			pkg: `
				package main
				import "path/to/foo"
				x = 10
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				}),
			sideEffects: []values.Value{
				values.NewInt(0), // side effect from `sideEffect()`
			},
		},
	}
	builtins := map[string]values.Value{"sideEffect": &function{
		name: "sideEffect",
		t: semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Required: nil,
			Return:   semantic.Int,
		}),
		call: func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			return values.NewInt(0), nil
		},
		hasSideEffect: true,
	}}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			importer := &importer{
				packages: make(map[string]*interpreter.Package),
			}
			scope := interpreter.NewNestedScope(nil, values.NewObjectWithValues(builtins))
			for _, imp := range tc.imports {
				var path, pkg string
				for k, v := range imp {
					path = k
					pkg = v
				}
				itrp := interpreter.NewInterpreter(context.Background(), executetest.Default())
				if _, err := interptest.Eval(itrp, scope, importer, pkg); err != nil {
					t.Fatal(err)
				}
				importer.packages[path] = itrp.Package()
			}
			itrp := interpreter.NewInterpreter(context.Background(), executetest.Default())
			if err := interptest.Eval(itrp, scope, importer, tc.pkg); err != nil {
				t.Fatal(err)
			}
			got := itrp.Package()
			if tc.want != nil && !got.Equal(tc.want) {
				t.Errorf("unexpected package object -want/+got\n%s", cmp.Diff(tc.want, got))
			}
			sideEffects := got.SideEffects()
			if tc.sideEffects != nil && !cmp.Equal(tc.sideEffects, sideEffects) {
				t.Errorf("unexpected side effects -want/+got\n%s", cmp.Diff(tc.sideEffects, sideEffects))
			}
		})
	}
}
*/

func TestInterpreter_SetNewOption(t *testing.T) {
	pkg := interpreter.NewPackage("alert")
	ctx := dependenciestest.Default().Inject(context.Background())
	itrp := interpreter.NewInterpreter(pkg)
	script := `
		package alert
		option state = "Warning"
		state
`
	if _, err := interptest.Eval(ctx, itrp, values.NewNestedScope(nil, pkg), nil, script); err != nil {
		t.Fatalf("failed to evaluate package: %v", err)
	}
	option, ok := pkg.Get("state")
	if !ok {
		t.Errorf("missing option %q in package %s", "state", "alert")
	}
	if got, want := option.Type().Nature(), semantic.String; want != got {
		t.Fatalf("unexpected option type; want=%s got=%s value: %v", want, got, option)
	}
	if got, want := option.Str(), "Warning"; want != got {
		t.Errorf("unexpected option value; want=%s got=%s", want, got)
	}
}

func TestInterpreter_SetQualifiedOption(t *testing.T) {
	externalPackage := interpreter.NewPackage("alert")
	externalPackage.SetOption("state", values.NewString("Warning"))
	importer := &importer{
		packages: map[string]*interpreter.Package{
			"alert": externalPackage,
		},
	}
	ctx := dependenciestest.Default().Inject(context.Background())
	itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
	pkg := `
		package foo
		import "alert"
		option alert.state = "Error"
		alert.state
`
	if _, err := interptest.Eval(ctx, itrp, values.NewScope(), importer, pkg); err != nil {
		t.Fatalf("failed to evaluate package: %v", err)
	}
	option, ok := externalPackage.Get("state")
	if !ok {
		t.Errorf("missing option %q in package %s", "state", "alert")
	}
	if got, want := option.Type().Nature(), semantic.String; want != got {
		t.Fatalf("unexpected option type; want=%s got=%s value: %v", want, got, option)
	}
	if got, want := option.Str(), "Error"; want != got {
		t.Errorf("unexpected option value; want=%s got=%s", want, got)
	}
}
