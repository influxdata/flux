package flux_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/stdlib" // Import stdlib
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	flux.FinalizeBuiltIns()
}

var ignoreUnexportedQuerySpec = cmpopts.IgnoreUnexported(flux.Spec{})

func TestSpec_JSON(t *testing.T) {
	srcData := []byte(`
{
	"operations":[
		{
			"id": "from",
			"kind": "from",
			"spec": {
				"bucket":"mybucket"
			}
		},
		{
			"id": "range",
			"kind": "range",
			"spec": {
				"start": "-4h",
				"stop": "now"
			}
		},
		{
			"id": "sum",
			"kind": "sum"
		}
	],
	"edges":[
		{"parent":"from","child":"range"},
		{"parent":"range","child":"sum"}
	]
}
	`)

	// Ensure we can properly unmarshal a query
	gotQ := flux.Spec{}
	if err := json.Unmarshal(srcData, &gotQ); err != nil {
		t.Fatal(err)
	}
	expQ := flux.Spec{
		Operations: []*flux.Operation{
			{
				ID: "from",
				Spec: &influxdb.FromOpSpec{
					Bucket: "mybucket",
				},
			},
			{
				ID: "range",
				Spec: &universe.RangeOpSpec{
					Start: flux.Time{
						Relative:   -4 * time.Hour,
						IsRelative: true,
					},
					Stop: flux.Time{
						IsRelative: true,
					},
				},
			},
			{
				ID:   "sum",
				Spec: &universe.SumOpSpec{},
			},
		},
		Edges: []flux.Edge{
			{Parent: "from", Child: "range"},
			{Parent: "range", Child: "sum"},
		},
	}
	if !cmp.Equal(gotQ, expQ, ignoreUnexportedQuerySpec) {
		t.Errorf("unexpected query:\n%s", cmp.Diff(gotQ, expQ, ignoreUnexportedQuerySpec))
	}

	// Ensure we can properly marshal a query
	data, err := json.Marshal(expQ)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &gotQ); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(gotQ, expQ, ignoreUnexportedQuerySpec) {
		t.Errorf("unexpected query after marshalling: -want/+got %s", cmp.Diff(expQ, gotQ, ignoreUnexportedQuerySpec))
	}
}

func TestSpec_Walk(t *testing.T) {
	testCases := []struct {
		query     *flux.Spec
		walkOrder []flux.OperationID
		err       error
	}{
		{
			query: &flux.Spec{},
			err:   errors.New("query has no root nodes"),
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
				},
			},
			err: errors.New("edge references unknown child operation \"c\""),
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "b"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "b"},
				},
			},
			err: errors.New("found duplicate operation ID \"b\""),
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
					{Parent: "d", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []flux.OperationID{
				"a", "b", "c", "d",
			},
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []flux.OperationID{
				"a", "c", "b", "d",
			},
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []flux.OperationID{
				"b", "a", "c", "d",
			},
		},
		{
			query: &flux.Spec{
				Operations: []*flux.Operation{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []flux.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
				},
			},
			walkOrder: []flux.OperationID{
				"b", "d", "a", "c",
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var gotOrder []flux.OperationID
			err := tc.query.Walk(func(o *flux.Operation) error {
				gotOrder = append(gotOrder, o.ID)
				return nil
			})
			if tc.err == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error: %q", tc.err)
				} else if got, exp := err.Error(), tc.err.Error(); got != exp {
					t.Fatalf("unexpected errors: got %q exp %q", got, exp)
				}
			}

			if !cmp.Equal(gotOrder, tc.walkOrder) {
				t.Fatalf("unexpected walk order -want/+got %s", cmp.Diff(tc.walkOrder, gotOrder))
			}
		})
	}
}

// Example_option demonstrates retrieving an option value from a scope object
func Example_option() {

	// Import the universe package.
	importer := flux.StdLib()
	universe, _ := importer.ImportPackageObject("universe")

	// Retrieve the default value for the now option
	nowFunc, _ := universe.Get("now")

	// The now option is a function value whose default behavior is to return
	// the current system time when called. The function now() doesn't take
	// any arguments so can be called with nil.
	nowTime, _ := nowFunc.Function().Call(dependenciestest.Default().Inject(context.TODO()), nil)
	fmt.Fprintf(os.Stderr, "The current system time (UTC) is: %v\n", nowTime)
	// Output:
}

// TODO(algow): This method doesn't work since you cannot set new options from outside of the package.
// Example_setOption demonstrates setting an option value on a scope object
// func Example_setOption() {
//
// 	// Import the universe package.
// 	importer := flux.StdLib()
// 	universe, _ := importer.ImportPackageObject("universe")
//
// 	// Create a new option binding
// 	universe.Set("dummy_option", values.NewInt(3))
//
// 	v, _ := universe.Lookup("dummy_option")
//
// 	fmt.Printf("dummy_option = %d", v.Int())
// 	// Output: dummy_option = 3
// }

// Example_overrideDefaultOptionExternally demonstrates how declaring an option
// in a Flux script will change that option's binding globally.
func Example_overrideDefaultOptionExternally() {
	queryString := `
		option now = () => 2018-07-13T00:00:00Z
		what_time_is_it = now()`

	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := flux.Eval(ctx, queryString)
	if err != nil {
		fmt.Println(err)
	}

	// After evaluating the package, lookup the value of what_time_is_it
	now, _ := scope.Lookup("what_time_is_it")

	// what_time_is_it? Why it's ....
	fmt.Printf("The new current time (UTC) is: %v", now)
	// Output: The new current time (UTC) is: 2018-07-13T00:00:00.000000000Z
}

// Example_overrideDefaultOptionInternally demonstrates how one can override a default
// option that is used in a query before that query is evaluated by the interpreter.
// func Example_overrideDefaultOptionInternally() {
// 	queryString := `what_time_is_it = now()`
//
// 	ctx := dependenciestest.Default().Inject(context.Background())
//
// 	importer := flux.StdLib()
// 	universe, _ := importer.ImportPackageObject("universe")
//
// 	// Define a new now function which returns a static time value of 2018-07-13T00:00:00.000000000Z
// 	timeValue := time.Date(2018, 7, 13, 0, 0, 0, 0, time.UTC)
// 	functionName := "newTime"
// 	// TODO (algow): determine correct type
// 	functionType := semantic.NewFunctionType(semantic.MonoType{}, nil)
// 	functionCall := func(ctx context.Context, args values.Object) (values.Value, error) {
// 		return values.NewTime(values.ConvertTime(timeValue)), nil
// 	}
// 	sideEffect := false
//
// 	newNowFunc := values.NewFunction(functionName, functionType, functionCall, sideEffect)
//
// 	// Override the default now function with the new one
//  values.SetOption(universe, "now", newNowFunc)
//
// 	// Evaluate package
// 	_, err := itrp.Eval(ctx, semPkg, universe, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
//
// 	// After evaluating the package, lookup the value of what_time_is_it
// 	now, _ := universe.Lookup("what_time_is_it")
//
// 	// what_time_is_it? Why it's ....
// 	fmt.Printf("The new current time (UTC) is: %v", now)
// 	// Output: The new current time (UTC) is: 2018-07-13T00:00:00.000000000Z
// }
