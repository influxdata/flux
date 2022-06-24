package universe_test

import (
	"context"
	"testing"

	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/dependenciestest"
	"github.com/mvn-trinhnguyen2-dn/flux/dependency"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
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
	t.Helper()
	length := universe.MakeLengthFunc()
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	result, err := length.Call(
		ctx,
		values.NewObjectWithValues(map[string]values.Value{
			"arr": values.NewArrayWithBacking(semantic.NewArrayType(tc.typ), tc.arr),
		}),
	)

	if err != nil {
		t.Error(err.Error())
	} else if result.Int() != int64(tc.expected) {
		t.Error("expected %i, got %i", result.Int(), int64(tc.expected))
	}
}

func TestLength_ReceiveTableObjectIsError(t *testing.T) {
	src := `import "array"
			length(arr: array.from(rows: [{}]))`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error @2:16-2:38: expected [{}] (array) but found stream[{}] (argument arr)", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
