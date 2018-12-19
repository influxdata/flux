package builtin_test

import (
	"testing"

	_ "github.com/influxdata/flux/querytest/builtin"
)

func TestBuiltins(t *testing.T) {
	t.Log("Testing that importing testing builtins does not panic")
}
