package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/dependenciestest"
	"github.com/mvn-trinhnguyen2-dn/flux/dependency"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/types"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

type isTypeCase struct {
	name     string
	value    values.Value
	type_    string
	expected bool
}

func TestIsType(t *testing.T) {
	cases := []isTypeCase{
		{
			name:     "ints are ints",
			value:    values.NewInt(1),
			type_:    "int",
			expected: true,
		},
		{
			name:     "ints are not strings",
			value:    values.NewInt(1),
			type_:    "string",
			expected: false,
		},
		{
			name:     "strings are strings",
			value:    values.NewString(""),
			type_:    "string",
			expected: true,
		},
		{
			name:     "int arrays are not ints",
			value:    values.NewArray(semantic.NewArrayType(semantic.BasicInt)),
			type_:    "int",
			expected: false,
		},
	}

	for _, tc := range cases {
		isTypeTestHelper(t, tc)
	}
}

func isTypeTestHelper(t *testing.T, tc isTypeCase) {
	t.Run(tc.name, func(t *testing.T) {
		isTypeFn := types.IsType()

		fluxArg := values.NewObjectWithValues(map[string]values.Value{
			"v":    tc.value,
			"type": values.NewString(tc.type_),
		})

		ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
		defer deps.Finish()

		got, err := isTypeFn.Call(ctx, fluxArg)
		if err != nil {
			t.Error(err.Error())
			return
		}

		want := tc.expected
		if got.Bool() != want {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want, got))
		}
	})
}
