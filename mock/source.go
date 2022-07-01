package mock

import (
	"context"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

// Source is a mock source that performs the given functions.
// By default it does nothing.
type Source struct {
	execute.ExecutionNode
	AddTransformationFn func(transformation execute.Transformation)
	RunFn               func(ctx context.Context)
}

func (s *Source) AddTransformation(t execute.Transformation) {
	if s.AddTransformationFn != nil {
		s.AddTransformationFn(t)
	}
}

func (s *Source) Run(ctx context.Context) {
	if s.RunFn != nil {
		s.RunFn(ctx)
	}
}

// CreateMockFromSource will register a mock "from" source.  Use it like this in the init()
// of your test:
//    execute.RegisterSource(influxdb.FromKind, mock.CreateMockFromSource)
func CreateMockFromSource(spec plan.ProcedureSpec, id execute.DatasetID, ctx execute.Administration) (execute.Source, error) {
	return &Source{}, nil
}
