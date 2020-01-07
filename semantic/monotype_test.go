package semantic_test

import (
	"testing"

	"github.com/influxdata/flux/semantic"
)

func assertEq(t *testing.T, want, got string) {
	if want != got {
		t.Errorf("expected %q but got %q", want, got)
	}
}

func TestBasicTypes(t *testing.T) {
	assertEq(t, "bool", semantic.BasicBool.String())
	assertEq(t, "int", semantic.BasicInt.String())
	assertEq(t, "uint", semantic.BasicUint.String())
	assertEq(t, "float", semantic.BasicFloat.String())
	assertEq(t, "string", semantic.BasicString.String())
	assertEq(t, "duration", semantic.BasicDuration.String())
	assertEq(t, "time", semantic.BasicTime.String())
	assertEq(t, "regexp", semantic.BasicRegexp.String())
	assertEq(t, "bytes", semantic.BasicBytes.String())
}
