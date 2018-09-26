package planner

// Cost stores various dimensions of the cost of a query plan
type Cost struct {
	Disk int
	CPU  int
	GPU  int
	MEM  int
	NET  int
}

// Add two cost structures together
func Add(a Cost, b Cost) Cost {
	return Cost{
		Disk: a.Disk + b.Disk,
		CPU:  a.CPU + b.CPU,
		GPU:  a.GPU + b.GPU,
		MEM:  a.MEM + b.MEM,
		NET:  a.NET + b.NET,
	}
}

// PhysicalCost computes the total cost of a physical query plan
func PhysicalCost(node PhysicalPlan) Cost {
	var cumulative Cost
	for _, pred := range node.Predecessors() {
		cost := PhysicalCost(pred)
		cumulative = Add(cumulative, cost)
	}
	return Add(node.Cost(), cumulative)
}
