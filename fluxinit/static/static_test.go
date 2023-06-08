package static_test

import (
	"testing"

	_ "github.com/InfluxCommunity/flux/fluxinit/static"
)

func TestBuiltins(t *testing.T) {
	t.Log("Testing that importing builtins does not panic")
}
