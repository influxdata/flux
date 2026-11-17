package array_test

import (
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	apachearray "github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/array"
	fluxmemory "github.com/influxdata/flux/memory"
	"github.com/stretchr/testify/assert"
)

var _ arrow.Array = (*array.String)(nil)

func TestString(t *testing.T) {
	for _, tc := range []struct {
		name       string
		build      func(b *array.StringBuilder)
		bsz        int
		sz         int
		want       []interface{}
		wantString string
	}{
		{
			name: "Constant",
			build: func(b *array.StringBuilder) {
				for i := 0; i < 10; i++ {
					b.Append("abcdefghij")
				}
			},
			bsz: 64 + // values null bitmap
				64 + // values offset array
				64, // values data array
			sz: 64 + // run ends null bitmap
				64 + // run-ends array
				64 + // values null bitmap
				64 + // values offset array
				64, // values data array
			want: []any{
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
			wantString: `["abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij" "abcdefghij"]`,
		},
		{
			name: "RLE",
			build: func(b *array.StringBuilder) {
				for range 5 {
					b.Append("a")
				}
				for range 5 {
					b.Append("b")
				}
			},
			bsz: 64 + // values null bitmap
				128 + // values offset array
				64, // values data array
			sz: 64 + // values null bitmap
				128 + // values offset array
				64, // values data array
			want: []any{
				"a", "a", "a", "a", "a",
				"b", "b", "b", "b", "b",
			},
			wantString: `["a" "a" "a" "a" "a" "b" "b" "b" "b" "b"]`,
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
			bsz: 64 + // values null bitmap
				128 + // values offset array
				64, // values data array
			sz: 64 + // values null bitmap
				128 + // values offset array
				64, // values data array
			want: []any{
				"a", "b", "c", "d", "e",
				nil, "g", "h", "i", "j",
			},
			wantString: `["a" "b" "c" "d" "e" (null) "g" "h" "i" "j"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			// Construct a string builder, resize it to a capacity larger than
			// what is required, then verify we see that value, or a higher
			// value where the builder type has a minimum capacity.
			b := array.NewStringBuilder(mem)
			b.Resize(len(tc.want) + 2)
			if want, got := len(tc.want)+2, b.Cap(); want > got {
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
			if want, got := len(tc.want)+2, b.Cap(); want > got {
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
			assert.Equal(t, tc.wantString, arr.String())
		})
	}
}

func TestNewStringData(t *testing.T) {
	alloc := fluxmemory.NewResourceAllocator(nil)
	// Need to use the Apache binary builder to be able to create an actual
	// Arrow Binary array.
	sb := apachearray.NewBinaryBuilder(alloc, array.StringType)
	vals := []string{"a", "b", "c"}
	for _, v := range vals {
		sb.AppendString(v)
	}
	a := sb.NewArray()
	s := array.NewStringData(a.Data())
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
		t.Errorf("expected allocated to be %v, was %v", want, got)
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
		assert.Equal(t, 64+64+64+64+64, mem.CurrentAlloc(), "unexpected memory allocation.")
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
		assert.Equal(t, 64+128+64, mem.CurrentAlloc(), "unexpected memory allocation.")
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

func TestRunEndEncodedString(t *testing.T) {
	t.Run("DirectCreation", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		// Create run-end encoded array directly using Arrow API
		vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		vb.AppendString("hello")
		vb.AppendString("world")
		vb.AppendNull()
		values := vb.NewArray()
		defer values.Release()

		reb := apachearray.NewInt32Builder(mem)
		reb.Append(3) // First run ends at index 3 (3 "hello"s)
		reb.Append(5) // Second run ends at index 5 (2 "world"s)
		reb.Append(8) // Third run ends at index 8 (3 nulls)
		runEnds := reb.NewArray()
		defer runEnds.Release()

		reeArr := apachearray.NewRunEndEncodedArray(runEnds, values, 8, 0)
		defer reeArr.Release()

		arr := array.NewStringData(reeArr.Data())
		defer arr.Release()

		// Verify the array properties
		assert.Equal(t, 8, arr.Len())
		assert.Equal(t, 3, arr.NullN())

		// Check values
		expected := []interface{}{
			"hello", "hello", "hello",
			"world", "world",
			nil, nil, nil,
		}
		for i := 0; i < arr.Len(); i++ {
			if expected[i] == nil {
				assert.True(t, arr.IsNull(i), "Expected null at index %d", i)
				assert.False(t, arr.IsValid(i), "Expected invalid at index %d", i)
				assert.Equal(t, "", arr.Value(i), "Expected empty string for null at index %d", i)
			} else {
				assert.False(t, arr.IsNull(i), "Expected non-null at index %d", i)
				assert.True(t, arr.IsValid(i), "Expected valid at index %d", i)
				assert.Equal(t, expected[i], arr.Value(i), "Unexpected value at index %d", i)
			}
		}
	})

	t.Run("WithOffset", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		// Create values
		vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		vb.AppendString("a")
		vb.AppendString("b")
		vb.AppendString("c")
		values := vb.NewArray()
		defer values.Release()

		// Create run ends for logical array [a,a,a,b,b,c,c,c,c,c]
		reb := apachearray.NewInt32Builder(mem)
		reb.Append(3)  // "a" runs to index 3
		reb.Append(5)  // "b" runs to index 5
		reb.Append(10) // "c" runs to index 10
		runEnds := reb.NewArray()
		defer runEnds.Release()

		// Create full array
		fullArr := apachearray.NewRunEndEncodedArray(runEnds, values, 10, 0)
		defer fullArr.Release()

		// Create sliced array with offset 3, length 5
		// This should give us [b,b,c,c,c]
		slicedArr := apachearray.NewSlice(fullArr, 3, 8)
		defer slicedArr.Release()

		arr := array.NewStringData(slicedArr.Data())
		defer arr.Release()

		assert.Equal(t, 5, arr.Len())

		expected := []string{"b", "b", "c", "c", "c"}
		for i := 0; i < arr.Len(); i++ {
			assert.Equal(t, expected[i], arr.Value(i), "Unexpected value at index %d", i)
		}
	})

	t.Run("WithOffsetMidRun", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		// Create values
		vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		vb.AppendString("first")
		vb.AppendString("second")
		vb.AppendString("third")
		values := vb.NewArray()
		defer values.Release()

		// Create run ends for logical array:
		// [first,first,first,first,first,second,second,second,third,third]
		// Indices: 0,1,2,3,4,5,6,7,8,9
		reb := apachearray.NewInt32Builder(mem)
		reb.Append(5)  // "first" runs to index 5 (5 elements)
		reb.Append(8)  // "second" runs to index 8 (3 elements)
		reb.Append(10) // "third" runs to index 10 (2 elements)
		runEnds := reb.NewArray()
		defer runEnds.Release()

		// Create full array
		fullArr := apachearray.NewRunEndEncodedArray(runEnds, values, 10, 0)
		defer fullArr.Release()

		// Create sliced array with offset 2, length 6
		// Starting partway through the first run
		// This should give us [first,first,first,second,second,second]
		slicedArr := apachearray.NewSlice(fullArr, 2, 8)
		defer slicedArr.Release()

		arr := array.NewStringData(slicedArr.Data())
		defer arr.Release()

		assert.Equal(t, 6, arr.Len())

		expected := []string{"first", "first", "first", "second", "second", "second"}
		for i := 0; i < arr.Len(); i++ {
			assert.Equal(t, expected[i], arr.Value(i), "Unexpected value at index %d", i)
		}

		// Also test individual value access to ensure offset calculation is correct
		assert.Equal(t, "first", arr.Value(0))
		assert.Equal(t, "first", arr.Value(1))
		assert.Equal(t, "first", arr.Value(2))
		assert.Equal(t, "second", arr.Value(3))
		assert.Equal(t, "second", arr.Value(4))
		assert.Equal(t, "second", arr.Value(5))
	})

	t.Run("StringMethod", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		vb.AppendString("foo")
		vb.AppendNull()
		vb.AppendString("bar")
		values := vb.NewArray()
		defer values.Release()

		reb := apachearray.NewInt32Builder(mem)
		reb.Append(2) // "foo" x2
		reb.Append(3) // null x1
		reb.Append(5) // "bar" x2
		runEnds := reb.NewArray()
		defer runEnds.Release()

		reeArr := apachearray.NewRunEndEncodedArray(runEnds, values, 5, 0)
		defer reeArr.Release()

		arr := array.NewStringData(reeArr.Data())
		defer arr.Release()

		expected := `["foo" "foo" (null) "bar" "bar"]`
		assert.Equal(t, expected, arr.String())
	})

	t.Run("ValueLen", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
		vb.AppendString("short")
		vb.AppendString("longer string")
		vb.AppendNull()
		values := vb.NewArray()
		defer values.Release()

		reb := apachearray.NewInt32Builder(mem)
		reb.Append(2) // "short" x2
		reb.Append(4) // "longer string" x2
		reb.Append(5) // null x1
		runEnds := reb.NewArray()
		defer runEnds.Release()

		reeArr := apachearray.NewRunEndEncodedArray(runEnds, values, 5, 0)
		defer reeArr.Release()

		arr := array.NewStringData(reeArr.Data())
		defer arr.Release()

		expectedLengths := []int{5, 5, 13, 13, 0}
		for i := 0; i < arr.Len(); i++ {
			assert.Equal(t, expectedLengths[i], arr.ValueLen(i), "Unexpected length at index %d", i)
		}
	})
}

func TestStringBuilderRunEndEncoding(t *testing.T) {
	t.Run("SingleValueRun", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		b := array.NewStringBuilder(mem)
		defer b.Release()

		// Append the same value multiple times
		for i := 0; i < 5; i++ {
			b.Append("repeated")
		}

		arr := b.NewStringArray()
		defer arr.Release()

		// Verify it's using run-end encoding
		assert.Equal(t, arrow.RUN_END_ENCODED, arr.DataType().ID())
		assert.Equal(t, 5, arr.Len())

		for i := 0; i < 5; i++ {
			assert.Equal(t, "repeated", arr.Value(i))
		}
	})

	t.Run("TransitionToHydrated", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		b := array.NewStringBuilder(mem)
		defer b.Release()

		// Start with repeated values
		for i := 0; i < 3; i++ {
			b.Append("first")
		}

		// Add a different value (should trigger hydration)
		b.Append("second")

		// Add more values
		b.Append("third")

		arr := b.NewStringArray()
		defer arr.Release()

		assert.Equal(t, 5, arr.Len())
		expected := []string{"first", "first", "first", "second", "third"}
		for i := 0; i < 5; i++ {
			assert.Equal(t, expected[i], arr.Value(i))
		}
	})

	t.Run("NullHandling", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		b := array.NewStringBuilder(mem)
		defer b.Release()

		// Start with repeated values
		for i := 0; i < 3; i++ {
			b.Append("value")
		}

		// This should trigger hydration
		b.AppendNull()
		b.Append("value")
		b.AppendNull()

		arr := b.NewStringArray()
		defer arr.Release()

		assert.Equal(t, 6, arr.Len())
		assert.Equal(t, 2, arr.NullN())

		expected := []interface{}{"value", "value", "value", nil, "value", nil}
		for i := 0; i < 6; i++ {
			if expected[i] == nil {
				assert.True(t, arr.IsNull(i))
			} else {
				assert.False(t, arr.IsNull(i))
				assert.Equal(t, expected[i], arr.Value(i))
			}
		}
	})

	t.Run("AppendBytes", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		b := array.NewStringBuilder(mem)
		defer b.Release()

		// Use AppendBytes instead of Append
		testBytes := []byte("test")
		for i := 0; i < 4; i++ {
			b.AppendBytes(testBytes)
		}

		arr := b.NewStringArray()
		defer arr.Release()

		assert.Equal(t, 4, arr.Len())
		for i := 0; i < 4; i++ {
			assert.Equal(t, "test", arr.Value(i))
		}
	})

	t.Run("MultipleBuilds", func(t *testing.T) {
		mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)

		b := array.NewStringBuilder(mem)
		defer b.Release()

		// First build with run-end encoding
		for i := 0; i < 3; i++ {
			b.Append("a")
		}
		arr1 := b.NewStringArray()
		defer arr1.Release()
		assert.Equal(t, arrow.RUN_END_ENCODED, arr1.DataType().ID())

		// Second build with different pattern
		b.Append("b")
		b.Append("c")
		arr2 := b.NewStringArray()
		defer arr2.Release()

		// Verify second array
		assert.Equal(t, 2, arr2.Len())
		assert.Equal(t, "b", arr2.Value(0))
		assert.Equal(t, "c", arr2.Value(1))
	})
}

func TestStringRepeatRunEndEncoding(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	arr := array.StringRepeat("repeated", 100, mem)
	defer arr.Release()

	// Verify it's using run-end encoding
	assert.Equal(t, arrow.RUN_END_ENCODED, arr.DataType().ID())
	assert.Equal(t, 100, arr.Len())

	// Verify all values are the same
	for i := 0; i < 100; i++ {
		assert.Equal(t, "repeated", arr.Value(i))
		assert.False(t, arr.IsNull(i))
	}

	// Test String() method doesn't crash with large arrays
	str := arr.String()
	assert.Contains(t, str, "repeated")
}

func TestRunEndEncodedSlice(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Create a run-end encoded array using StringRepeat
	original := array.StringRepeat("value", 10, mem)
	defer original.Release()

	// Slice the array
	sliced := array.Slice(original, 3, 8).(*array.String)
	defer sliced.Release()

	assert.Equal(t, 5, sliced.Len())
	for i := 0; i < 5; i++ {
		assert.Equal(t, "value", sliced.Value(i))
	}
}

func TestMakeFromDataRunEndEncoded(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Create run-end encoded data
	vb := apachearray.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
	vb.AppendString("test")
	values := vb.NewArray()
	defer values.Release()

	reb := apachearray.NewInt32Builder(mem)
	reb.Append(5)
	runEnds := reb.NewArray()
	defer runEnds.Release()

	reeArr := apachearray.NewRunEndEncodedArray(runEnds, values, 5, 0)
	defer reeArr.Release()

	// Use MakeFromData
	arr := array.MakeFromData(reeArr.Data())
	defer arr.Release()

	strArr, ok := arr.(*array.String)
	assert.True(t, ok, "Expected String array from MakeFromData")
	assert.Equal(t, 5, strArr.Len())
	for i := 0; i < 5; i++ {
		assert.Equal(t, "test", strArr.Value(i))
	}
}
