package array_test

import (
	"testing"

	apachearray "github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/array"
	fluxmemory "github.com/influxdata/flux/memory"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(b *array.StringBuilder)
		bsz   int
		sz    int
		want  []interface{}
	}{
		{
			name: "Constant",
			build: func(b *array.StringBuilder) {
				for i := 0; i < 10; i++ {
					b.Append("abcdefghij")
				}
			},
			bsz: 64, // 64 bytes data.
			sz:  64, // The minimum size of a buffer is 64 bytes
			want: []interface{}{
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
				"abcdefghij",
			},
		},
		{
			name: "RLE",
			build: func(b *array.StringBuilder) {
				for i := 0; i < 5; i++ {
					b.Append("a")
				}
				for i := 0; i < 5; i++ {
					b.Append("b")
				}
			},
			bsz: 192,
			sz:  192,
			want: []interface{}{
				"a", "a", "a", "a", "a",
				"b", "b", "b", "b", "b",
			},
		},
		{
			name: "Random",
			build: func(b *array.StringBuilder) {
				for _, v := range []string{"a", "b", "c", "d", "e"} {
					b.Append(v)
				}
				b.AppendNull()
				for _, v := range []string{"g", "h", "i", "j"} {
					b.Append(v)
				}
			},
			bsz: 192,
			sz:  192,
			want: []interface{}{
				"a", "b", "c", "d", "e",
				nil, "g", "h", "i", "j",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			// Construct a string builder, resize it to a capacity larger than
			// what is required, then verify we see that value.
			b := array.NewStringBuilder(mem)
			b.Resize(len(tc.want) + 2)
			if want, got := len(tc.want)+2, b.Cap(); want != got {
				t.Errorf("unexpected builder cap -want/+got:\n\t- %d\n\t+ %d", want, got)
			}

			// Build the array using the string builder.
			tc.build(b)

			// Verify the string builder attributes.
			if want, got := countNulls(tc.want), b.NullN(); want != got {
				t.Errorf("unexpected builder null count -want/+got:\n\t- %d\n\t+ %d", want, got)
			}
			if want, got := len(tc.want), b.Len(); want != got {
				t.Errorf("unexpected builder len -want/+got:\n\t- %d\n\t+ %d", want, got)
			}
			if want, got := len(tc.want)+2, b.Cap(); want != got {
				t.Errorf("unexpected builder cap -want/+got:\n\t- %d\n\t+ %d", want, got)
			}
			assert.Equal(t, tc.bsz, mem.CurrentAlloc(), "unexpected memory allocation.")

			arr := b.NewStringArray()
			defer arr.Release()
			assert.Equal(t, tc.sz, mem.CurrentAlloc(), "unexpected memory allocation.")

			if want, got := len(tc.want), arr.Len(); want != got {
				t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
			}

			for i, sz := 0, arr.Len(); i < sz; i++ {
				if arr.IsValid(i) == arr.IsNull(i) {
					t.Errorf("valid and null checks are not consistent for index %d", i)
				}

				if tc.want[i] == nil {
					if arr.IsValid(i) {
						t.Errorf("unexpected value -want/+got:\n\t- %v\n\t+ %v", tc.want[i], arr.Value(i))
					}
				} else if arr.IsNull(i) {
					t.Errorf("unexpected value -want/+got:\n\t- %v\n\t+ %v", tc.want[i], nil)
				} else {
					want, got := tc.want[i].(string), arr.Value(i)
					if want != got {
						t.Errorf("unexpected value -want/+got:\n\t- %v\n\t+ %v", want, got)
					}
				}
			}
		})
	}
}

func TestNewStringFromBinaryArray(t *testing.T) {
	alloc := fluxmemory.NewResourceAllocator(nil)
	// Need to use the Apache binary builder to be able to create an actual
	// Arrow Binary array.
	sb := apachearray.NewBinaryBuilder(alloc, array.StringType)
	vals := []string{"a", "b", "c"}
	for _, v := range vals {
		sb.AppendString(v)
	}
	a := sb.NewArray()
	s := array.NewStringFromBinaryArray(a.(*apachearray.Binary))
	if want, got := len(vals), s.Len(); want != got {
		t.Errorf("wanted length of %v, got %v", want, got)
		t.Fail()
	}
	for i, v := range vals {
		if want, got := v, s.Value(i); want != got {
			t.Errorf("at offset %v, wanted %v, got %v", i, want, got)
		}
	}

	a.Release()
	s.Release()

	if want, got := int64(0), alloc.Allocated(); want != got {
		t.Errorf("epxected allocated to be %v, was %v", want, got)
	}
}

func TestStringBuilder_NewArray(t *testing.T) {
	// Reuse the same builder over and over and ensure
	// it is using the proper amount of memory.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	b := array.NewStringBuilder(mem)

	for i := 0; i < 3; i++ {
		b.Resize(10)
		b.ReserveData(10)
		for i := 0; i < 10; i++ {
			b.Append("a")
		}

		arr := b.NewArray()
		assert.Equal(t, 64, mem.CurrentAlloc(), "unexpected memory allocation.")
		arr.Release()
		mem.AssertSize(t, 0)

		b.Resize(10)
		b.ReserveData(10)
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				b.Append("a")
			} else {
				b.Append("b")
			}
		}
		arr = b.NewArray()
		assert.Equal(t, 192, mem.CurrentAlloc(), "unexpected memory allocation.")
		arr.Release()
		mem.AssertSize(t, 0)
	}
}

func TestSlice(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(mem memory.Allocator) array.Array
		i, j  int
		want  []interface{}
	}{
		{
			name: "Int",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewIntBuilder(mem)
				for i := 0; i < 10; i++ {
					if i == 6 {
						b.AppendNull()
						continue
					}
					b.Append(int64(i))
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				int64(5), nil, int64(7), int64(8), int64(9),
			},
		},
		{
			name: "Uint",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewUintBuilder(mem)
				for i := 0; i < 10; i++ {
					if i == 6 {
						b.AppendNull()
						continue
					}
					b.Append(uint64(i))
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				uint64(5), nil, uint64(7), uint64(8), uint64(9),
			},
		},
		{
			name: "Float",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewFloatBuilder(mem)
				for i := 0; i < 10; i++ {
					if i == 6 {
						b.AppendNull()
						continue
					}
					b.Append(float64(i))
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				float64(5), nil, float64(7), float64(8), float64(9),
			},
		},
		{
			name: "String_Constant",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewStringBuilder(mem)
				for i := 0; i < 10; i++ {
					b.Append("a")
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				"a", "a", "a", "a", "a",
			},
		},
		{
			name: "String_RLE",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewStringBuilder(mem)
				for i := 0; i < 5; i++ {
					b.Append("a")
				}
				for i := 0; i < 5; i++ {
					b.Append("b")
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				"b", "b", "b", "b", "b",
			},
		},
		{
			name: "String_Random",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewStringBuilder(mem)
				for _, v := range []string{"a", "b", "c", "d", "e"} {
					b.Append(v)
				}
				b.AppendNull()
				for _, v := range []string{"g", "h", "i", "j"} {
					b.Append(v)
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				nil, "g", "h", "i", "j",
			},
		},
		{
			name: "Boolean",
			build: func(mem memory.Allocator) array.Array {
				b := array.NewBooleanBuilder(mem)
				for i := 0; i < 10; i++ {
					if i == 6 {
						b.AppendNull()
						continue
					}
					b.Append(i%2 == 0)
				}
				return b.NewArray()
			},
			i: 5,
			j: 10,
			want: []interface{}{
				false, nil, false, true, false,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			arr := tc.build(mem)
			slice := array.Slice(arr, tc.i, tc.j)
			arr.Release()
			defer slice.Release()

			if want, got := len(tc.want), slice.Len(); want != got {
				t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
			}

			for i, sz := 0, slice.Len(); i < sz; i++ {
				want, got := tc.want[i], getValue(slice, i)
				if want != got {
					t.Errorf("unexpected value -want/+got:\n\t- %v\n\t+ %v", want, got)
				}
			}
		})
	}
}

func TestCopyValid(t *testing.T) {
	for _, tc := range []struct {
		name       string
		input      func(memory.Allocator) *array.Int
		nullBitmap func(memory.Allocator) *array.Int
		want       []interface{}
	}{
		{
			name: "nulls",
			input: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.AppendNull()
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(5)
				b.AppendNull()
				return b.NewIntArray()
			},
			want: []interface{}{int64(2), int64(3), int64(5)},
		},
		{
			name: "nulls with nullBitmap",
			input: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.AppendNull()
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(5)
				b.AppendNull()
				return b.NewIntArray()
			},
			nullBitmap: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.Append(1)
				b.AppendNull()
				b.Append(3)
				b.AppendNull()
				b.AppendNull()
				b.Append(6)
				return b.NewIntArray()
			},
			want: []interface{}{nil, int64(3), nil},
		},
		{
			name: "all nulls in nullBitmap",
			input: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.AppendNull()
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(5)
				b.AppendNull()
				return b.NewIntArray()
			},
			nullBitmap: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.AppendNull()
				b.AppendNull()
				b.AppendNull()
				b.AppendNull()
				b.AppendNull()
				b.AppendNull()
				return b.NewIntArray()
			},
			want: []interface{}{},
		},
		{
			name: "no nulls",
			input: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				b.Append(1)
				b.Append(2)
				b.Append(3)
				b.Append(4)
				return b.NewIntArray()
			},
			want: []interface{}{int64(1), int64(2), int64(3), int64(4)},
		},
		{
			name: "empty",
			input: func(mem memory.Allocator) *array.Int {
				b := array.NewIntBuilder(mem)
				return b.NewIntArray()
			},
			want: []interface{}{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			input := tc.input(mem)
			nullBitmap := input
			if tc.nullBitmap != nil {
				nullBitmap = tc.nullBitmap(mem)
				defer nullBitmap.Release()
			}

			got := array.CopyValidValues(mem, input, nullBitmap).(*array.Int)

			want := tc.want

			if diff := cmp.Diff(want, arrayToInterfaceSlice(got)); diff != "" {
				t.Errorf("unexpected value -want/+got:\n%v", diff)
			}

			input.Release()
			got.Release()
		})
	}
}

func arrayToInterfaceSlice(array *array.Int) []interface{} {
	out := []interface{}{}
	for i := 0; i < array.Len(); i++ {
		if array.IsValid(i) {
			out = append(out, array.Value(i))
		} else {
			out = append(out, nil)
		}
	}
	return out
}

func getValue(arr array.Array, i int) interface{} {
	if arr.IsNull(i) {
		return nil
	}

	switch arr := arr.(type) {
	case *array.Int:
		return arr.Value(i)
	case *array.Uint:
		return arr.Value(i)
	case *array.Float:
		return arr.Value(i)
	case *array.String:
		return arr.Value(i)
	case *array.Boolean:
		return arr.Value(i)
	default:
		panic("unimplemented")
	}
}

func countNulls(arr []interface{}) (n int) {
	for _, v := range arr {
		if v == nil {
			n++
		}
	}
	return n
}
