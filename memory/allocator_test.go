package memory_test

import (
	"testing"

	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/memory"
)

func TestAllocator_Allocate(t *testing.T) {
	mem := arrowmemory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	allocator := &memory.Allocator{Allocator: mem}
	b := allocator.Allocate(64)

	mem.AssertSize(t, 64)
	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	allocator.Free(b)

	mem.AssertSize(t, 0)
	if want, got := int64(0), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

func TestAllocator_Reallocate(t *testing.T) {
	mem := arrowmemory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	allocator := &memory.Allocator{Allocator: mem}
	b := allocator.Allocate(64)

	mem.AssertSize(t, 64)
	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	b = allocator.Reallocate(128, b)

	mem.AssertSize(t, 128)
	if want, got := int64(128), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(128), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	allocator.Free(b)

	mem.AssertSize(t, 0)
	if want, got := int64(0), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(128), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

func TestAllocator_MaxAfterFree(t *testing.T) {
	allocator := &memory.Allocator{}
	if err := allocator.Account(64); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Free should restore the memory to zero, but have max be the same.
	_ = allocator.Account(-64)

	if want, got := int64(0), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Allocate a smaller amount of memory and the max should still be 64.
	if err := allocator.Account(32); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(32), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

func TestAllocator_Limit(t *testing.T) {
	maxLimit := int64(64)
	allocator := &memory.Allocator{Limit: &maxLimit}
	if err := allocator.Account(64); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Attempts to allocate more should result in an error.
	if err := allocator.Account(1); err == nil {
		t.Fatal("expected error")
	}

	// The Allocate method should panic.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Fatal("expected panic")
			}
		}()

		b := allocator.Allocate(64)
		allocator.Free(b)
	}()

	// The counts should not change.
	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Free should restore the memory so we can allocate more.
	_ = allocator.Account(-64)

	if want, got := int64(0), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// This allocation should succeed.
	if err := allocator.Account(32); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(32), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// This allocation should fail.
	if err := allocator.Account(64); err == nil {
		t.Fatal("expected error")
	}

	if want, got := int64(32), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

func TestAllocator_Free(t *testing.T) {
	allocator := &memory.Allocator{}
	if err := allocator.Account(64); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(64), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Free the memory.
	_ = allocator.Account(-64)

	if want, got := int64(0), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), allocator.MaxAllocated(); want != got {
		t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}
