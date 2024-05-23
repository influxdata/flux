package array_test

import (
	"testing"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/values"
	"github.com/stretchr/testify/assert"
)

func TestRepeat(t *testing.T) {
	for _, tc := range []struct {
		name string
		t    flux.ColType
		v    values.Value
		sz   int
	}{
		{
			name: "Int",
			t:    flux.TInt,
			v:    values.NewInt(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "Uint",
			t:    flux.TUInt,
			v:    values.NewUInt(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "Float",
			t:    flux.TFloat,
			v:    values.NewFloat(4),
			sz:   192, // 128 bytes (ints), 64 bytes (nulls)
		},
		{
			name: "String",
			t:    flux.TString,
			v:    values.NewString("a"),
			sz:   64, // optimized to a single instance - 64 bytes
		},
		{
			name: "Boolean",
			t:    flux.TBool,
			v:    values.NewBool(true),
			sz:   128, // 64 bytes (bools), 64 bytes (nulls)
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
			defer mem.AssertSize(t, 0)

			arr := arrow.Repeat(tc.t, tc.v, 10, mem)
			assert.Equal(t, tc.sz, mem.CurrentAlloc(), "unexpected memory allocation.")
			arr.Release()
		})
	}
}
