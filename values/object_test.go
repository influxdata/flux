package values_test

import (
	"testing"

	"github.com/influxdata/flux/values"
)

func TestObjectEqual(t *testing.T) {
	r := values.NewObject()
	r.Set("a", values.NewInt(1))
	l := values.NewObject()
	l.Set("a", values.NewInt(1))

	if !l.Equal(r) {
		t.Fatal("expected objects to be equal")
	}

	l.Set("a", values.NewInt(2))
	if l.Equal(r) {
		t.Fatal("expected objects to be unequal")
	}

	r.Set("a", values.NewInt(2))
	if !l.Equal(r) {
		t.Fatal("expected objects to be equal")
	}
	l.Set("b", values.NewInt(1))
	if l.Equal(r) {
		t.Fatal("expected objects to be unequal")
	}
}
