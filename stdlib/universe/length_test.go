package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

type lengthCase struct {
	name     string
	arr      []values.Value
	typ      semantic.MonoType
	expected int
}

func TestLength_NewQuery(t *testing.T) {

	cases := []lengthCase{
		{
			name:     "empty arr",
			arr:      []values.Value{},
			typ:      semantic.BasicInt,
			expected: 0,
		},
		{
			name:     "nonempty arr",
			arr:      []values.Value{values.NewInt(3), values.NewInt(2), values.NewInt(1)},
			typ:      semantic.BasicInt,
			expected: 3,
		},
		{
			name:     "string arr",
			arr:      []values.Value{values.NewString("abcd")},
			typ:      semantic.BasicString,
			expected: 1,
		},
		{
			name:     "chinese string arr",
			arr:      []values.Value{values.NewString("汉"), values.NewString("汉")},
			typ:      semantic.BasicString,
			expected: 2,
		},
		{
			name: "bool arr",
			arr: []values.Value{values.NewBool(true), values.NewBool(false),
				values.NewBool(true), values.NewBool(false), values.NewBool(true), values.NewBool(false)},
			typ:      semantic.BasicBool,
			expected: 6,
		},
		{
			name:     "float arr",
			arr:      []values.Value{values.NewFloat(12.423), values.NewFloat(-0.294)},
			typ:      semantic.BasicFloat,
			expected: 2,
		},
	}

	for _, tc := range cases {
		lengthTestHelper(t, tc)
	}
}

func lengthTestHelper(t *testing.T, tc lengthCase) {
	t.Skip("https://github.com/influxdata/flux/issues/2481")
	t.Helper()
	length := universe.MakeLengthFunc()
	result, err := length.Call(
		dependenciestest.Default().Inject(context.Background()),
		values.NewObjectWithValues(map[string]values.Value{
			"arr": values.NewArrayWithBacking(tc.typ, tc.arr),
		}),
	)

	if err != nil {
		t.Error(err.Error())
	} else if result.Int() != int64(tc.expected) {
		t.Error("expected %i, got %i", result.Int(), int64(tc.expected))
	}
}
