package operation_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/internal/operation"
)

func TestSpec_Walk(t *testing.T) {
	testCases := []struct {
		query     *operation.Spec
		walkOrder []operation.NodeID
		err       error
	}{
		{
			query: &operation.Spec{},
			err:   errors.New("query has no root nodes"),
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
				},
			},
			err: errors.New("edge references unknown child operation \"c\""),
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "b"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "b"},
				},
			},
			err: errors.New("found duplicate operation ID \"b\""),
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
					{Parent: "d", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []operation.NodeID{
				"a", "b", "c", "d",
			},
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
				},
				Edges: []operation.Edge{},
			},
			walkOrder: []operation.NodeID{
				"a",
			},
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []operation.NodeID{
				"a", "c", "b", "d",
			},
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []operation.NodeID{
				"b", "a", "c", "d",
			},
		},
		{
			query: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "a"},
					{ID: "b"},
					{ID: "c"},
					{ID: "d"},
				},
				Edges: []operation.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
				},
			},
			walkOrder: []operation.NodeID{
				"b", "d", "a", "c",
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var gotOrder []operation.NodeID
			err := tc.query.Walk(func(o *operation.Node) error {
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
