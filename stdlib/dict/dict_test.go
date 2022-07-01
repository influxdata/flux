package dict_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/dict"
	"github.com/influxdata/flux/values"
)

func TestFromList(t *testing.T) {
	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"pairs": func() values.Array {
				vals := []values.Value{
					values.NewObjectWithValues(
						map[string]values.Value{
							"key":   values.NewString("a"),
							"value": values.NewInt(4),
						},
					),
					values.NewObjectWithValues(
						map[string]values.Value{
							"key":   values.NewString("b"),
							"value": values.NewInt(8),
						},
					),
				}
				arr := values.NewArray(semantic.NewArrayType(vals[0].Type()))
				for _, v := range vals {
					arr.Append(v)
				}
				return arr
			}(),
		},
	))

	v, err := dict.FromList(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should be a dictionary.
	if want, got := semantic.Dictionary, v.Type().Nature(); want != got {
		t.Fatalf("unexpected nature -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	dict := v.Dict()

	// Should be two entries.
	if want, got := 2, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	got := make(map[string]int64)
	dict.Range(func(key, value values.Value) {
		got[key.Str()] = value.Int()
	})

	want := map[string]int64{
		"a": int64(4),
		"b": int64(8),
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected values -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestGet(t *testing.T) {
	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"dict": func() values.Dictionary {
				dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
				b := values.NewDictBuilder(dictType)
				b.Insert(values.NewString("a"), values.NewInt(4))
				return b.Dict()
			}(),
			"key":     values.NewString("a"),
			"default": values.NewInt(0),
		},
	))

	v, err := dict.Get(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := semantic.Int, v.Type().Nature(); want != got {
		t.Errorf("unexpected nature -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if want, got := int64(4), v.Int(); want != got {
		t.Errorf("unexpected value -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestInsert(t *testing.T) {
	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"dict": func() values.Dictionary {
				dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
				b := values.NewDictBuilder(dictType)
				b.Insert(values.NewString("a"), values.NewInt(4))
				return b.Dict()
			}(),
			"key":   values.NewString("b"),
			"value": values.NewInt(8),
		},
	))

	v, err := dict.Insert(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should be a dictionary.
	if want, got := semantic.Dictionary, v.Type().Nature(); want != got {
		t.Fatalf("unexpected nature -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	dict := v.Dict()

	// Should be two entries.
	if want, got := 2, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	got := make(map[string]int64)
	dict.Range(func(key, value values.Value) {
		got[key.Str()] = value.Int()
	})

	want := map[string]int64{
		"a": int64(4),
		"b": int64(8),
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected values -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestRemove(t *testing.T) {
	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"dict": func() values.Dictionary {
				dictType := semantic.NewDictType(semantic.BasicString, semantic.BasicInt)
				b := values.NewDictBuilder(dictType)
				b.Insert(values.NewString("a"), values.NewInt(4))
				b.Insert(values.NewString("b"), values.NewInt(8))
				b.Insert(values.NewString("c"), values.NewInt(12))
				return b.Dict()
			}(),
			"key": values.NewString("c"),
		},
	))

	v, err := dict.Remove(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should be a dictionary.
	if want, got := semantic.Dictionary, v.Type().Nature(); want != got {
		t.Fatalf("unexpected nature -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	dict := v.Dict()

	// Should be two entries.
	if want, got := 2, dict.Len(); want != got {
		t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	got := make(map[string]int64)
	dict.Range(func(key, value values.Value) {
		got[key.Str()] = value.Int()
	})

	want := map[string]int64{
		"a": int64(4),
		"b": int64(8),
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected values -want/+got:\n%s", cmp.Diff(want, got))
	}
}
