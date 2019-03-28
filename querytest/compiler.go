package querytest

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/plan"
)

// ReplaceSpecCompiler provides a compiler that will produce programs
// that are modified my the given rule, e.g., so you can replace influxdb.from
// with csv.from.
type ReplaceSpecCompiler struct {
	Spec *flux.Spec
	Rule plan.Rule
}

func NewReplaceSpecCompiler(rule plan.Rule) *ReplaceSpecCompiler {
	return &ReplaceSpecCompiler{
		Rule: rule,
	}
}

func (c *ReplaceSpecCompiler) Compile(ctx context.Context) (flux.Program, error) {
	pb := &plan.PlannerBuilder{}
	pb.AddLogicalOptions(plan.AddLogicalRules(c.Rule))
	planner := pb.Build()
	plan, err := planner.Plan(c.Spec)
	if err != nil {
		return nil, err
	}

	return lang.NewProgram(plan), nil
}
