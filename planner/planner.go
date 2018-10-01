package planner

// Planner implements a rule-based query planner
type Planner interface {
	// Add rules to be enacted by the planner
	AddRules([]Rule)

	// Remove rules no longer used by the planner
	RemoveRules([]Rule)

	// Plan takes an initial query plan and returns an optimized plan
	Plan(PlanNode) PlanNode
}
