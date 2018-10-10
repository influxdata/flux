package execute

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/planner"
)

type Node interface {
	AddTransformation(t Transformation)
}

type Source interface {
	Node
	Run(ctx context.Context)
}

type CreateSource func(spec plan.ProcedureSpec, id DatasetID, ctx Administration) (Source, error)
type CreateNewPlannerSource func(spec planner.ProcedureSpec, id DatasetID, ctx Administration) (Source, error)


var procedureToSource = make(map[plan.ProcedureKind]CreateSource)

func RegisterSource(k plan.ProcedureKind, c CreateNewPlannerSource) {
	if procedureToSource[k] != nil {
		panic(fmt.Errorf("duplicate registration for source with procedure kind %v", k))
	}

	createFn := func(spec plan.ProcedureSpec, id DatasetID, ctx Administration) (Source, error) {
		plannerProcSpec := spec.(planner.ProcedureSpec)
		return c(plannerProcSpec, id, ctx)
	}

	procedureToSource[k] = createFn
}
