package execute_test

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// This tests that groups keys are sorted lexicographically according to the entire key.
func TestGroupKey_LexicographicOrder(t *testing.T) {
	in := flux.GroupKeys{
		// This key has fewer columns, but is lexicographically > than the other key.
		execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "_f", Type: flux.TString},
				{Label: "_m", Type: flux.TString},
				{Label: "host", Type: flux.TString},
			},
			[]values.Value{
				values.NewString("0.001"),
				values.NewString("query_control_all_duration_seconds"),
				values.NewString("prod1.rsavage.local"),
			},
		),
		// This key is lexicographically < than the previous key.
		execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "_f", Type: flux.TString},
				{Label: "_m", Type: flux.TString},
				{Label: "handler", Type: flux.TString},
				{Label: "host", Type: flux.TString},
			},
			[]values.Value{
				values.NewString("0.001"),
				values.NewString("http_api_request_duration_seconds"),
				values.NewString("platform"),
				values.NewString("prod1.rsavage.local"),
			},
		),
		// This key is the same as the following key but has an additional column
		execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "_f", Type: flux.TString},
				{Label: "_m", Type: flux.TString},
				{Label: "foo", Type: flux.TString},
			},
			[]values.Value{
				values.NewString("0.002"),
				values.NewString("http_api_request_duration_seconds"),
				values.NewString("bar"),
			},
		),
		execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "_f", Type: flux.TString},
				{Label: "_m", Type: flux.TString},
			},
			[]values.Value{
				values.NewString("0.002"),
				values.NewString("http_api_request_duration_seconds"),
			},
		),
	}
	exp := flux.GroupKeys{in[1], in[0], in[3], in[2]}
	sort.Sort(in)

	if got := in; !reflect.DeepEqual(got, exp) {
		t.Fatalf("got keys %s\n, expected %s", got, exp)
	}
}

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

func testGroupLookup_LookupOrCreate(tb testing.TB, run func(name string, keys []flux.GroupKey)) {
	// Generate a small schema to use for the group lookup.
	schema := gen.Schema{
		Start: time.Now(),
		Tags: []gen.Tag{ // total cardinality is 105
			{Name: "_measurement", Cardinality: 1},
			{Name: "_field", Cardinality: 5},
			{Name: "t0", Cardinality: 3},
			{Name: "t1", Cardinality: 7},
		},
		NumPoints: 1,
		Types: map[flux.ColType]int{
			flux.TFloat: 1,
		},
		Seed: func(x int64) *int64 { return &x }(0),
	}
	input, err := gen.Input(schema)
	if err != nil {
		tb.Fatalf("unexpected error: %s", err)
	}

	// There should be one point per key so read from the input
	// and put them into a slice.
	canonicalKeys := make([]flux.GroupKey, 0, 105)
	for input.More() {
		res := input.Next()
		if err := res.Tables().Do(func(table flux.Table) error {
			canonicalKeys = append(canonicalKeys, table.Key())
			return nil
		}); err != nil {
			tb.Fatalf("unexpected error: %s", err)
		}
	}

	if len(canonicalKeys) != 105 {
		tb.Fatalf("unexpected number of keys: %d", len(canonicalKeys))
	}

	run("Ordered", func() []flux.GroupKey {
		keys := make([]flux.GroupKey, len(canonicalKeys))
		copy(keys, canonicalKeys)
		sort.Sort(flux.GroupKeys(keys))
		return keys
	}())

	run("Interweaved", func() []flux.GroupKey {
		keys := make([]flux.GroupKey, len(canonicalKeys))
		copy(keys, canonicalKeys)
		sort.Sort(flux.GroupKeys(keys))

		// Fisher-yates shuffle with groups of 5.
		tmp := make([]flux.GroupKey, 5)
		for i := 100; i > 0; i -= 5 {
			j := rand.Intn(i/5+1) * 5
			if i == j {
				continue
			}
			// Copy the current group into the tmp slice.
			copy(tmp, keys[i:i+5])
			// Copy the chosen range into the current group.
			copy(keys[i:], keys[j:j+5])
			// Copy the temporary keys into the chosen one.
			copy(keys[j:], tmp)
		}
		return keys
	}())

	run("Unordered", func() []flux.GroupKey {
		keys := make([]flux.GroupKey, len(canonicalKeys))
		copy(keys, canonicalKeys)

		// Fisher-yates shuffle.
		for i := len(keys) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			keys[i], keys[j] = keys[j], keys[i]
		}
		return keys
	}())

	run("Reversed", func() []flux.GroupKey {
		keys := make([]flux.GroupKey, len(canonicalKeys))
		copy(keys, canonicalKeys)

		// Reverse the array.
		for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
			keys[i], keys[j] = keys[j], keys[i]
		}
		return keys
	}())

	run("ManySplits", func() []flux.GroupKey {
		keys := make([]flux.GroupKey, len(canonicalKeys))
		copy(keys, canonicalKeys)

		// Swap every 2nd element with every 3rd element to
		// force a split to occur.
		for i := 0; i < len(keys)-2; i += 3 {
			keys[i+1], keys[i+2] = keys[i+2], keys[i+1]
		}
		return keys
	}())
}

func TestGroupLookup_LookupOrCreate(t *testing.T) {
	testGroupLookup_LookupOrCreate(t, func(name string, keys []flux.GroupKey) {
		t.Run(name, func(t *testing.T) {
			l := execute.NewGroupLookup()
			for _, key := range keys {
				if _, ok := l.Lookup(key); !ok {
					l.Set(key, true)
				} else {
					t.Errorf("unexpected key lookup: %s", key)
				}
			}

			// Lookup should work on everything.
			for _, key := range keys {
				if _, ok := l.Lookup(key); !ok {
					t.Errorf("key lookup failed: %s", key)
				}
			}
		})
	})
}

func BenchmarkGroupLookup_LookupOrCreate(b *testing.B) {
	testGroupLookup_LookupOrCreate(b, func(name string, keys []flux.GroupKey) {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				l := execute.NewGroupLookup()
				for _, key := range keys {
					if _, ok := l.Lookup(key); !ok {
						l.Set(key, true)
					}
				}
			}
		})
	})
}
