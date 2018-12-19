package arrow_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/math"
	"github.com/apache/arrow/go/arrow/memory"
)

func TestSum_Float64_Empty(t *testing.T) {
	t.Skip("https://issues.apache.org/jira/browse/ARROW-4081")

	b := array.NewFloat64Builder(memory.NewGoAllocator())
	vs := b.NewFloat64Array()
	b.Release()

	defer func() {
		if err := recover(); err != nil {
			t.Errorf("unexpected panic: %s", err)
		}
	}()

	if got, want := math.Float64.Sum(vs), float64(0); got != want {
		t.Errorf("unexpected sum: %v != %v", got, want)
	}
}

func TestSum_Int64_Empty(t *testing.T) {
	t.Skip("https://issues.apache.org/jira/browse/ARROW-4081")

	b := array.NewInt64Builder(memory.NewGoAllocator())
	vs := b.NewInt64Array()
	b.Release()

	defer func() {
		if err := recover(); err != nil {
			t.Errorf("unexpected panic: %s", err)
		}
	}()

	if got, want := math.Int64.Sum(vs), int64(0); got != want {
		t.Errorf("unexpected sum: %v != %v", got, want)
	}
}

func TestSum_Uint64_Empty(t *testing.T) {
	t.Skip("https://issues.apache.org/jira/browse/ARROW-4081")

	b := array.NewUint64Builder(memory.NewGoAllocator())
	vs := b.NewUint64Array()
	b.Release()

	defer func() {
		if err := recover(); err != nil {
			t.Errorf("unexpected panic: %s", err)
		}
	}()

	if got, want := math.Uint64.Sum(vs), uint64(0); got != want {
		t.Errorf("unexpected sum: %v != %v", got, want)
	}
}
