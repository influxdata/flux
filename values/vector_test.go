package values

import (
	"testing"

	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/semantic"
)

func TestVectorTypes(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		wantType semantic.MonoType
	}{
		{
			name:     "int vector",
			input:    []interface{}{int64(1), int64(2), int64(3)},
			wantType: semantic.BasicInt,
		},
		{
			name:     "uint vector",
			input:    []interface{}{uint64(1), uint64(2), uint64(3)},
			wantType: semantic.BasicUint,
		},
		{
			name:     "float vector",
			input:    []interface{}{3.4, 5.6, 7.8},
			wantType: semantic.BasicFloat,
		},
		{
			name:     "string vector",
			input:    []interface{}{"one", "two", "three"},
			wantType: semantic.BasicString,
		},
		{
			name:     "bool vector",
			input:    []interface{}{true, false, true},
			wantType: semantic.BasicBool,
		},
	}
	for _, tc := range testCases {
		mem := memory.NewResourceAllocator(nil)
		got := NewVectorFromElements(mem, tc.input...)

		if !got.ElementType().Equal(tc.wantType) {
			t.Errorf("expected %v, got %v", tc.wantType, got.ElementType())
		}

		got.Release()

		if mem.Allocated() != 0 {
			t.Errorf("expected bytes allocated to be 0, got %d", mem.Allocated())
		}
	}
}
