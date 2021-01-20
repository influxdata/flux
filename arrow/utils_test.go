package arrow_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
)

// TestEmpty asserts that the assumption in the Empty call is correct.
// Calling NewBuilder and then calling NewArray will not use any memory
// and will not leak memory if not released.
func TestEmpty(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	b := arrow.NewBuilder(flux.TInt, mem)
	_ = b.NewArray()
}
