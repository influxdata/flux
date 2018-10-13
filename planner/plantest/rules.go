package plantest

import "github.com/influxdata/flux/planner"

// CreateSimpleRuleFn returns a function that creates a rule that will store any
// NodeIDs for the nodes it has visited in the slice pointed to by seenNodes.
func CreateSimpleRuleFn(seenNodes *[]planner.NodeID) func() planner.Rule {
	return func() planner.Rule {
		return &simpleRule{
			seenNodes: seenNodes,
		}
	}
}

type simpleRule struct {
	seenNodes *[]planner.NodeID
}

func (sr *simpleRule) Pattern() planner.Pattern {
	return planner.Any()
}

func (sr *simpleRule) Rewrite(node planner.PlanNode) (planner.PlanNode, bool) {
	*sr.seenNodes = append(*(sr.seenNodes), node.ID())
	return node, false
}

func (sr *simpleRule) Name() string {
	return "simple"
}

