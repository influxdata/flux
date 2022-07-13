package universe

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/opentracing/opentracing-go"
)

const (
	ParallelMergeKind = "ParallelMergeKind"
)

type PartitionMergeProcedureSpec struct {
	plan.DefaultCost
	Factor int
}

func (o *PartitionMergeProcedureSpec) OutputAttributes() plan.PhysicalAttributes {
	return plan.PhysicalAttributes{
		plan.ParallelMergeKey: plan.ParallelMergeAttribute{Factor: o.Factor},
	}
}

func (o *PartitionMergeProcedureSpec) RequiredAttributes() []plan.PhysicalAttributes {
	return []plan.PhysicalAttributes{
		{
			plan.ParallelRunKey: plan.ParallelRunAttribute{Factor: o.Factor},
		},
	}
}

func (o *PartitionMergeProcedureSpec) Kind() plan.ProcedureKind {
	return ParallelMergeKind
}

func (o *PartitionMergeProcedureSpec) Copy() plan.ProcedureSpec {
	return &PartitionMergeProcedureSpec{
		DefaultCost: o.DefaultCost,
		Factor:      o.Factor,
	}
}

func init() {
	execute.RegisterTransformation(ParallelMergeKind, createPartitionMergeTransformation)
}

func createPartitionMergeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*PartitionMergeProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	alloc := a.Allocator()

	d := execute.NewPassthroughDataset(id)

	t, err := NewPartitionMergeTransformation(a.Context(), d, alloc, s, a.Parents())
	if err != nil {
		return nil, nil, err
	}

	return t, d, nil
}

type PartitionMergeTransformation struct {
	execute.ExecutionNode
	ctx     context.Context
	dataset *execute.PassthroughDataset
	span    opentracing.Span
	alloc   memory.Allocator

	mu               sync.Mutex
	predecessorState map[execute.DatasetID]*parallelPredecessorState
	finished         bool
}

type parallelPredecessorState struct {
	mark       execute.Time
	processing execute.Time
	finished   bool
}

func (t *PartitionMergeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.dataset.RetractTable(key)
}

func NewPartitionMergeTransformation(ctx context.Context, dataset *execute.PassthroughDataset, alloc memory.Allocator, spec *PartitionMergeProcedureSpec, predecessors []execute.DatasetID) (*PartitionMergeTransformation, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "PartitionMergeTransformation.Process")

	predecessorState := make(map[execute.DatasetID]*parallelPredecessorState, len(predecessors))
	for _, id := range predecessors {
		predecessorState[id] = new(parallelPredecessorState)
	}

	return &PartitionMergeTransformation{
		ctx:              ctx,
		dataset:          dataset,
		span:             span,
		alloc:            alloc,
		predecessorState: predecessorState,
	}, nil
}

func (t *PartitionMergeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	passthroughBuilder := table.NewBufferedBuilder(tbl.Key(), t.alloc)

	err := tbl.Do(func(er flux.ColReader) error {
		return passthroughBuilder.AppendBuffer(er)
	})
	if err != nil {
		return err
	}

	out, err := passthroughBuilder.Table()
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.finished {
		return nil
	}

	return t.dataset.Process(out)
}

func (t *PartitionMergeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.finished {
		return nil
	}

	t.predecessorState[id].mark = mark

	min := execute.Time(math.MaxInt64)
	for _, state := range t.predecessorState {
		if state.mark < min {
			min = state.mark
		}
	}

	return t.dataset.UpdateWatermark(min)
}

func (t *PartitionMergeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.finished {
		return nil
	}

	t.predecessorState[id].processing = pt

	min := execute.Time(math.MaxInt64)
	for _, state := range t.predecessorState {
		if state.processing < min {
			min = state.processing
		}
	}

	return t.dataset.UpdateProcessingTime(min)
}

func (t *PartitionMergeTransformation) Finish(id execute.DatasetID, err error) {
	defer t.span.Finish()

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.finished {
		return
	}

	t.predecessorState[id].finished = true

	if err != nil {
		// When we get an error, pass it on immediately. All other messages are
		// stopped.
		t.finished = true
	} else {
		t.finished = true
		for _, state := range t.predecessorState {
			t.finished = t.finished && state.finished
		}
	}

	if t.finished {
		t.dataset.Finish(err)
	}
}
