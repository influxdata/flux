package mutable_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/internal/mutable"
)

func TestInt64Array_Append(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Appending will change the length to 1.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Appending again changes the length to 2.
	b.Append(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewInt64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), int64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), int64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestInt64Array_AppendValues(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	b.AppendValues([]int64{1, 2})
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewInt64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), int64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), int64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestInt64Array_Reserve(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Reserve will increase the capacity, but not the length.
	// We do not verify the exact capacity because that doesn't matter,
	// just that it is greater than or equal to 2.
	b.Reserve(2)
	if got, want := b.Len(), 0; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := b.Cap(); got < 2 {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %s\n\t+ %d", "at least 2", got)
	}

	// If we append a value, the capacity should not change.
	want := b.Cap()
	b.Append(1)
	if got := b.Cap(); got != want {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestInt64Array_Resize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Resize to 2 for 2 elements.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with default values.
	a := b.NewInt64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), int64(0); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), int64(0); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestInt64Array_Set(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	// Resize the array to two values and set each of the values.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Set the values.
	b.Set(0, 1)
	b.Set(1, 2)

	// Verify the values using Value.
	if got, want := b.Value(0), int64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := b.Value(1), int64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with the same values.
	a := b.NewInt64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), int64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), int64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestInt64Array_NewArray(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append a value.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Construct an array. This should reset the builder.
	b.NewArray().Release()

	// Append a value to the builder again.
	b.Append(2)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 1 with the same value.
	a := b.NewInt64Array()
	if got, want := a.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), int64(2); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestInt64Array_MixReserveResize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewInt64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	s1 := 10
	s2 := 15
	total := s1 + s2

	// Mix resize and reserve.
	b.Resize(s1)
	b.Reserve(s2)
	if got, want := b.Len(), s1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := b.Cap(), total; got < want {
		t.Fatalf("unexpected capacity got < want: %d < %d", got, want)
	}

	for i := 0; i < total; i++ {
		if i < s1 {
			b.Set(i, int64(i))
		} else {
			b.Append(int64(i))
		}
	}
	if got, want := b.Len(), total; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	for i := 0; i < total; i++ {
		if got, want := b.Value(i), int64(i); got != want {
			t.Fatalf("unexpected value at index %d: %d != %d", i, want, got)
		}
	}
}

func TestUint64Array_Append(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Appending will change the length to 1.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Appending again changes the length to 2.
	b.Append(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewUint64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), uint64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), uint64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestUint64Array_AppendValues(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	b.AppendValues([]uint64{1, 2})
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewUint64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), uint64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), uint64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestUint64Array_Reserve(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Reserve will increase the capacity, but not the length.
	// We do not verify the exact capacity because that doesn't matter,
	// just that it is greater than or equal to 2.
	b.Reserve(2)
	if got, want := b.Len(), 0; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := b.Cap(); got < 2 {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %s\n\t+ %d", "at least 2", got)
	}

	// If we append a value, the capacity should not change.
	want := b.Cap()
	b.Append(1)
	if got := b.Cap(); got != want {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestUint64Array_Resize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Resize to 2 for 2 elements.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with default values.
	a := b.NewUint64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), uint64(0); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), uint64(0); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestUint64Array_Set(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	// Resize the array to two values and set each of the values.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Set the values.
	b.Set(0, 1)
	b.Set(1, 2)

	// Verify the values using Value.
	if got, want := b.Value(0), uint64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := b.Value(1), uint64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with the same values.
	a := b.NewUint64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), uint64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(1), uint64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestUint64Array_NewArray(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append a value.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Construct an array. This should reset the builder.
	b.NewArray().Release()

	// Append a value to the builder again.
	b.Append(2)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 1 with the same value.
	a := b.NewUint64Array()
	if got, want := a.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), uint64(2); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	a.Release()
}

func TestUint64Array_MixReserveResize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewUint64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	s1 := 10
	s2 := 15
	total := s1 + s2

	// Mix resize and reserve.
	b.Resize(s1)
	b.Reserve(s2)
	if got, want := b.Len(), s1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := b.Cap(), total; got < want {
		t.Fatalf("unexpected capacity got < want: %d < %d", got, want)
	}

	for i := 0; i < total; i++ {
		if i < s1 {
			b.Set(i, uint64(i))
		} else {
			b.Append(uint64(i))
		}
	}
	if got, want := b.Len(), total; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	for i := 0; i < total; i++ {
		if got, want := b.Value(i), uint64(i); got != want {
			t.Fatalf("unexpected value at index %d: %d != %d", i, want, got)
		}
	}
}

func TestFloat64Array_Append(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Appending will change the length to 1.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Appending again changes the length to 2.
	b.Append(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewFloat64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), float64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if got, want := a.Value(1), float64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestFloat64Array_AppendValues(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	b.AppendValues([]float64{1, 2})
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewFloat64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), float64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if got, want := a.Value(1), float64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}

	// IsNull should not fail.
	if got := a.IsNull(0); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.IsNull(1); got {
		t.Fatalf("unexpected null check -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestFloat64Array_Reserve(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Reserve will increase the capacity, but not the length.
	// We do not verify the exact capacity because that doesn't matter,
	// just that it is greater than or equal to 2.
	b.Reserve(2)
	if got, want := b.Len(), 0; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := b.Cap(); got < 2 {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %s\n\t+ %d", "at least 2", got)
	}

	// If we append a value, the capacity should not change.
	want := b.Cap()
	b.Append(1)
	if got := b.Cap(); got != want {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestFloat64Array_Resize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Resize to 2 for 2 elements.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with default values.
	a := b.NewFloat64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), float64(0); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if got, want := a.Value(1), float64(0); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	a.Release()
}

func TestFloat64Array_Set(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	// Resize the array to two values and set each of the values.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Set the values.
	b.Set(0, 1)
	b.Set(1, 2)

	// Verify the values using Value.
	if got, want := b.Value(0), float64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if got, want := b.Value(1), float64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with the same values.
	a := b.NewFloat64Array()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), float64(1); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	if got, want := a.Value(1), float64(2); got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	a.Release()
}

func TestFloat64Array_NewArray(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append a value.
	b.Append(1)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Construct an array. This should reset the builder.
	b.NewArray().Release()

	// Append a value to the builder again.
	b.Append(2)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 1 with the same value.
	a := b.NewFloat64Array()
	if got, want := a.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.Value(0), float64(2); got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", want, got)
	}
	a.Release()
}

func TestFloat64Array_MixReserveResize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewFloat64Array(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	s1 := 10
	s2 := 15
	total := s1 + s2

	// Mix resize and reserve.
	b.Resize(s1)
	b.Reserve(s2)
	if got, want := b.Len(), s1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := b.Cap(), total; got < want {
		t.Fatalf("unexpected capacity got < want: %d < %d", got, want)
	}

	for i := 0; i < total; i++ {
		if i < s1 {
			b.Set(i, float64(i))
		} else {
			b.Append(float64(i))
		}
	}
	if got, want := b.Len(), total; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	for i := 0; i < total; i++ {
		if got, want := b.Value(i), float64(i); got != want {
			t.Fatalf("unexpected value at index %d: %v != %v", i, want, got)
		}
	}
}
