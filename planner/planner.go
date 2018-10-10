package planner

import "github.com/influxdata/flux"

// Planner implements a rule-based query planner
type Planner interface {
	// Add rules to be enacted by the planner
	AddRules([]Rule)


	ConvertID()

	// Plan takes an initial query plan and returns an optimized plan
	Plan(*QueryPlan) (*QueryPlan, error)
}

type LogicalToPhysicalPlanner struct {
	rules      map[ProcedureKind][]Rule
}

func (ltpp *LogicalToPhysicalPlanner) ConvertID(foid flux.OperationID) ProcedureID {
	return ProcedureIDFromOperationID(foid)
}

func NewLogicalToPhysicalPlanner(rules []Rule) *LogicalToPhysicalPlanner {
	transformations := make(map[ProcedureKind][]Rule, len(rules))
	for _, rule := range rules {
		kindRules := transformations[rule.Pattern().Root()]
		transformations[rule.Pattern().Root()] = append(kindRules, rule)
	}
	return &LogicalToPhysicalPlanner{
		rules: transformations,
	}
}

// matchRules applies any applicable rules to the given plan node,
// and returns the rewritten plan node and whether or not any rewriting was done.
func (p *LogicalToPhysicalPlanner) matchRules(node PlanNode) (PlanNode, bool) {
	anyChanged := false

	for _, rule := range p.rules[AnyKind] {
		newNode, changed := rule.Rewrite(node)
		anyChanged = anyChanged || changed
		node = newNode
	}

	for _, rule := range p.rules[node.Kind()] {
		newNode, changed := rule.Rewrite(node)
		anyChanged = anyChanged || changed
		node = newNode
	}

	return node, anyChanged
}

func (p LogicalToPhysicalPlanner) AddRules(rules []Rule) {
	for _, rule := range rules {
		ruleSlice := p.rules[rule.Pattern().Root()]
		p.rules[rule.Pattern().Root()] = append(ruleSlice, rule)
	}
}

// Plan is a fixed-point query planning algorithm.
// It traverses the DAG depth-first, attempting to apply rewrite rules at each node.
// Traversal is repeated until a pass over the DAG results in no changes with the given rule set.
//
// Plan may change its argument and/or return a new instance of QueryPlan, so the correct way to call Plan is:
//     plan, err = planner.Plan(plan)
func (p *LogicalToPhysicalPlanner) Plan(inputPlan *QueryPlan) (*QueryPlan, error) {

	for anyChanged := true; anyChanged == true; {

		visited := make(map[PlanNode]struct{})

		nodeStack := make([]PlanNode, len(inputPlan.Roots()))
		copy(nodeStack, inputPlan.Roots())

		anyChanged = false
		for ; len(nodeStack) > 0; {
			node := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[0 : len(nodeStack)-1]

			_, alreadyVisited := visited[node]

			if ! alreadyVisited {
				newNode, changed := p.matchRules(node)
				anyChanged = anyChanged || changed
				if node != newNode {
					updateSuccessors(inputPlan, node, newNode)
				}

				// append to stack in reverse order so lower-indexed children
				// are visited first.
				for i := len(newNode.Predecessors()); i > 0; i-- {
					nodeStack = append(nodeStack, newNode.Predecessors()[i - 1])
				}

				visited[newNode] = struct{}{}
			}
		}
	}

	return inputPlan, nil
}

// updateSuccessors looks at all the successors of oldNode
// and rewires them to point them at newNode.
// Predecessors of oldNode and newNode are not touched.
//
//  A   B             A   B     <-- successors
//   \ /               \ /
//   node  becomes   newNode
//   / \               / \
//  D   E             D'  E'    <-- predecessors
func updateSuccessors(plan *QueryPlan, oldNode, newNode PlanNode) {
	newNode.ClearSuccessors()

	if len(oldNode.Successors()) == 0 {
		// This is a new root node.
		plan.Replace(oldNode, newNode)
		return
	}

	for _, succ := range oldNode.Successors() {
		i := 0
		for ; i < len(succ.Predecessors()); i++ {
			succ.Predecessors()[i] = newNode
		}

		if i == len(succ.Predecessors()) {
			panic("Inconsistent plan graph: successor does not have edge back to predecessor")
		}
	}

	newNode.AddSuccessors(oldNode.Successors()...)
}
