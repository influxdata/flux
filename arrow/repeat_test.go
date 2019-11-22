package arrow_test

import (
	"testing"

	stdarrow "github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func TestRepeat(t *testing.T) {
	for _, tt := range []struct {
		name     string
		v        interface{}
		dataType stdarrow.DataType
		sz       int
		check    func(t *testing.T, arr array.Interface, mem *memory.Allocator)
	}{
		{
			name:     "Int",
			v:        int64(5),
			dataType: stdarrow.PrimitiveTypes.Int64,
			sz:       stdarrow.Int64Traits.BytesRequired(100),
			check: func(t *testing.T, arr array.Interface, mem *memory.Allocator) {
				// Convert the array to an integer.
				a := arrow.AsInts(arr)
				for i := 0; i < 100; i++ {
					if got, want := a.Value(i), int64(5); got != want {
						t.Fatalf("unexpected value at index %d -want/+got:\n\t- %d\n\t+ %d", i, want, got)
					}
				}

				// Use AsInts again.
				b := arrow.AsInts(arr)
				for i := 0; i < 100; i++ {
					if got, want := a.Value(i), int64(5); got != want {
						t.Fatalf("unexpected value at index %d -want/+got:\n\t- %d\n\t+ %d", i, want, got)
					}
				}

				// Retain a reference and ensure the memory is still allocated.
				a.Retain()
				b.Retain()
				arr.Release()
				mem.AssertSizeAtLeast(t, 1)

				// Release both of these arrays.
				a.Release()
				b.Release()
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mem := &memory.Allocator{}
			defer mem.AssertSize(t, 0)

			// Create an array with repeat using the value.
			arr := arrow.Repeat(values.New(tt.v), 100, mem)

			// This should not actually trigger any allocations.
			mem.AssertSize(t, 0)

			// The length should be 10 and should not trigger allocations.
			if got, want := arr.Len(), 100; got != want {
				t.Errorf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
			}

			// The data type is the same.
			if got, want := arr.DataType(), tt.dataType; got != want {
				t.Errorf("unexpected data type -want/+got:\n\t- %v\n\t+ %v", want, got)
			}

			// IsValid should return true.
			if got := arr.IsValid(0); !got {
				t.Error("unexpected result for checking validity")
			}

			// The size should not be zero anymore.
			mem.AssertSizeAtLeast(t, 1)

			// Run custom checks on the actual array.
			tt.check(t, arr, mem)

			// Release the array.
			arr.Release()
		})
	}
}
