package plan

import (
	"errors"
	"sort"
)

// TopDownWalk will execute f for each plan node in the Spec.
// It always visits a node before visiting its predecessors.
func (plan *Spec) TopDownWalk(f func(node Node) error) error {
	visited := make(map[Node]struct{})

	roots := make([]Node, 0, len(plan.Roots))
	for root := range plan.Roots {
		roots = append(roots, root)
	}

	// Make sure to sort the roots first otherwise
	// an in-consistent walk order is possible.
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	postFn := func(Node) error {
		return nil
	}

	for _, root := range roots {
		err := walk(root, f, postFn, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

// BottomUpWalk will execute f for each plan node in the Spec,
// starting from the sources, and only visiting a node after all its
// predecessors have been visited.
func (plan *Spec) BottomUpWalk(f func(Node) error) error {
	visited := make(map[Node]struct{})

	roots := make([]Node, 0, len(plan.Roots))
	for root := range plan.Roots {
		roots = append(roots, root)
	}

	// Make sure to sort the roots first otherwise
	// an in-consistent walk order is possible.
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	preFn := func(Node) error {
		return nil
	}

	for _, root := range roots {
		err := walk(root, preFn, f, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

func walk(node Node, preFn, postFn func(Node) error, visited map[Node]struct{}) error {
	if _, ok := visited[node]; ok {
		return nil
	}

	visited[node] = struct{}{}

	// Pre-order traversal
	if err := preFn(node); err != nil {
		return err
	}

	for _, pred := range node.Predecessors() {
		if err := walk(pred, preFn, postFn, visited); err != nil {
			return err
		}
	}

	// Post-order traversal
	return postFn(node)
}

// WalkPredecessor visits every node in the plan rooted at `roots` in topological order,
// following predecessor links. If a cycle is detected, no node is visited and
// an error is returned.
func WalkPredecessors(roots []Node, visitFn func(node Node) error) error {
	tw := newTopologicalWalk(Node.Predecessors, visitFn)
	for _, root := range roots {
		if err := tw.walk(root); err != nil {
			return err
		}
	}

	return tw.visit()
}

// WalkSuccessors visits every node in the plan rooted at `roots` in topological order,
// following successor links. If a cycle is detected, no node is visited and
// an error is returned.
func WalkSuccessors(roots []Node, visitFn func(node Node) error) error {
	tw := newTopologicalWalk(Node.Successors, visitFn)
	for _, root := range roots {
		if err := tw.walk(root); err != nil {
			return err
		}
	}

	return tw.visit()
}

// TopologicalWalk visits every node in the plan in topological order.
// If a cycle is detected, no node is visited and an error is returned.
func (plan *Spec) TopologicalWalk(visitFn func(node Node) error) error {
	tw := newTopologicalWalk(Node.Predecessors, visitFn)

	roots := make([]Node, 0, len(plan.Roots))
	for root := range plan.Roots {
		roots = append(roots, root)
	}

	// Make sure to sort the roots first otherwise
	// an in-consistent walk order is possible.
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	for _, root := range roots {
		if err := tw.walk(root); err != nil {
			return err
		}
	}

	return tw.visit()
}

type topologicalWalk struct {
	navigationFn func(node Node) []Node
	visitFn      func(node Node) error

	temporaryMarks map[Node]bool
	permanentMarks map[Node]bool
	callStack      []func() error
}

func newTopologicalWalk(navigationFn func(node Node) []Node, visitFn func(node Node) error) *topologicalWalk {
	return &topologicalWalk{
		navigationFn:   navigationFn,
		visitFn:        visitFn,
		temporaryMarks: make(map[Node]bool),
		permanentMarks: make(map[Node]bool),
		callStack:      make([]func() error, 0),
	}
}

func (tw *topologicalWalk) pushVisit(node Node) {
	fn := func() error {
		return tw.visitFn(node)
	}
	tw.callStack = append(tw.callStack, fn)
}

func (tw *topologicalWalk) walk(node Node) error {
	if tw.temporaryMarks[node] {
		return errors.New("cycle detected")
	}

	if !tw.permanentMarks[node] {
		tw.temporaryMarks[node] = true

		for _, n := range tw.navigationFn(node) {
			if err := tw.walk(n); err != nil {
				return err
			}
		}

		tw.permanentMarks[node] = true
		tw.temporaryMarks[node] = false
		tw.pushVisit(node)
	}

	return nil
}

func (tw *topologicalWalk) visit() error {
	for i := len(tw.callStack) - 1; i >= 0; i-- {
		if err := tw.callStack[i](); err != nil {
			return err
		}
	}

	return nil
}
