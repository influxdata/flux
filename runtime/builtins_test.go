package runtime

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestValidatePackageBuiltins(t *testing.T) {
	testCases := []struct {
		name   string
		pkg    map[string]values.Value
		semPkg *semantic.Package
		err    error
	}{
		{
			name: "no errors",
			pkg: map[string]values.Value{
				"foo": values.NewInt(0),
			},
			semPkg: &semantic.Package{
				Files: []*semantic.File{{
					Body: []semantic.Statement{
						&semantic.BuiltinStatement{
							ID: &semantic.Identifier{Name: "foo"},
						},
					},
				}},
			},
		},
		{
			name: "extra values",
			pkg: map[string]values.Value{
				"foo": values.NewInt(0),
			},
			semPkg: &semantic.Package{},
			err:    errors.New("missing builtin values [], extra builtin values [foo]"),
		},
		{
			name: "missing values",
			pkg:  map[string]values.Value{},
			semPkg: &semantic.Package{
				Files: []*semantic.File{{
					Body: []semantic.Statement{
						&semantic.BuiltinStatement{
							ID: &semantic.Identifier{Name: "foo"},
						},
					},
				}},
			},
			err: errors.New("missing builtin values [foo], extra builtin values []"),
		},
		{
			name: "missing and values",
			pkg: map[string]values.Value{
				"foo": values.NewInt(0),
				"bar": values.NewInt(0),
			},
			semPkg: &semantic.Package{
				Files: []*semantic.File{{
					Body: []semantic.Statement{
						&semantic.BuiltinStatement{
							ID: &semantic.Identifier{Name: "foo"},
						},
						&semantic.BuiltinStatement{
							ID: &semantic.Identifier{Name: "baz"},
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
			err := validatePackageBuiltins(tc.pkg, tc.semPkg)
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
