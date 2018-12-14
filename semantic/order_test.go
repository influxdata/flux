package semantic_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
)

func TestDependencyOrder(t *testing.T) {
	testcases := []struct {
		name string
		prog string
		want string
	}{
		{
			name: "simple",
			prog: `
				package foo
				a = b
				b = 0
`,
			want: `
				package foo
				b = 0
				a = b
`,
		},
		{
			name: "medium",
			prog: `
			    package foo
			    a = b + c
			    b = f()
			    c = f()
			    d = 0
			    f = () => d + 1
`,
			want: `
			    package foo
			    d = 0
			    f = () => d + 1
			    b = f()
			    c = f()
			    a = b + c
`,
		},
		{
			name: "complex",
			prog: `
			    package foo
			   	f = (a, b) => {
					   c = a + b
					   d = a - b
					   return c + d
				   }
				h = 10
				a = f(a: g, b: h)
				g = () => {
					v = () => 0
					return x + v()
				}
				x = 1
`,
			want: `
			    package foo
			   	f = (a, b) => {
					   c = a + b
					   d = a - b
					   return c + d
				   }
				h = 10
				x = 1
				g = () => {
					v = () => 0
					return x + v()
				}
				a = f(a: g, b: h)
`,
		},
		{
			name: "shadow",
			prog: `
			    package foo
			   	a = () => x
				f = () => {
					a = "b"
					return a
				}
				x = 1
`,
			want: `
			    package foo
				f = () => {
					a = "b"
					return a
				}
				x = 1
				a = () => x
`,
		},
		{
			name: "defaults",
			prog: `
			    package foo
				b = 0
			   	f = (a=b, c=d) => a + c
				d = 1
`,
			want: `
			    package foo
				b = 0
				d = 1
			   	f = (a=b, c=d) => a + c
`,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			prog, err := parser.NewAST(tc.prog)
			if err != nil {
				t.Fatal(err)
			}
			node, err := semantic.New(prog)
			if err != nil {
				t.Fatal(err)
			}
			got, err := semantic.OrderVarDependencies(node, nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			wantProg, err := parser.NewAST(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			wantNode, err := semantic.New(wantProg)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(wantNode, got, semantictest.CmpOptions...) {
				t.Errorf("unexpected graph -want/+got\n%s", cmp.Diff(wantNode, got, semantictest.CmpOptions...))
			}
		})
	}
}
