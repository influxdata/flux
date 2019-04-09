package querytest

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/stdlib/universe"
)

type NewQueryTestCase struct {
	Name    string
	Raw     string
	Want    *flux.Spec
	WantErr bool
}

var opts = append(
	semantictest.CmpOptions,
	cmp.AllowUnexported(flux.Spec{}),
	cmp.AllowUnexported(universe.JoinOpSpec{}),
	cmpopts.IgnoreUnexported(flux.Spec{}),
	cmpopts.IgnoreUnexported(universe.JoinOpSpec{}),
)

func NewQueryTestHelper(t *testing.T, tc NewQueryTestCase) {
	t.Helper()

	now := time.Now().UTC()
	got, err := spec.FromScript(context.Background(), tc.Raw, now)
	if (err != nil) != tc.WantErr {
		t.Errorf("error compiling spec error = %v, wantErr %v", err, tc.WantErr)
		return
	}
	if tc.WantErr {
		return
	}
	if tc.Want != nil {
		tc.Want.Now = now
		if !cmp.Equal(tc.Want, got, opts...) {
			t.Errorf("unexpected specs -want/+got %s", cmp.Diff(tc.Want, got, opts...))
		}
	}
}

func RunAndCheckResult(t testing.TB, querier *Querier, c flux.Compiler, d flux.Dialect, want string) {
	t.Helper()

	var buf bytes.Buffer
	_, err := querier.Query(context.Background(), &buf, c, d)
	if err != nil {
		t.Errorf("failed to run query: %v", err)
		return
	}

	got := buf.String()

	if g, w := strings.TrimSpace(got), strings.TrimSpace(want); g != w {
		t.Errorf("result not as expected want(-) got (+):\n%v", diff.LineDiff(w, g))
	}
}
