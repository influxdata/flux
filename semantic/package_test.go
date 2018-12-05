package semantic_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
)

func TestCreatePackage(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		importer semantic.Importer
		want     semantic.Package
	}{
		{
			name: "simple",
			script: `
package foo

a = 1
b = 2.0

1 + 1
`,
			want: semantic.Package{
				Name: "foo",
				Type: semantic.NewObjectPolyType(
					map[string]semantic.PolyType{
						"a": semantic.Int,
						"b": semantic.Float,
					},
					nil,
					semantic.LabelSet{"a", "b"},
				),
			},
		},
		{
			name: "polymorphic package",
			script: `
package foo

identity = (x) => x
`,
			want: semantic.Package{
				Name: "foo",
				Type: semantic.NewObjectPolyType(
					map[string]semantic.PolyType{
						"identity": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Parameters: map[string]semantic.PolyType{
								"x": semantic.Tvar(3),
							},
							Required: semantic.LabelSet{"x"},
							Return:   semantic.Tvar(3),
						}),
					},
					nil,
					semantic.LabelSet{"identity"},
				),
			},
		},
		{
			name: "nested variables",
			script: `
package bar

a = () => {
	b = 2.0
	return b
}
`,
			want: semantic.Package{
				Name: "bar",
				Type: semantic.NewObjectPolyType(
					map[string]semantic.PolyType{
						"a": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
							Return: semantic.Float,
						}),
					},
					nil,
					semantic.LabelSet{"a"},
				),
			},
		},
		{
			name: "wrap internal package",
			script: `
package baz

import "internal"

a = internal.a
`,
			importer: importer{
				packages: map[string]semantic.Package{
					"internal": semantic.Package{
						Name: "internal",
						Type: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"a": semantic.Int,
								"b": semantic.Float,
							},
							nil,
							semantic.LabelSet{"a", "b"},
						),
					},
				},
			},
			want: semantic.Package{
				Name: "baz",
				Type: semantic.NewObjectPolyType(
					map[string]semantic.PolyType{
						"a": semantic.Int,
					},
					nil,
					semantic.LabelSet{"a"},
				),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			program, err := parser.NewAST(tc.script)
			if err != nil {
				t.Fatal(err)
			}
			node, err := semantic.New(program)
			if err != nil {
				t.Fatal(err)
			}
			got, err := semantic.CreatePackage(node, tc.importer)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected package -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
