package flux_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

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
