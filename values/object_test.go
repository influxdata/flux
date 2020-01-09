package values_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/values"
)

func TestObjectEqual(t *testing.T) {
	r := values.NewObjectWithValues(map[string]values.Value{
		"a": values.NewInt(1),
	})
	l := values.NewObjectWithValues(map[string]values.Value{
		"a": values.NewInt(1),
	})

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

	l, _ = values.BuildObject(func(set values.ObjectSetter) error {
		l.Range(func(name string, v values.Value) {
			set(name, v)
		})
		set("b", values.NewInt(1))
		return nil
	})
	if l.Equal(r) {
		t.Fatal("expected objects to be unequal")
	}
}

func TestBuildObject(t *testing.T) {
	object, err := values.BuildObject(func(set values.ObjectSetter) error {
		set("b", values.NewInt(2))
		set("a", values.NewString("foo"))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 2, object.Len(); want != got {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	got := make(map[string]values.Value)
	object.Range(func(name string, v values.Value) {
		got[name] = v
	})

	want := map[string]values.Value{
		"b": values.NewInt(2),
		"a": values.NewString("foo"),
	}
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected values -want/+got:\n%s", cmp.Diff(want, got))
	}
}
