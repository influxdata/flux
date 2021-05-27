package mutable_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/internal/mutable"
)

func TestBinaryArray_Append(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Appending will change the length to 1.
	b.AppendString("a")
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Appending again changes the length to 2.
	b.AppendString("b")
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewBinaryArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.ValueString(0), "a"; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	if got, want := a.ValueString(1), "b"; got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %q\n\t+ %q", want, got)
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

func TestBinaryArray_AppendValues(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	b.AppendStringValues([]string{"a", "b"})
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2.
	a := b.NewBinaryArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.ValueString(0), "a"; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	if got, want := a.ValueString(1), "b"; got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %q\n\t+ %q", want, got)
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

func TestBinaryArray_Reserve(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
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
	b.AppendString("a")
	if got := b.Cap(); got != want {
		t.Fatalf("unexpected capacity -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestBinaryArray_Resize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Resize to 2 for 2 elements.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with default values.
	a := b.NewBinaryArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.ValueString(0), ""; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	if got, want := a.ValueString(1), ""; got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	a.Release()
}

func TestBinaryArray_Set(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append two values.
	// Resize the array to two values and set each of the values.
	b.Resize(2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Set the values.
	b.SetString(0, "a")
	b.SetString(1, "b")

	// Verify the values using Value.
	if got, want := b.ValueString(0), "a"; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	if got, want := b.ValueString(1), "b"; got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}

	// Constructing the array creates an arrow array of length 2 with the same values.
	a := b.NewBinaryArray()
	if got, want := a.Len(), 2; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.ValueString(0), "a"; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	if got, want := a.ValueString(1), "b"; got != want {
		t.Fatalf("unexpected value at index 1 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	a.Release()
}

func TestBinaryArray_NewArray(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
	defer b.Release()
	mem.AssertSize(t, 0)

	// Append a value.
	b.AppendString("a")
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Construct an array. This should reset the builder.
	b.NewArray().Release()

	// Append a value to the builder again.
	b.AppendString("b")
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// Constructing the array creates an arrow array of length 1 with the same value.
	a := b.NewBinaryArray()
	if got, want := a.Len(), 1; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	if got, want := a.ValueString(0), "b"; got != want {
		t.Fatalf("unexpected value at index 0 -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
	a.Release()
}

func TestBinaryArray_MixReserveResize(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	defer mem.AssertSize(t, 0)

	b := mutable.NewBinaryArray(mem, arrow.BinaryTypes.String)
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
		ch := byte('a') + byte(i)
		if i < s1 {
			b.SetString(i, string([]byte{ch}))
		} else {
			b.AppendString(string([]byte{ch}))
		}
	}
	if got, want := b.Len(), total; got != want {
		t.Fatalf("unexpected length -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
	for i := 0; i < total; i++ {
		ch := byte('a') + byte(i)
		if got, want := b.ValueString(i), string([]byte{ch}); got != want {
			t.Fatalf("unexpected value at index %d: %v != %v", i, want, got)
		}
	}
}
