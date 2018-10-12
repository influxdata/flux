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

var procedureToSource = make(map[planner.ProcedureKind]CreateNewPlannerSource)

func RegisterSource(k planner.ProcedureKind, c CreateNewPlannerSource) {
	if procedureToSource[k] != nil {
		panic(fmt.Errorf("duplicate registration for source with procedure kind %v", k))
	}
	procedureToSource[k] = c
}
