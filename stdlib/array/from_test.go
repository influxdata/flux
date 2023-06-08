package array_test

import (
	"context"
	"testing"

	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/runtime"
)

func TestArrayFrom_ReceiveTableObjectIsError(t *testing.T) {
	src := `import "array"
			array.from(rows: array.from(rows: [{}]))`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error @2:21-2:43: expected [{}] (array) but found stream[{}] (argument rows)", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
