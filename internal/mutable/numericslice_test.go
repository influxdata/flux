package mutable_test

import (
	"testing"

	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/memory"
)

func assertSize(t *testing.T, a *memory.Allocator, n int) {
	if got, want := a.Allocated(), int64(n); want != got {
		t.Errorf("allocator got unexpected bytes allocated -want/+got:\t\n- %d\n\t+ %d", want, got)
	}
}

func TestInt64TrackedSlice_Append(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_AppendValues(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_Reserve(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_Resize(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_Set(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_NewArray(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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

func TestInt64TrackedSlice_MixReserveResize(t *testing.T) {
	mem := new(memory.Allocator)
	defer assertSize(t, mem, 0)

	b := mutable.NewInt64TrackedSlice(mem, 0, 0)
	defer b.Release()
	assertSize(t, mem, 0)

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
