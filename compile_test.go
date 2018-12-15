package flux_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

func TestCompile(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	for _, tc := range []struct {
		q  string
		ok bool
	}{
		{q: "0/0", ok: true},
		{q: `t=""t.t`},
		{q: `t=0t.s`},
	} {
		_, err := flux.Compile(ctx, tc.q, now)
		if tc.ok && err != nil {
			t.Errorf("expected query %q to compile successfully but got error %v", tc.q, err)
		} else if !tc.ok && err == nil {
			t.Errorf("expected query %q to compile with error but got no error", tc.q)
		}
	}
}

func init() {
	flux.RegisterBuiltInValueWithPackage("testValue", "exp", values.NewInt(10), false)
	flux.RegisterBuiltInValueWithPackage("testValueWithSideEffect", "exp", values.NewInt(11), true)
}

func TestPackageRegistration(t *testing.T) {
	testcases := []struct {
		name        string
		prog        string
		want        values.Object
		sideEffects []values.Value
	}{
		{
			name: "import from std lib",
			prog: `
				package foo
				import "exp"
				x = exp.testValue
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				},
			),
		},
		{
			name: "import alias",
			prog: `
				package foo
				import e "exp"
				x = e.testValue
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				},
			),
		},
		{
			name: "main package",
			prog: `
				import "exp"
				x = exp.testValue
`,
			want: values.NewObjectWithValues(
				map[string]values.Value{
					"x": values.NewInt(10),
				},
			),
		},
		{
			name: "side effect",
			prog: `
				package foo
				import "exp"
`,
			sideEffects: []values.Value{values.NewInt(11)},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			itrp := flux.NewInterpreter()
			if err := flux.Eval(itrp, tc.prog); err != nil {
				t.Fatal(err)
			}
			if tc.want != nil && !tc.want.Equal(itrp.Package()) {
				got := values.NewObject()
				itrp.Package().Range(func(k string, v values.Value) {
					got.Set(k, v)
				})
				t.Errorf("unexpected result -want/+got\n%s", cmp.Diff(tc.want, got))
			}
			sideEffects := itrp.Package().SideEffects()
			if tc.sideEffects != nil && !cmp.Equal(tc.sideEffects, sideEffects) {
				t.Errorf("unexpected side effects -want/+got\n%s", cmp.Diff(tc.sideEffects, sideEffects))
			}
		})
	}
}
