package plan

import (
	"context"

	"github.com/influxdata/flux/dependencies/plan"
)

func DefaultPlanService() plan.Service {
	return defaultPlanService{}
}

type defaultPlanService struct{}

func (s defaultPlanService) Plan(ctx context.Context, p plan.Spec) (plan.Spec, error) {
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

func PlanLogically(ctx context.Context, p plan.Spec, rules ...Rule) (plan.Spec, error) {
	lp := NewLogicalPlanner(OnlyLogicalRules(rules...))
	newPlan, err := lp.Plan(ctx, p.(*Spec))
	if err != nil {
		return nil, err
	}
	return newPlan, nil
}

func PlanPhysically(ctx context.Context, p plan.Spec, rules ...Rule) (plan.Spec, error) {
	pp := NewPhysicalPlanner(OnlyPhysicalRules(rules...))
	newPlan, err := pp.Plan(ctx, p.(*Spec))
	if err != nil {
		return nil, err
	}
	return newPlan, nil
}

func PlanPhysicallySkipValidation(ctx context.Context, p plan.Spec, rules ...Rule) (plan.Spec, error) {
	pp := NewPhysicalPlanner(OnlyPhysicalRules(rules...), DisableValidation())
	newPlan, err := pp.Plan(ctx, p.(*Spec))
	if err != nil {
		return nil, err
	}
	return newPlan, nil
}

func CloneSpec(p plan.Spec) (plan.Spec, error) {
	return cloneSpec(p.(*Spec))
}

func ValidatePlan(p plan.Spec) error {
	return ValidatePhysicalPlan(p.(*Spec))
}
