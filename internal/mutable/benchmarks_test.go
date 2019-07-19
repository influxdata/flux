package mutable_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/internal/mutable"
	fluxmemory "github.com/influxdata/flux/memory"
)

func createArraysWithSize(n, s int) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	for i := 0; i < n; i++ {
		b := mutable.NewInt64Array(mem)
		b.Resize(s)
		b.Release()
	}
}

func createSlicesWithSize(n, s int) {
	mem := &fluxmemory.Allocator{}
	for i := 0; i < n; i++ {
		b := mutable.NewInt64TrackedSlice(mem, s, s)
		b.Release()
	}
}

func BenchmarkInt64SizedArray(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArraysWithSize(100, 100)
	}
}

func BenchmarkInt64SizedSlice(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createSlicesWithSize(100, 100)
	}
}

func createArraysWithSizeAndCapacity(n, s int) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	for i := 0; i < n; i++ {
		b := mutable.NewInt64Array(mem)
		b.Resize(s)
		b.Reserve(s)
		b.Release()
	}
}

func createSlicesWithSizeAndCapacity(n, s int) {
	mem := &fluxmemory.Allocator{}
	for i := 0; i < n; i++ {
		b := mutable.NewInt64TrackedSlice(mem, s, s*2)
		b.Release()
	}
}

func BenchmarkInt64ArrayWCapacity(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArraysWithSizeAndCapacity(100, 100)
	}
}

func BenchmarkInt64SliceWCapacity(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createSlicesWithSizeAndCapacity(100, 100)
	}
}

func createArraySetReserveAppend(s int) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	b := mutable.NewInt64Array(mem)
	defer b.Release()
	b.Resize(s)
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	b.Reserve(s)
	for i := 0; i < s; i++ {
		b.Append(int64(i))
	}
}

func createSliceSetReserveAppend(s int) {
	mem := &fluxmemory.Allocator{}
	b := mutable.NewInt64TrackedSlice(mem, s, s)
	defer b.Release()
	b.Resize(s)
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	b.Reserve(s)
	for i := 0; i < s; i++ {
		b.Append(int64(i))
	}
}

func BenchmarkInt64ArrayAddValues(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArraySetReserveAppend(1000)
	}
}

func BenchmarkInt64SliceAddValues(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createSliceSetReserveAppend(1000)
	}
}

func createArrayKnownCapacity(s int) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	b := mutable.NewInt64Array(mem)
	defer b.Release()
	b.Resize(s)
	b.Reserve(s)
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	for i := 0; i < s; i++ {
		b.Append(int64(i))
	}
}

func createSliceKnownCapacity(s int) {
	mem := &fluxmemory.Allocator{}
	b := mutable.NewInt64TrackedSlice(mem, s, s*2)
	defer b.Release()
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	for i := 0; i < s; i++ {
		b.Append(int64(i))
	}
}

func BenchmarkInt64ArrayAddValuesKnownCapacity(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArrayKnownCapacity(1000)
	}
}

func BenchmarkInt64SliceAddValuesKnownCapacity(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createSliceKnownCapacity(1000)
	}
}

func createArrowArrayFromArray(s int) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	b := mutable.NewInt64Array(mem)
	defer b.Release()
	b.Resize(s)
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	b.NewInt64Array().Release()
}

func createArrowArrayFromSlice(s int) {
	mem := &fluxmemory.Allocator{}
	b := mutable.NewInt64TrackedSlice(mem, s, s)
	defer b.Release()
	for i := 0; i < s; i++ {
		b.Set(i, int64(i))
	}
	b.NewInt64Array().Release()
}

func BenchmarkInt64ArrayArrowArray(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArrowArrayFromArray(1000)
	}
}

func BenchmarkInt64SliceArrowArray(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createArrowArrayFromSlice(1000)
	}
}
