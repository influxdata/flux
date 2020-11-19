package values_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestDict_Type(t *testing.T) {
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	dict := values.NewDict(dictType)

	// Dictionary should return a valid dictionary type.
	if want, got := dictType, dict.Type(); !cmp.Equal(want, got) {
		t.Errorf("unexpected type -want/+got:\n%s", cmp.Diff(want, got))
	}

	b := values.NewDictBuilder(dictType)
	dict = b.Dict()

	// Should have a valid dictionary type when using the builder.
	if want, got := dictType, dict.Type(); !cmp.Equal(want, got) {
		t.Errorf("unexpected type -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Should continue to have a valid dictionary type after an insert.
	dict, _ = dict.Insert(values.NewInt(2), values.NewInt(4))
	if want, got := dictType, dict.Type(); !cmp.Equal(want, got) {
		t.Errorf("unexpected type -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Should continue to have a valid dictionary type after a removal.
	dict = dict.Remove(values.NewInt(2))
	if want, got := dictType, dict.Type(); !cmp.Equal(want, got) {
		t.Errorf("unexpected type -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestDict_Get(t *testing.T) {
	// Insert values to ensure this works with multiple values.
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	b.Insert(values.NewInt(0), values.NewInt(0))
	b.Insert(values.NewInt(1), values.NewInt(1))
	b.Insert(values.NewInt(2), values.NewInt(1))
	b.Insert(values.NewInt(3), values.NewInt(2))
	b.Insert(values.NewInt(4), values.NewInt(3))
	b.Insert(values.NewInt(5), values.NewInt(5))
	b.Insert(values.NewInt(6), values.NewInt(8))
	b.Insert(values.NewInt(7), values.NewInt(13))
	dict := b.Dict()

	// Retrieve existing value.
	if want, got := values.NewInt(8), dict.Get(values.NewInt(6), values.Null); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Retrieve default value when non-existant value is retrieved.
	if want, got := values.NewInt(-1), dict.Get(values.NewInt(8), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestDict_Insert(t *testing.T) {
	// Insert a single value into the dictionary.
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	_ = b.Insert(values.NewInt(7), values.NewInt(13))
	dict := b.Dict()

	// Retrieve a non-existant value.
	if want, got := values.NewInt(-1), dict.Get(values.NewInt(6), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Insert that value and check that it exists.
	dict2, _ := dict.Insert(values.NewInt(6), values.NewInt(8))
	if want, got := values.NewInt(8), dict2.Get(values.NewInt(6), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// The previous dictionary should not be changed.
	if want, got := values.NewInt(-1), dict.Get(values.NewInt(6), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Replace an existing value.
	dict3, _ := dict2.Insert(values.NewInt(7), values.NewInt(20))
	if want, got := values.NewInt(20), dict3.Get(values.NewInt(7), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// The previous dictionary should not be changed.
	if want, got := values.NewInt(13), dict2.Get(values.NewInt(7), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// Attempts to insert null return an error.
	if _, err := dict3.Insert(values.Null, values.NewInt(0)); err == nil {
		t.Error("expected error")
	}
}

func TestDict_Remove(t *testing.T) {
	// Insert two values into the dictionary.
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	_ = b.Insert(values.NewInt(6), values.NewInt(8))
	_ = b.Insert(values.NewInt(7), values.NewInt(13))
	dict := b.Dict()

	// Remove one value.
	dict2 := dict.Remove(values.NewInt(6))

	// It should no longer be present in the dictionary.
	if want, got := values.NewInt(-1), dict2.Get(values.NewInt(6), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// The other value should be present.
	if want, got := values.NewInt(13), dict2.Get(values.NewInt(7), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}

	// If we check the original dictionary, the value is still present.
	if want, got := values.NewInt(8), dict.Get(values.NewInt(6), values.NewInt(-1)); !want.Equal(got) {
		t.Errorf("unexpected value -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestDict_Range(t *testing.T) {
	dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	b.Insert(values.NewString("a"), values.NewInt(2))
	b.Insert(values.NewString("b"), values.NewInt(6))
	b.Insert(values.NewString("c"), values.NewInt(4))
	dict := b.Dict()

	want := map[string]int64{
		"a": 2,
		"b": 6,
		"c": 4,
	}
	dict.Range(func(key, value values.Value) {
		if want, got := want[key.Str()], value.Int(); want != got {
			t.Errorf("unexpected value -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		delete(want, key.Str())
	})

	if len(want) > 0 {
		t.Errorf("some values were not checked: %v", want)
	}
}

func TestDict_Len(t *testing.T) {
	dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	b.Insert(values.NewString("a"), values.NewInt(2))
	b.Insert(values.NewString("b"), values.NewInt(6))
	b.Insert(values.NewString("c"), values.NewInt(4))
	dict := b.Dict()

	// The starting length should be 3.
	if want, got := 3, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Insert a value and check the length.
	dict2, _ := dict.Insert(values.NewString("d"), values.NewInt(3))
	if want, got := 4, dict2.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// The original length should not change.
	if want, got := 3, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Removing an element changes the length.
	dict3 := dict.Remove(values.NewString("c"))
	if want, got := 2, dict3.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// The original length should not change.
	if want, got := 3, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestDict_Equal(t *testing.T) {
	dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
	b := values.NewDictBuilder(dictType)
	b.Insert(values.NewString("a"), values.NewInt(2))
	b.Insert(values.NewString("b"), values.NewInt(6))
	b.Insert(values.NewString("c"), values.NewInt(4))
	dict := b.Dict()

	// Should equal itself.
	if !dict.Equal(dict) {
		t.Error("expected values to be equal")
	}

	// Should not equal other values.
	if dict.Equal(values.NewString("a")) {
		t.Error("expected values to be not equal")
	}

	// Insert a value and they should not be equal.
	dict2, _ := dict.Insert(values.NewString("d"), values.NewInt(5))
	if dict.Equal(dict2) {
		t.Error("expected values to be not equal")
	}

	// Remove the value and they should be equal again.
	dict3 := dict2.Remove(values.NewString("d"))
	if !dict.Equal(dict3) {
		t.Error("expected values to be equal")
	}

	// Overwrite an existing value and they should not be equal.
	dict4, _ := dict.Insert(values.NewString("c"), values.NewInt(0))
	if dict.Equal(dict4) {
		t.Error("expected values to be not equal")
	}
}

var benchmarkKeys []values.Value

func init() {
	benchmarkKeys = make([]values.Value, 0, 100)

	gen := rand.New(rand.NewSource(0))
	for i := 0; i < 100; i++ {
		key := values.NewInt(gen.Int63())
		benchmarkKeys = append(benchmarkKeys, key)
	}
}

func BenchmarkDict_Get(b *testing.B) {
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	dict := values.NewDict(dictType)
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, key := range benchmarkKeys {
		value := values.NewInt(gen.Int63())
		dict, _ = dict.Insert(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	def := values.NewInt(0)
	for i := 0; i < b.N; i++ {
		for _, key := range benchmarkKeys {
			dict.Get(key, def)
		}
	}
}

func BenchmarkDict_Insert(b *testing.B) {
	// We're going to insert repeatedly to the dictionary
	// with random pre-determined values to random locations.
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))

	dvalues := make([]values.Value, len(benchmarkKeys))
	for i := 0; i < len(benchmarkKeys); i++ {
		dvalues[i] = values.NewInt(gen.Int63())
	}
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dict := values.NewDict(dictType)
		for i, key := range benchmarkKeys {
			dict, _ = dict.Insert(key, dvalues[i])
		}
	}
}

func BenchmarkDict_Remove(b *testing.B) {
	// We're going to insert values to each of the benchmark
	// keys and then determine an order to remove them.
	dictType := semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)
	dict := values.NewDict(dictType)
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate random values for each of the benchmark keys.
	for _, key := range benchmarkKeys {
		value := values.NewInt(gen.Int63())
		dict, _ = dict.Insert(key, value)
	}

	// Determine an order to remove them. We use a pre-determined
	// random order to give an idea of how the dictionary deals
	// with random access while minimizing the affect the random
	// number generator has on timing.
	// The order is determined by doing a Fisher-Yates shuffle.
	indices := make([]int, len(benchmarkKeys))
	for i := range indices {
		indices[i] = i
	}
	for i := len(indices) - 1; i > 0; i-- {
		j := gen.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	keys := make([]values.Value, len(indices))
	for i, idx := range indices {
		keys[i] = benchmarkKeys[idx]
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := dict
		for _, key := range keys {
			d = d.Remove(key)
		}
	}
}
