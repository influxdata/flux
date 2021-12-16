package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

type containsCase struct {
	name     string
	value    values.Value
	set      []values.Value
	expected bool
}

func TestContains_NewQuery(t *testing.T) {

	cases := []containsCase{
		{
			name:     "empty set",
			value:    values.NewInt(1),
			set:      []values.Value{},
			expected: false,
		},
		{
			name:     "integer found",
			value:    values.NewInt(1),
			set:      []values.Value{values.NewInt(3), values.NewInt(2), values.NewInt(1)},
			expected: true,
		},
		{
			name:     "integer not found",
			value:    values.NewInt(1),
			set:      []values.Value{values.NewInt(11), values.NewInt(2), values.NewInt(3)},
			expected: false,
		},
		{
			name:     "unsigned integer found",
			value:    values.NewUInt(1),
			set:      []values.Value{values.NewUInt(3), values.NewUInt(2), values.NewUInt(1)},
			expected: true,
		},
		{
			name:     "unsigned integer not found",
			value:    values.NewUInt(1),
			set:      []values.Value{values.NewUInt(11), values.NewUInt(2), values.NewUInt(3)},
			expected: false,
		},
		{
			name:     "float found",
			value:    values.NewFloat(1.0),
			set:      []values.Value{values.NewFloat(3.0), values.NewFloat(2.0), values.NewFloat(1.0)},
			expected: true,
		},
		{
			name:     "float not found",
			value:    values.NewFloat(1.0),
			set:      []values.Value{values.NewFloat(11.0), values.NewFloat(2.0), values.NewFloat(3.0)},
			expected: false,
		},
		{
			name:     "string found",
			value:    values.NewString("1.0"),
			set:      []values.Value{values.NewString("3.0"), values.NewString("2.0"), values.NewString("1.0")},
			expected: true,
		},
		{
			name:     "string not found",
			value:    values.NewString("1.0"),
			set:      []values.Value{values.NewString("11.0"), values.NewString("2.0"), values.NewString("3.0")},
			expected: false,
		},
		{
			name:     "bool found",
			value:    values.NewBool(true),
			set:      []values.Value{values.NewBool(true), values.NewBool(false), values.NewBool(true)},
			expected: true,
		},
		{
			name:     "bool not found",
			value:    values.NewBool(false),
			set:      []values.Value{values.NewBool(true), values.NewBool(true)},
			expected: false,
		},
		{
			name:     "time found",
			value:    values.NewTime(1),
			set:      []values.Value{values.NewTime(3), values.NewTime(2), values.NewTime(1)},
			expected: true,
		},
		{
			name:     "time not found",
			value:    values.NewTime(1),
			set:      []values.Value{values.NewTime(11), values.NewTime(2), values.NewTime(3)},
			expected: false,
		},
	}

	for _, tc := range cases {
		containsTestHelper(t, tc)
	}
}

func containsTestHelper(t *testing.T, tc containsCase) {
	t.Helper()
	contains := universe.MakeContainsFunc()
	result, err := contains.Call(dependenciestest.Default().Inject(context.Background()),
		values.NewObjectWithValues(map[string]values.Value{
			"value": tc.value,
			"set":   values.NewArrayWithBacking(semantic.NewArrayType(tc.value.Type()), tc.set),
		}),
	)

	if err != nil {
		t.Error(err.Error())
	} else if result.Bool() != tc.expected {
		t.Error("expected true, got false")
	}
}

func TestContains_Empty(t *testing.T) {
	script := `
		ok = not contains( value: "nothing", set: [] )
	`
	s := evalOrFail(t, script)

	if !mustLookup(s, "ok").Bool() {
		t.Errorf("ok was not OK indeed")
	}
}

func TestContains_ReceiveTableObjectIsError(t *testing.T) {
	src := `
	import "array"
	contains(value: 1, set: array.from(rows: [{_value: 1}]) |> map(fn: (r) => r._value ))`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error calling function \"contains\" @3:2-3:87: got a table stream; expected an array", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
