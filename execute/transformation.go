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
	//SetLabel(label string)
	//Label() string
}

// TransformationSet is a group of transformations.
type TransformationSet []Transformation

func (ts TransformationSet) RetractTable(id DatasetID, key flux.GroupKey) error {
	for _, t := range ts {
		if err := t.RetractTable(id, key); err != nil {
			return err
		}
	}
	return nil
}

func (ts TransformationSet) Process(id DatasetID, tbl flux.Table) error {
	if len(ts) == 0 {
		return nil
	} else if len(ts) == 1 {
		return ts[0].Process(id, tbl)
	}

	// There is more than one transformation so we need to
	// copy the table for each transformation.
	bufTable, err := CopyTable(tbl)
	if err != nil {
		return err
	}
	defer bufTable.Done()

	for _, t := range ts {
		if err := t.Process(id, bufTable.Copy()); err != nil {
			return err
		}
	}
	return nil
}

func (ts TransformationSet) UpdateWatermark(id DatasetID, time Time) error {
	for _, t := range ts {
		if err := t.UpdateWatermark(id, time); err != nil {
			return err
		}
	}
	return nil
}

func (ts TransformationSet) UpdateProcessingTime(id DatasetID, time Time) error {
	for _, t := range ts {
		if err := t.UpdateProcessingTime(id, time); err != nil {
			return err
		}
	}
	return nil
}

func (ts TransformationSet) Finish(id DatasetID, err error) {
	for _, t := range ts {
		t.Finish(id, err)
	}
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
}

type CreateTransformation func(id DatasetID, mode AccumulationMode, spec plan.ProcedureSpec, a Administration) (Transformation, Dataset, error)

var procedureToTransformation = make(map[plan.ProcedureKind]CreateTransformation)

// RegisterTransformation adds a new registration mapping of procedure kind to transformation.
func RegisterTransformation(k plan.ProcedureKind, c CreateTransformation) {
	if procedureToTransformation[k] != nil {
		panic(fmt.Errorf("duplicate registration for transformation with procedure kind %v", k))
	}
	procedureToTransformation[k] = c
}

// ReplaceTransformation changes an existing transformation registration.
func ReplaceTransformation(k plan.ProcedureKind, c CreateTransformation) {
	if procedureToTransformation[k] == nil {
		panic(fmt.Errorf("missing registration for transformation with procedure kind %v", k))
	}
	procedureToTransformation[k] = c
}
