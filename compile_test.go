package flux_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
)

func TestCompile(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	for _, tc := range []struct {
		q  string
		ok bool
	}{
		{q: `from(bucket: "foo")`, ok: true},
		{q: `t=""t.t`},
		{q: `t=0t.s`},
		{q: `x = from(bucket: "foo")`},
		{q: `x = from(bucket: "foo") |> yield()`, ok: true},
		{q: `from(bucket: "foo")`, ok: true},
	} {
		_, err := flux.Compile(ctx, tc.q, now)
		if tc.ok && err != nil {
			t.Errorf("expected query %q to compile successfully but got error %v", tc.q, err)
		} else if !tc.ok && err == nil {
			t.Errorf("expected query %q to compile with error but got no error", tc.q)
		}
	}
}
