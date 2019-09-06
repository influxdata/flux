package execute

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
)

// Transformation represents functions that stream a set of tables, performs
// data processing on them and produces an output stream of tables
type Transformation interface {
	RetractTable(id DatasetID, key flux.GroupKey) error
	// Process takes in one flux Table, performs data processing on it and
	// writes that table to a DataCache
	Process(id DatasetID, tbl flux.Table) error
	UpdateWatermark(id DatasetID, t Time) error
	UpdateProcessingTime(id DatasetID, t Time) error
	// Finish indicates that the Transformation is done processing. It is
	// the last method called on the Transformation
	Finish(id DatasetID, err error)
}

// StreamContext represents necessary context for a single stream of
// query data.
type StreamContext interface {
	Bounds() *Bounds
}

type Administration interface {
	Context() context.Context

	ResolveTime(qt flux.Time) Time
	StreamContext() StreamContext
	Allocator() *memory.Allocator
	Parents() []DatasetID

	Dependencies() Dependencies
}

// Dependencies represents the provided dependencies to the execution environment.
// The dependencies is opaque.
type Dependencies map[string]interface{}

type CreateTransformation func(id DatasetID, mode AccumulationMode, spec plan.ProcedureSpec, a Administration) (Transformation, Dataset, error)
type CreateNewPlannerTransformation func(id DatasetID, mode AccumulationMode, spec plan.ProcedureSpec, a Administration) (Transformation, Dataset, error)

var procedureToTransformation = make(map[plan.ProcedureKind]CreateNewPlannerTransformation)

// RegisterTransformation adds a new registration mapping of procedure kind to transformation.
func RegisterTransformation(k plan.ProcedureKind, c CreateNewPlannerTransformation) {
	if procedureToTransformation[k] != nil {
		panic(fmt.Errorf("duplicate registration for transformation with procedure kind %v", k))
	}
	procedureToTransformation[k] = c
}

// ReplaceTransformation changes an existing transformation registration.
func ReplaceTransformation(k plan.ProcedureKind, c CreateNewPlannerTransformation) {
	if procedureToTransformation[k] == nil {
		panic(fmt.Errorf("missing registration for transformation with procedure kind %v", k))
	}
	procedureToTransformation[k] = c
}
