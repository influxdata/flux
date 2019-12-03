package flux

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

func TestValidatePackageBuiltins(t *testing.T) {
	testCases := []struct {
		name   string
		pkg    *interpreter.Package
		astPkg *ast.Package
		err    error
	}{
		{
			name: "no errors",
			pkg: interpreter.NewPackageWithValues("test", values.NewObjectWithValues(map[string]values.Value{
				"foo": values.NewInt(0),
			})),
			astPkg: &ast.Package{
				Files: []*ast.File{{
					Body: []ast.Statement{
						&ast.BuiltinStatement{
							ID: &ast.Identifier{Name: "foo"},
						},
					},
				}},
			},
		},
		{
			name: "extra values",
			pkg: interpreter.NewPackageWithValues("test", values.NewObjectWithValues(map[string]values.Value{
				"foo": values.NewInt(0),
			})),
			astPkg: &ast.Package{},
			err:    errors.New("missing builtin values [], extra builtin values [foo]"),
		},
		{
			name: "missing values",
			pkg:  interpreter.NewPackageWithValues("test", values.NewObject()),
			astPkg: &ast.Package{
				Files: []*ast.File{{
					Body: []ast.Statement{
						&ast.BuiltinStatement{
							ID: &ast.Identifier{Name: "foo"},
						},
					},
				}},
			},
			err: errors.New("missing builtin values [foo], extra builtin values []"),
		},
		{
			name: "missing and values",
			pkg: interpreter.NewPackageWithValues("test", values.NewObjectWithValues(map[string]values.Value{
				"foo": values.NewInt(0),
				"bar": values.NewInt(0),
			})),
			astPkg: &ast.Package{
				Files: []*ast.File{{
					Body: []ast.Statement{
						&ast.BuiltinStatement{
							ID: &ast.Identifier{Name: "foo"},
						},
						&ast.BuiltinStatement{
							ID: &ast.Identifier{Name: "baz"},
						},
					},
				}},
			},
			err: errors.New("missing builtin values [baz], extra builtin values [bar]"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validatePackageBuiltins(tc.pkg, tc.astPkg)
			switch {
			case err == nil && tc.err == nil:
				// Test passes
			case err == nil && tc.err != nil:
				t.Errorf("expected error %v", tc.err)
			case err != nil && tc.err == nil:
				t.Errorf("unexpected error %v", err)
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Errorf("differing error messages -want/+got:\n%s", cmp.Diff(tc.err.Error(), err.Error()))
				}
				// else test passes
			}
		})
	}
}

func Test_options(t *testing.T) {
	pkg := &ast.Package{
		Files: []*ast.File{
			{
				Body: []ast.Statement{
					&ast.VariableAssignment{},
					&ast.OptionStatement{},
				},
			},
			{
				Body: []ast.Statement{
					&ast.VariableAssignment{},
				},
			},
		},
	}

	actual := options(pkg)
	if len(actual.Files) != 1 || len(actual.Files[0].Body) != 1 {
		t.Fail()
	}
}
