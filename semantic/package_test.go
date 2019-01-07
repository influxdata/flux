package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/ast"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
)

func TestCreatePackage(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		importer semantic.Importer
		want     semantic.PackageType
		wantErr  bool
		skip     bool
	}{
		{
			name: "simple",
			script: `
package foo

a = 1
b = 2.0

1 + 1
`,
			want: semantic.PackageType{
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
			want: semantic.PackageType{
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
			want: semantic.PackageType{
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
				packages: map[string]semantic.PackageType{
					"internal": semantic.PackageType{
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
			want: semantic.PackageType{
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
		{
			name: "qualified option",
			script: `
package foo

import "alert"

option alert.state = "Warning"
`,
			importer: importer{
				packages: map[string]semantic.PackageType{
					"alert": semantic.PackageType{
						Name: "alert",
						Type: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"state": semantic.String,
							},
							nil,
							semantic.LabelSet{"state"},
						),
					},
				},
			},
			want: semantic.PackageType{
				Name: "foo",
				Type: semantic.NewEmptyObjectPolyType(),
			},
		},
		{
			name: "assign qualified option new type",
			script: `
package foo

import "alert"

option alert.state = 0
`,
			importer: importer{
				packages: map[string]semantic.PackageType{
					"alert": semantic.PackageType{
						Name: "alert",
						Type: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"state": semantic.String,
							},
							nil,
							semantic.LabelSet{"state"},
						),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "modify exported identifier",
			script: `
package foo

import "bar"

bar.x = 10
`,
			importer: importer{
				packages: map[string]semantic.PackageType{
					"bar": semantic.PackageType{
						Name: "bar",
						Type: semantic.NewObjectPolyType(
							map[string]semantic.PolyType{
								"x": semantic.Int,
							},
							nil,
							semantic.LabelSet{"x"},
						),
					},
				},
			},
			wantErr: true,
			skip:    true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			pkg := parser.ParseSource(tc.script)
			if ast.Check(pkg) > 0 {
				t.Fatal(ast.GetError(pkg))
			}
			node, err := semantic.New(pkg)
			if err != nil {
				t.Fatal(err)
			}
			got, err := semantic.CreatePackage(node, tc.importer)
			if !tc.wantErr {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				if !cmp.Equal(tc.want, got) {
					t.Errorf("unexpected package -want/+got\n%s", cmp.Diff(tc.want, got))
				}
			} else if err == nil {
				t.Errorf("expected error")
			}
		})
	}
}
