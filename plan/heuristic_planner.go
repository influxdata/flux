package plan

import (
	"context"
	"fmt"
	"sort"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/testing"
	"github.com/influxdata/flux/internal/errors"
)

// heuristicPlanner applies a set of rules to the nodes in a Spec
// until a fixed point is reached and no more rules can be applied.
type heuristicPlanner struct {
	rules         map[ProcedureKind][]Rule
	disabledRules map[string]bool
}

func newHeuristicPlanner() *heuristicPlanner {
	return &heuristicPlanner{
		rules:         make(map[ProcedureKind][]Rule),
		disabledRules: make(map[string]bool),
	}
}

func (p *heuristicPlanner) addRules(rules ...Rule) {
	for _, rule := range rules {
		for _, root := range rule.Pattern().Roots() {
			ruleSlice := p.rules[root]
			p.rules[root] = append(ruleSlice, rule)
		}
	}
}

func (p *heuristicPlanner) removeRules(ruleNames ...string) {
	for _, n := range ruleNames {
		p.disabledRules[n] = true
	}
}

func (p *heuristicPlanner) clearRules() {
	p.rules = make(map[ProcedureKind][]Rule)
}

func applyRule(ctx context.Context, spec *Spec, rule Rule, node Node) (Node, bool, error) {
	newNode, changed, err := rule.Rewrite(ctx, node)
	if err != nil {
		return nil, false, err
	} else if changed {
		if newNode == nil {
			return nil, false, errors.Newf(
				codes.Internal,
				"rule %q returned a nil plan node even though it seems to have changed the plan",
				rule.Name(),
			)
		}
		testing.MarkInvokedPlannerRule(ctx, rule.Name())
		if err := updateSuccessors(spec, node, newNode); err != nil {
			return node, false, errors.Wrap(
				err,
				codes.Internal,
				fmt.Sprintf("updating successors after applying rule %q", rule.Name()),
			)
		}
		return newNode, true, nil
	}

	return node, changed, nil
}

// matchRules applies any applicable rules to the given plan node,
// and returns the rewritten plan node and whether or not any rewriting was done.
func (p *heuristicPlanner) matchRules(ctx context.Context, spec *Spec, node Node) (Node, bool, error) {
	anyChanged := false

	for _, rule := range p.rules[AnyKind] {
		if p.disabledRules[rule.Name()] {
			continue
		}
		if rule.Pattern().Match(node) {
			newNode, changed, err := applyRule(ctx, spec, rule, node)
			if err != nil {
				return nil, false, err
			}

			anyChanged = anyChanged || changed
			node = newNode
		}
	}

	for _, rule := range p.rules[node.Kind()] {
		if p.disabledRules[rule.Name()] {
			continue
		}
		if rule.Pattern().Match(node) {
			newNode, changed, err := applyRule(ctx, spec, rule, node)
			if err != nil {
				return nil, false, err
			}

			anyChanged = anyChanged || changed
			node = newNode
		}
	}

	return node, anyChanged, nil
}

// Plan is a fixed-point query planning algorithm.
// It traverses the DAG depth-first, attempting to apply rewrite rules at each node.
// Traversal is repeated until a pass over the DAG results in no changes with the given rule set.
//
// Plan may change its argument and/or return a new instance of Spec, so the correct way to call Plan is:
// plan, err = plan.Plan(plan)
func (p *heuristicPlanner) Plan(ctx context.Context, inputPlan *Spec) (*Spec, error) {
	for anyChanged := true; anyChanged; {
		visited := make(map[Node]struct{})
		visitedSuccessors := make(map[Node]int)
		// the plan is traversed starting from the sinks,
		// moving toward the sources.
		nodeStack := make([]Node, 0, len(inputPlan.Roots))
		for root := range inputPlan.Roots {
			nodeStack = append(nodeStack, root)
		}

		// Sort the roots so that we always traverse deterministically
		// (sort descending so that we pop off the stack in ascending order)
		sort.Slice(nodeStack, func(i, j int) bool {
			return nodeStack[i].ID() > nodeStack[j].ID()
		})

		anyChanged = false
		for len(nodeStack) > 0 {
			node := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[0 : len(nodeStack)-1]

			_, alreadyVisited := visited[node]
			visitable := visitedSuccessors[node] == len(node.Successors())

			if !alreadyVisited && visitable {
				newNode, changed, err := p.matchRules(ctx, inputPlan, node)
				if err != nil {
					return nil, err
				}
				anyChanged = anyChanged || changed

				// append to stack in reverse order so lower-indexed children
				// are visited first.
				for i := len(newNode.Predecessors()); i > 0; i-- {
					visitedSuccessors[newNode.Predecessors()[i-1]] += 1
					nodeStack = append(nodeStack, newNode.Predecessors()[i-1])
				}

				visited[node] = struct{}{}
			}
		}
	}

	return inputPlan, nil
}

// updateSuccessors looks at all the successors of oldNode
// and rewires them to point them at newNode.
// Predecessors of oldNode and newNode are not touched.
//
//	A   B             A   B     <-- successors
//	 \ /               \ /
//	 node  becomes   newNode
//	 / \               / \
//	D   E             D'  E'    <-- predecessors
func updateSuccessors(plan *Spec, oldNode, newNode Node) error {
	// no need to update successors if the node hasn't actually changed
	if oldNode == newNode {
		return nil
	}

	newNode.ClearSuccessors()

	if len(oldNode.Successors()) == 0 {
		// This is a new root (sink) node.
		plan.Replace(oldNode, newNode)
		return nil
	}

	for _, succ := range oldNode.Successors() {
		i := IndexOfNode(oldNode, succ.Predecessors())
		if i < 0 {
			return errors.Newf(
				codes.Internal,
				"inconsistent plan graph; successor %q does not have edge back to predecessor %q",
				succ.ID(), oldNode.ID(),
			)
		}
		succ.Predecessors()[i] = newNode
	}

	newNode.AddSuccessors(oldNode.Successors()...)
	return nil
}
