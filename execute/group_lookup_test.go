package execute_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	cols = []flux.ColMeta{
		{Label: "a", Type: flux.TString},
		{Label: "b", Type: flux.TString},
		{Label: "c", Type: flux.TString},
	}
	key0 = execute.NewGroupKey(
		cols,
		[]values.Value{
			values.NewString("I"),
			values.NewString("J"),
			values.NewString("K"),
		},
	)
	key1 = execute.NewGroupKey(
		cols,
		[]values.Value{
			values.NewString("L"),
			values.NewString("M"),
			values.NewString("N"),
		},
	)
	key2 = execute.NewGroupKey(
		cols,
		[]values.Value{
			values.NewString("X"),
			values.NewString("Y"),
			values.NewString("Z"),
		},
	)
	key3 = execute.NewGroupKey(
		cols,
		[]values.Value{
			values.NewString("Z"),
			values.NewNull(semantic.String),
			values.NewString("A"),
		},
	)
)

func TestGroupLookup(t *testing.T) {
	l := execute.NewGroupLookup()
	l.Set(key0, 0)
	if v, ok := l.Lookup(key0); !ok || v != 0 {
		t.Error("failed to lookup key0")
	}
	l.Set(key1, 1)
	if v, ok := l.Lookup(key1); !ok || v != 1 {
		t.Error("failed to lookup key1")
	}
	l.Set(key2, 2)
	if v, ok := l.Lookup(key2); !ok || v != 2 {
		t.Error("failed to lookup key2")
	}
	l.Set(key3, 3)
	if v, ok := l.Lookup(key3); !ok || v != 3 {
		t.Error("failed to lookup key3")
	}

	want := []entry{
		{Key: key0, Value: 0},
		{Key: key1, Value: 1},
		{Key: key2, Value: 2},
		{Key: key3, Value: 3},
	}

	var got []entry
	l.Range(func(k flux.GroupKey, v interface{}) {
		got = append(got, entry{
			Key:   k,
			Value: v.(int),
		})
	})

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected range: -want/+got:\n%s", cmp.Diff(want, got))
	}

	l.Set(key0, -1)
	if v, ok := l.Lookup(key0); !ok || v != -1 {
		t.Error("failed to lookup key0 after set")
	}

	l.Delete(key1)
	if _, ok := l.Lookup(key1); ok {
		t.Error("failed to delete key1")
	}
	l.Delete(key0)
	if _, ok := l.Lookup(key0); ok {
		t.Error("failed to delete key0")
	}
	l.Delete(key2)
	if _, ok := l.Lookup(key2); ok {
		t.Error("failed to delete key2")
	}
	l.Delete(key3)
	if _, ok := l.Lookup(key3); ok {
		t.Error("failed to delete key3")
	}
}

// Test that the lookup supports Deletes while rangeing.
func TestGroupLookup_RangeWithDelete(t *testing.T) {
	l := execute.NewGroupLookup()
	l.Set(key0, 0)
	if v, ok := l.Lookup(key0); !ok || v != 0 {
		t.Error("failed to lookup key0")
	}
	l.Set(key1, 1)
	if v, ok := l.Lookup(key1); !ok || v != 1 {
		t.Error("failed to lookup key1")
	}
	l.Set(key2, 2)
	if v, ok := l.Lookup(key2); !ok || v != 2 {
		t.Error("failed to lookup key2")
	}

	want := []entry{
		{Key: key0, Value: 0},
		{Key: key1, Value: 1},
	}
	var got []entry
	l.Range(func(k flux.GroupKey, v interface{}) {
		// Delete the current key
		l.Delete(key0)
		// Delete a future key
		l.Delete(key2)

		got = append(got, entry{
			Key:   k,
			Value: v.(int),
		})
	})
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected range: -want/+got:\n%s", cmp.Diff(want, got))
	}
}

type entry struct {
	Key   flux.GroupKey
	Value int
}
