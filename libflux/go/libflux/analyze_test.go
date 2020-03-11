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
			err:  errors.New("cannot unify string with int"),
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
