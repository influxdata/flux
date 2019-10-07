package memory_test

import (
	"sync"
	"testing"

	arrowmemory "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
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

type MockMemoryManager struct {
	Left      int64
	RequestFn func(want int64) int64
}

func (m *MockMemoryManager) RequestMemory(want int64) (n int64, err error) {
	if m.Left < want {
		return 0, errors.New(codes.ResourceExhausted)
	}
	n = want
	if m.RequestFn != nil {
		n = m.RequestFn(want)
	}
	m.Left -= n
	return n, nil
}

func (m *MockMemoryManager) FreeMemory(bytes int64) {
	m.Left += bytes
}

func TestAllocator_RequestMemory(t *testing.T) {
	// Allow the memory manager to allocate 64 bytes.
	manager := &MockMemoryManager{
		Left: 64,
	}

	// Set the Limit to 64 and allocate 32 bytes of it.
	// This should not request more memory from the manager.
	allocator := &memory.Allocator{
		Limit:   func(v int64) *int64 { return &v }(64),
		Manager: manager,
	}
	if err := allocator.Account(32); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(32), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), *allocator.Limit; want != got {
		t.Fatalf("unexpected allocater limit -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(64), manager.Left; want != got {
		t.Fatalf("unexpected memory left in the manager -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Now request more than would be normally allowed and see that the manager
	// actually gives it more memory.
	if err := allocator.Account(64); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(96), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(96), *allocator.Limit; want != got {
		t.Fatalf("unexpected allocater limit -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(32), manager.Left; want != got {
		t.Fatalf("unexpected memory left in the manager -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Now request too much memory that the manager can't and won't give it.
	if err := allocator.Account(64); err == nil {
		t.Fatal("expected error")
	}

	// Change the allocator so it will give double the amount of requested memory
	// instead of the exact amount requested.
	manager.RequestFn = func(want int64) int64 {
		return want * 2
	}

	// Request 16 bytes of memory. The manager should give us 32 bytes.
	// We now test that the allocator increases its limit by the amount
	// the manager gives it rather than the request.
	if err := allocator.Account(16); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, got := int64(112), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(128), *allocator.Limit; want != got {
		t.Fatalf("unexpected allocater limit -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(0), manager.Left; want != got {
		t.Fatalf("unexpected memory left in the manager -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

// This test makes a lot of small allocations and has the memory manager
// only give small amounts of memory so that multiple goroutines are requesting
// memory from the manager concurrently. This is to ensure that requesting
// memory many times concurrently doesn't create a race condition.
func TestAllocator_RequestMemory_Concurrently(t *testing.T) {
	// Allow the memory manager to allocate 64 bytes.
	manager := &MockMemoryManager{
		Left: 128 * 128,
	}

	// Set the Limit to 64 and allocate 32 bytes of it.
	// This should not request more memory from the manager.
	allocator := &memory.Allocator{
		Limit:   func(v int64) *int64 { return &v }(0),
		Manager: manager,
	}
	var wg sync.WaitGroup
	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 128; i++ {
				if err := allocator.Account(1); err != nil {
					t.Errorf("unexpected error: %s", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	// Once we are finished, the allocator should have its limit set
	// to the total amount of the memory manager.
	if want, got := int64(128*128), allocator.Allocated(); want != got {
		t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(128*128), *allocator.Limit; want != got {
		t.Fatalf("unexpected allocater limit -want/+got\n\t- %d\n\t+ %d", want, got)
	}
	if want, got := int64(0), manager.Left; want != got {
		t.Fatalf("unexpected memory left in the manager -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}
