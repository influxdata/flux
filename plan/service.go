package plan

import (
	"context"

	"github.com/influxdata/flux/dependencies/plan"
)

func DefaultPlanService() plan.Service {
	return planService{}
}

type planService struct{}

func (s planService) Plan(ctx context.Context, p plan.Spec) (plan.Spec, error) {
	pspec := p.(*Spec)
	pb := PlannerBuilder{}

	// TODO(cwolff): handle options here as before
	//	planOptions := opts.planOptions
	//
	//	lopts := planOptions.logical
	//	popts := planOptions.physical
	//
	//	pb.AddLogicalOptions(lopts...)
	//	pb.AddPhysicalOptions(popts...)

	ps, err := pb.Build().Plan(ctx, pspec)
	if err != nil {
		return nil, err
	}
	return ps, nil
}
