package plantest

import "github.com/influxdata/flux/planner"

// SimpleRule is a simple rule whose pattern matches any plan node and
// just stores the NodeIDs of nodes it has visited in SeenNodes.
type SimpleRule struct {
	SeenNodes []planner.NodeID
}

func (sr *SimpleRule) Pattern() planner.Pattern {
	return planner.Any()
}

func (sr *SimpleRule) Rewrite(node planner.PlanNode) (planner.PlanNode, bool) {
	sr.SeenNodes = append(sr.SeenNodes, node.ID())
	return node, false
}

func (sr *SimpleRule) Name() string {
	return "simple"
}
