package execute

import (
	"context"
	"fmt"

	"github.com/influxdata/flux/metadata"
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
	Metadata() metadata.Metadata
}

type Source interface {
	Node
	Run(ctx context.Context)
	//SetLabel(label string)
	//Label() string
}

type CreateSource func(spec plan.ProcedureSpec, id DatasetID, ctx Administration) (Source, error)

var procedureToSource = make(map[plan.ProcedureKind]CreateSource)

func RegisterSource(k plan.ProcedureKind, c CreateSource) {
	if procedureToSource[k] != nil {
		panic(fmt.Errorf("duplicate registration for source with procedure kind %v", k))
	}
	procedureToSource[k] = c
}

type ExecutionNode struct {
	label string
}

func (n *ExecutionNode) SetLabel(label string) {
	n.label = label
}

func (n *ExecutionNode) Label() string {
	return n.label
}
