package flux_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
)

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
