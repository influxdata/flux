package array_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/runtime"
)

func TestArrayFrom_ReceiveTableObjectIsError(t *testing.T) {
	src := `import "array"
			array.from(rows: array.from(rows: [{}]))`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error calling function \"from\" @2:4-2:44: rows cannot be a table stream; expected an array", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
