package operation

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
)

// Node denotes a single operation in a query.
type Node struct {
	ID     NodeID             `json:"id"`
	Spec   flux.OperationSpec `json:"spec"`
	Source NodeSource         `json:"source"`
}

// NodeSource specifies the source location that created
// an operation.
type NodeSource struct {
	Stack []interpreter.StackEntry `json:"stack"`
}

// NodeID is a unique ID within a query for the operation.
type NodeID string

// Spec specifies a query.
type Spec struct {
	Operations []*Node                 `json:"operations"`
	Edges      []Edge                  `json:"edges"`
	Resources  flux.ResourceManagement `json:"resources"`
	Now        time.Time               `json:"now"`

	// HasConflict is true if one of the operations in this spec
	// was previously used in another spec. This indicates
	// that two nodes from distinct specs relate to the same
	// operation.
	HasConflict bool

	sorted   []*Node
	children map[NodeID][]*Node
	parents  map[NodeID][]*Node
}

// Edge is a data flow relationship between a parent and a child
type Edge struct {
	Parent NodeID `json:"parent"`
	Child  NodeID `json:"child"`
}

// Walk calls f on each operation exactly once.
// The function f will be called on an operation only after
// all of its parents have already been passed to f.
func (q *Spec) Walk(f func(o *Node) error) error {
	if len(q.sorted) == 0 {
		if err := q.prepare(); err != nil {
			return err
		}
	}
	for _, o := range q.sorted {
		err := f(o)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate ensures the query is a valid DAG.
func (q *Spec) Validate() error {
	if q.Now.IsZero() {
		return errors.New(codes.Invalid, "now time must be set")
	}
	return q.prepare()
}

// Children returns a list of children for a given operation.
// If the query is invalid no children will be returned.
func (q *Spec) Children(id NodeID) []*Node {
	if q.children == nil {
		err := q.prepare()
		if err != nil {
			return nil
		}
	}
	return q.children[id]
}

// Parents returns a list of parents for a given operation.
// If the query is invalid no parents will be returned.
func (q *Spec) Parents(id NodeID) []*Node {
	if q.parents == nil {
		err := q.prepare()
		if err != nil {
			return nil
		}
	}
	return q.parents[id]
}

// prepare populates the internal datastructure needed to quickly navigate the query DAG.
// As a result the query DAG is validated.
func (q *Spec) prepare() error {
	q.sorted = q.sorted[0:0]

	parents, children, roots, err := q.determineParentsChildrenAndRoots()
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return errors.New(codes.Invalid, "query has no root nodes")
	}

	q.parents = parents
	q.children = children

	tMarks := make(map[NodeID]bool)
	pMarks := make(map[NodeID]bool)

	for _, r := range roots {
		if err := q.visit(tMarks, pMarks, r); err != nil {
			return err
		}
	}
	// reverse q.sorted
	for i, j := 0, len(q.sorted)-1; i < j; i, j = i+1, j-1 {
		q.sorted[i], q.sorted[j] = q.sorted[j], q.sorted[i]
	}
	return nil
}

func (q *Spec) computeLookup() (map[NodeID]*Node, error) {
	lookup := make(map[NodeID]*Node, len(q.Operations))
	for _, o := range q.Operations {
		if _, ok := lookup[o.ID]; ok {
			return nil, errors.Newf(codes.Internal, "found duplicate operation ID %q", o.ID)
		}
		lookup[o.ID] = o
	}
	return lookup, nil
}

func (q *Spec) determineParentsChildrenAndRoots() (parents, children map[NodeID][]*Node, roots []*Node, _ error) {
	lookup, err := q.computeLookup()
	if err != nil {
		return nil, nil, nil, err
	}
	children = make(map[NodeID][]*Node, len(q.Operations))
	parents = make(map[NodeID][]*Node, len(q.Operations))
	for _, e := range q.Edges {
		// Build children map
		c, ok := lookup[e.Child]
		if !ok {
			return nil, nil, nil, errors.Newf(codes.Internal, "edge references unknown child operation %q", e.Child)
		}
		children[e.Parent] = append(children[e.Parent], c)

		// Build parents map
		p, ok := lookup[e.Parent]
		if !ok {
			return nil, nil, nil, errors.Newf(codes.Internal, "edge references unknown parent operation %q", e.Parent)
		}
		parents[e.Child] = append(parents[e.Child], p)
	}
	// Find roots, i.e operations with no parents.
	for _, o := range q.Operations {
		if len(parents[o.ID]) == 0 {
			roots = append(roots, o)
		}
	}
	return
}

// Depth first search topological sorting of a DAG.
// https://en.wikipedia.org/wiki/Topological_sorting#Algorithms
func (q *Spec) visit(tMarks, pMarks map[NodeID]bool, o *Node) error {
	id := o.ID
	if tMarks[id] {
		return errors.New(codes.Invalid, "found cycle in query")
	}

	if !pMarks[id] {
		tMarks[id] = true
		for _, c := range q.children[id] {
			if err := q.visit(tMarks, pMarks, c); err != nil {
				return err
			}
		}
		pMarks[id] = true
		tMarks[id] = false
		q.sorted = append(q.sorted, o)
	}
	return nil
}

// Functions return the names of all functions used in the plan
func (q *Spec) Functions() ([]string, error) {
	funcs := []string{}
	err := q.Walk(func(o *Node) error {
		funcs = append(funcs, string(o.Spec.Kind()))
		return nil
	})
	return funcs, err
}
