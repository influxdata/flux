package plan

import (
	"github.com/influxdata/flux"
)

// PlannerBuilder provides clients with an easy way to create planners.
type PlannerBuilder struct {
	lopts []LogicalOption
	popts []PhysicalOption
}

// AddLogicalOptions lets callers specify attributes of the logical planner
// that will be part of the created planner.
func (pb *PlannerBuilder) AddLogicalOptions(lopt ...LogicalOption) {
	pb.lopts = append(pb.lopts, lopt...)
}

// AddPhysicalOptions lets callers specify attributes of the physical planner
// that will be part of the created planner.
func (pb *PlannerBuilder) AddPhysicalOptions(popt ...PhysicalOption) {
	pb.popts = append(pb.popts, popt...)
}

// Build builds a planner with specified attributes.
func (pb PlannerBuilder) Build() Planner {
	return &planner{
		lp: NewLogicalPlanner(pb.lopts...),
		pp: NewPhysicalPlanner(pb.popts...),
	}
}

type planner struct {
	lp LogicalPlanner
	pp PhysicalPlanner
}

func (p *planner) Plan(fspec *flux.Spec) (*Spec, error) {
	ip, err := p.lp.CreateInitialPlan(fspec)
	if err != nil {
		return nil, err
	}
	lp, err := p.lp.Plan(ip)
	if err != nil {
		return nil, err
	}
	pp, err := p.pp.Plan(lp)
	if err != nil {
		return nil, err
	}
	return pp, nil
}
