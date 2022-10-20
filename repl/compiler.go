package repl

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/plan"
)

// CompilerType specific to the Flux REPL
const CompilerType = "REPL"

// Compiler specific to the Flux REPL
type Compiler struct {
	Spec *operation.Spec `json:"spec"`
}

func (c Compiler) Compile(ctx context.Context, runtime flux.Runtime) (flux.Program, error) {
	lp, err := plan.CreateLogicalPlan(c.Spec)
	if err != nil {
		return nil, err
	}

	planSvc, err := flux.GetDependencies(ctx).PlanService()
	if err != nil {
		return nil, err
	}

	pp, err := planSvc.Plan(ctx, lp)
	if err != nil {
		return nil, err
	}

	return &lang.Program{
		PlanSpec: pp.(*plan.Spec),
	}, err
}

func (c Compiler) CompilerType() flux.CompilerType {
	return CompilerType
}

func AddCompilerToMappings(mappings flux.CompilerMappings) error {
	return mappings.Add(CompilerType, func() flux.Compiler {
		return new(Compiler)
	})
}
