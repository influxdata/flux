package array_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
)

func TestString(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(b *array.StringBuilder)
		sz    int
		want  []interface{}
	}{
		{
			name: "Constant",
			build: func(b *array.StringBuilder) {
				for i := 0; i < 10; i++ {
					b.Append("a")
				}
			},
			sz: 256,
			want: []interface{}{
				"a", "a", "a", "a", "a",
				"a", "a", "a", "a", "a",
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
			sz: 256,
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
			sz: 256,
			want: []interface{}{
				"a", "b", "c", "d", "e",
				nil, "g", "h", "i", "j",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			b := array.NewStringBuilder(mem)
			tc.build(b)

			arr := b.NewStringArray()
			defer arr.Release()
			mem.AssertSize(t, tc.sz)

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

func TestSlice(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(mem memory.Allocator) array.Interface
		i, j  int
		want  []interface{}
	}{
		{
			name: "Int",
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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
			build: func(mem memory.Allocator) array.Interface {
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

func getValue(arr array.Interface, i int) interface{} {
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
