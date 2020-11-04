package static_test

import (
	"testing"

	_ "github.com/influxdata/flux/fluxinit/static"
)

func TestBuiltins(t *testing.T) {
	t.Log("Testing that importing builtins does not panic")
}
