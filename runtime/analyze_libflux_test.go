package runtime_test

import (
	"errors"
	"testing"

	"github.com/influxdata/flux/runtime"
)

func TestAnalyzeSource(t *testing.T) {
	tcs := []struct {
		name string
		flx  string
		err  error
	}{
		{
			name: "success",
			flx:  `x = 10`,
		},
		{
			name: "failure",
			flx:  `x = 10 + "foo"`,
			err:  errors.New("cannot unify int with string"),
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := runtime.AnalyzeSource(tc.flx)
			if err != nil {
				if tc.err == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if want, got := tc.err.Error(), err.Error(); want != got {
					t.Fatalf("wanted error %q, got %q", want, got)
				}
				return
			}
			if tc.err != nil {
				t.Fatalf("expected error %q, got none", tc.err)
			}
		})
	}
}
