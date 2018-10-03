package planner

// Planner implements a rule-based query planner
type Planner interface {
	// Add rules to be enacted by the planner
	AddRules([]Rule)

	// Remove rules no longer used by the planner
	RemoveRules([]Rule)

	// Plan takes an initial query plan and returns an optimized plan
	Plan(PlanNode) (PlanNode, error)
}

type LogicalToPhysicalPlanner struct {
	rules      map[ProcedureKind]Rule
	falseMatch map[PlanNode]bool
}

func NewLogicalToPhysicalPlanner(rules []Rule) *LogicalToPhysicalPlanner {
	transformations := make(map[ProcedureKind]Rule, len(rules))
	for _, rule := range rules {
		transformations[rule.Pattern().Root()] = rule
	}
	return &LogicalToPhysicalPlanner{
		rules: transformations,
	}
}

// matchRule matches a rule in the plan. Nodes that match
// a rule but cannot be rewritten are skipped if encountered.
func (p *LogicalToPhysicalPlanner) matchRule(root PlanNode, rule Rule) (PlanNode, bool) {
	if rule.Pattern().Match(root) && !p.falseMatch[root] {
		return root, true
	}
	for _, pred := range root.Predecessors() {
		if node, matched := p.matchRule(pred, rule); matched {
			return node, matched
		}
	}
	return nil, false
}

func (p LogicalToPhysicalPlanner) AddRules(rules []Rule) {
	for _, rule := range rules {
		p.rules[rule.Pattern().Root()] = rule
	}
}

func (p LogicalToPhysicalPlanner) RemoveRules(rules []Rule) {
	for _, rule := range rules {
		delete(p.rules, rule.Pattern().Root())
	}
}

// Plan is a fixed-point query planning algorithm.
// It enacts each rule on the query plan until no more rewrites are possible.
func (p LogicalToPhysicalPlanner) Plan(root PlanNode) (PlanNode, error) {
	var transformed bool
	var newNode PlanNode

	for _, rule := range p.rules {

		// Try to match the root node
		node, matched := p.matchRule(root, rule)

		if matched && node == root {

			if newNode, transformed = rule.Rewrite(node); !transformed {
				// Record false match if root cannot be rewritten
				p.falseMatch[node] = true
			} else {
				// Reassign root if successfully rewritten
				root = newNode
			}
		}

		for {

			// No match means rule will never match
			if node, matched = p.matchRule(root, rule); !matched {
				break
			}

			if newNode, transformed = rule.Rewrite(node); !transformed {
				p.falseMatch[node] = true
			} else {
				replacePlanNode(node, newNode)
			}
		}
	}
	return root, nil
}

//  A   B             A   B
//   \ /               \ /
//   node  becomes   newNode
//   / \               / \
//  D   E             D'  E'
func replacePlanNode(node, newNode PlanNode) {
	for _, n := range node.Successors() {
		newNode.AddSuccessors(n)
		n.RemovePredecessor(node)
		n.AddPredecessors(newNode)
	}
}
