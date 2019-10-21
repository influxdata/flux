package execute

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
)

type Node interface {
	AddTransformation(t Transformation)
}

// MetadataNode is a node that has additional metadata
// that should be added to the result after it is
// processed.
type MetadataNode interface {
	Node
	Metadata() flux.Metadata
}

type Source interface {
	Node
	Run(ctx context.Context)
}

type CreateSource func(spec plan.ProcedureSpec, id DatasetID, ctx Administration) (Source, error)

var procedureToSource = make(map[plan.ProcedureKind]CreateSource)

func RegisterSource(k plan.ProcedureKind, c CreateSource) {
	if procedureToSource[k] != nil {
		panic(fmt.Errorf("duplicate registration for source with procedure kind %v", k))
	}
	procedureToSource[k] = c
}
