package mutable_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/internal/mutable"
)

func TestBooleanArray_Append(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Appending will change the length to 1.
	b.Append(true)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Appending again changes the length to 2.
	b.Append(false)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewBooleanArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := a.Value(0); !got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", true, got)
	}
	if got := a.Value(1); got {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", false, got)
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

func TestBooleanArray_AppendValues(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	b.AppendValues([]bool{true, false})
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewBooleanArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := a.Value(0); !got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", true, got)
	}
	if got := a.Value(1); got {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", false, got)
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

func TestBooleanArray_Reserve(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
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
	b.Append(true)
	if got := b.Cap(); got != want {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestBooleanArray_Resize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Resize to 2 for 2 elements.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with default values.
	a := b.NewBooleanArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := a.Value(0); got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	if got := a.Value(1); got {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestBooleanArray_Set(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	// Resize the array to two values and set each of the values.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Set the values.
	b.Set(0, true)
	b.Set(1, false)

	// Verify the values using Value.
	if got := b.Value(0); !got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", true, got)
	}
	if got := b.Value(1); got {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", false, got)
	}

	// Constructing the array creates an arrow array of length 2 with the same values.
	a := b.NewBooleanArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := a.Value(0); !got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", true, got)
	}
	if got := a.Value(1); got {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %v\n\t+ %v", false, got)
	}
	a.Release()
}

func TestBooleanArray_NewArray(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append a value.
	b.Append(true)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Construct an array. This should reset the builder.
	b.NewArray().Release()

	// Append a value to the builder again.
	b.Append(true)
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 1 with the same value.
	a := b.NewBooleanArray()
	if got, want := a.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got := a.Value(0); !got {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %v\n\t+ %v", true, got)
	}
	a.Release()
}

func TestBooleanArray_MixReserveResize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBooleanArray(mem)
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
		v := i%2 == 0
		if i < s1 {
			b.Set(i, v)
		} else {
			b.Append(v)
		}
	}
	if got, want := b.Len(), total; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	for i := 0; i < total; i++ {
		if got, want := b.Value(i), i%2 == 0; got != want {
			t.Fatalf("unexpected value at index %d: %v != %v", i, want, got)
		}
	}
}
