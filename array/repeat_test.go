package array_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/array"
)

func TestRepeat(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    interface{}
		sz   int
	}{
		{
			name: "Int",
			v:    int64(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "Uint",
			v:    uint64(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "Float",
			v:    float64(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "String",
			v:    "a",
			sz:   0, // optimized away
		},
		{
			name: "Boolean",
			v:    true,
			sz:   128, // 64 bytes (bools), 64 bytes (nulls)
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			arr := array.Repeat(tc.v, 10, mem)
			mem.AssertSize(t, tc.sz)
			arr.Release()
		})
	}
}
