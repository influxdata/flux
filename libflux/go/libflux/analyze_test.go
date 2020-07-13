package libflux_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/flux/libflux/go/libflux"
)

func TestAnalyze(t *testing.T) {
	tcs := []struct {
		name string
		flx  string
		err  error
	}{
		{
			name: "success",
			flx: `
                package main
                from(bucket: "telegraf")
	              |> range(start: -5m)
	              |> mean()`,
		},
		{
			name: "failure",
			flx:  `x = "foo" + 10`,
			err:  errors.New("type error @1:13-1:15: expected string but found int"),
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ast := libflux.ParseString(tc.flx)
			defer ast.Free()

			sem, err := libflux.Analyze(ast)
			if err != nil {
				if tc.err == nil {
					t.Fatalf("expected no error, got: %q", err)
				}
				if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error: -want/+got: %v", diff)
				}
				return
			}

			if tc.err != nil {
				t.Fatalf("got no error, expected: %q", tc.err)
			}
			fbBuf, err := sem.MarshalFB()
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("flatbuffer has %v bytes\n", len(fbBuf))
		})
	}
}

func TestFindVarType(t *testing.T) {
	tcs := []struct {
		name string
		flx  string
		ty   string
		err  error
	}{
		{
			name: "success",
			flx: `
vint = v.int + 2
f = (v) => v.shadow
g = () => v.sweet
x = g()
vstr = v.str + "hello"`,
			ty: "{int: int | str: string | sweet: t0 | t1}",
		},
		{
			name: "failure",
			flx:  `x = "foo" + 10`,
			err:  errors.New("type error @1:13-1:15: expected string but found int"),
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ast := libflux.ParseString(tc.flx)
			monotype, err := libflux.FindVarType(ast, "v")
			if err != nil {
				if tc.err == nil {
					t.Fatalf("expected no error, got: %q", err)
				}
				if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error: -want/+got: %v", diff)
				}
				return
			}
			if tc.err != nil {
				t.Fatalf("got no error, expected: %q", tc.err)
			}
			if diff := cmp.Diff(tc.ty, monotype.CanonicalString()); diff != "" {
				t.Fatalf("unexpected type: -want/+got: %v", diff)
			}
		})
	}
}
