package execute

import (
	"context"
	"github.com/influxdata/flux/plan"
	"go.uber.org/zap"
)

// ConcurrencyQuotaFromPlan exposes the concurrency quota, supporting test
// cases in execute_test. The test cases there cannot be in the execute package
// due to circular dependencies.
func ConcurrencyQuotaFromPlan(ctx context.Context, p *plan.Spec, logger *zap.Logger) int {
	es := &executionState{
		p:         p,
		ctx:       ctx,
		resources: p.Resources,
		logger:    logger,
	}

	es.chooseDefaultResources(ctx, es.p)

	return es.resources.ConcurrencyQuota
}
