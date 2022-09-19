package universe

import (
	"context"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

func init() {
	plan.RegisterPhysicalRules(SortLimitRule{})
	execute.RegisterTransformation(SortLimitKind, createSortLimitTransformation)
}

const SortLimitKind = "sortLimit"

type SortLimitProcedureSpec struct {
	*SortProcedureSpec
	N int64
}

func (s *SortLimitProcedureSpec) Kind() plan.ProcedureKind {
	return SortLimitKind
}

func (s *SortLimitProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.SortProcedureSpec = s.SortProcedureSpec.Copy().(*SortProcedureSpec)
	return &ns
}

func createSortLimitTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SortLimitProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewSortLimitTransformation(id, s, a.Allocator())
}

type sortLimitTransformation struct {
	sortTransformation
	limit int64
}

func NewSortLimitTransformation(id execute.DatasetID, spec *SortLimitProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := sortLimitTransformation{
		sortTransformation: sortTransformation{
			mem:     mem,
			cols:    spec.Columns,
			compare: arrowutil.Compare,
		},
		limit: spec.N,
	}
	if spec.Desc {
		// If descending, use the descending comparison.
		t.compare = arrowutil.CompareDesc
	}
	return execute.NewAggregateTransformation(id, &t, mem)
}

func (s *sortLimitTransformation) Aggregate(chunk table.Chunk, state interface{}, mem memory.Allocator) (interface{}, bool, error) {
	var mh *sortTableMergeHeap
	if state != nil {
		mh = state.(*sortTableMergeHeap)
	} else {
		mh = &sortTableMergeHeap{
			cols:     chunk.Cols(),
			key:      chunk.Key(),
			sortCols: s.sortCols(chunk.Key(), chunk.Cols()),
			compare:  s.compare,
		}

	}

	// If the chunk is empty, ignore this chunk.
	// We still return the merge heap though as we still need to recognize that
	// the group key/schema exist.
	if chunk.Len() == 0 {
		return mh, true, nil
	}

	if err := s.appendChunk(mh, chunk, mem); err != nil {
		return nil, false, err
	}

	tbl, err := mh.Table(int(s.limit), mem)
	if err != nil {
		return nil, false, err
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		cr.Retain()
		mh.items = append(mh.items, &sortTableMergeHeapItem{cr: cr})
		return nil
	}); err != nil {
		return nil, false, err
	}
	return mh, true, nil
}

func (s *sortLimitTransformation) appendChunk(mh *sortTableMergeHeap, chunk table.Chunk, mem memory.Allocator) error {
	buffer := chunk.Buffer()
	buffer.Retain()
	s.reconcileSchema(mh, &buffer, mem)

	item := &sortTableMergeHeapItem{cr: &buffer}
	if !s.isSorted(&buffer, mh.sortCols) {
		item.indices = s.sort(&buffer, mh.sortCols)
		item.offset = int(item.indices.Value(0))
	}
	mh.items = append(mh.items, item)
	return nil
}

func (s *sortLimitTransformation) reconcileSchema(mh *sortTableMergeHeap, buffer *arrow.TableBuffer, mem memory.Allocator) {
	if len(buffer.Columns) == len(mh.cols) {
		equivalent := true
		for i, col := range mh.cols {
			if buffer.Columns[i] != col {
				equivalent = false
				break
			}
		}

		if equivalent {
			return
		}
	}

	vals := make([]array.Array, len(mh.cols))
	for i, col := range buffer.Columns {
		idx := execute.ColIdx(col.Label, mh.cols)
		if idx < 0 {
			// This column was not previously seen.
			// Backfill the schema and add null columns
			// in the relevant locations.
			mh.cols = append(mh.cols, col)
			if execute.ContainsStr(s.cols, col.Label) {
				mh.sortCols = append(mh.sortCols, len(mh.cols)-1)
			}
			s.backfillColumn(mh, len(mh.cols)-1, mem)
			vals = append(vals, buffer.Values[i])
			continue
		}
		vals[idx] = buffer.Values[i]
	}

	// If a previous column existed but doesn't exist in the current buffer,
	// we need to also backfill it.
	for i := range vals {
		if vals[i] == nil {
			vals[i] = arrow.Nulls(mh.cols[i].Type, buffer.Len(), mem)
		}
	}
	buffer.Columns = mh.cols
	buffer.Values = vals
}

func (s *sortLimitTransformation) backfillColumn(mh *sortTableMergeHeap, i int, mem memory.Allocator) {
	for _, item := range mh.items {
		cpy := &arrow.TableBuffer{
			GroupKey: item.cr.Key(),
			Columns:  mh.cols,
			Values:   make([]array.Array, len(mh.cols)),
		}
		for j := range item.cr.Cols() {
			cpy.Values[j] = table.Values(item.cr, j)
		}
		cpy.Values[i] = arrow.Nulls(mh.cols[i].Type, item.cr.Len(), mem)
		item.cr = cpy
	}
}

func (s *sortLimitTransformation) Compute(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
	// The chunks are in sorted order already and already chunked.
	mh := state.(*sortTableMergeHeap)
	for _, item := range mh.items {
		chunk := table.ChunkFromReader(item.cr)
		if err := d.Process(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (s *sortLimitTransformation) Close() error {
	return nil
}

type SortLimitRule struct{}

func (s SortLimitRule) Name() string {
	return "SortLimitRule"
}

func (s SortLimitRule) Pattern() plan.Pattern {
	return plan.MultiSuccessor(LimitKind, plan.SingleSuccessor(SortKind))
}

func (s SortLimitRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	limitSpec := node.ProcedureSpec().(*LimitProcedureSpec)
	if limitSpec.Offset != 0 {
		return node, false, nil
	}
	sortNode := node.Predecessors()[0]
	sortSpec := sortNode.ProcedureSpec().(*SortProcedureSpec)

	sortLimitSpec := &SortLimitProcedureSpec{
		SortProcedureSpec: sortSpec,
		N:                 limitSpec.N,
	}

	n, err := plan.MergeToPhysicalNode(node, sortNode, sortLimitSpec)
	if err != nil {
		return nil, false, err
	}
	return n, true, nil
}
