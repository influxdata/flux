package universe

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const RowNumberKind = "rowNumber"

func init() {
	mt := runtime.MustLookupBuiltinType("universe", "rowNumber")
	runtime.RegisterPackageValue("universe", "rowNumber",
		flux.MustValue(flux.FunctionValue("rowNumber", createRowNumberOpSpec, mt)))
	plan.RegisterProcedureSpec(RowNumberKind, newRowNumberProcedure, RowNumberKind)
	execute.RegisterTransformation(RowNumberKind, createRowNumberTransformation)
}

func createRowNumberOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	return &RowNumberOpSpec{}, nil
}

type RowNumberOpSpec struct{}

func (s *RowNumberOpSpec) Kind() flux.OperationKind {
	return RowNumberKind
}

type RowNumberProcedureSpec struct {
	plan.DefaultCost
}

func newRowNumberProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	if _, ok := qs.(*RowNumberOpSpec); !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &RowNumberProcedureSpec{}, nil
}

func (s *RowNumberProcedureSpec) Kind() plan.ProcedureKind {
	return RowNumberKind
}

func (s *RowNumberProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

type rowNumberTransformation struct{}

func createRowNumberTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	if _, ok := spec.(*RowNumberProcedureSpec); !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewRowNumberTransformation(id, a.Allocator())
}

func NewRowNumberTransformation(id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return execute.NewNarrowStateTransformation(id, &rowNumberTransformation{}, mem)
}

func (r *rowNumberTransformation) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	row := 0
	if state != nil {
		row = state.(int)
	}

	buffer := chunk.Buffer()
	idx := chunk.Index("_index")
	if idx >= 0 {
		buffer.Columns = make([]flux.ColMeta, chunk.NCols())
	} else {
		buffer.Columns = make([]flux.ColMeta, chunk.NCols()+1)
		idx = chunk.NCols()
	}
	copy(buffer.Columns, chunk.Cols())
	buffer.Columns[idx] = flux.ColMeta{
		Label: "_index",
		Type:  flux.TInt,
	}
	buffer.Values = make([]array.Interface, len(buffer.Columns))
	for i, col := range chunk.Cols() {
		if col.Label == "_index" {
			continue
		}
		buffer.Values[i] = chunk.Values(i)
		buffer.Values[i].Retain()
	}

	b := array.NewIntBuilder(mem)
	b.Resize(chunk.Len())
	for i, n := 0, chunk.Len(); i < n; i++ {
		b.Append(int64(row + i + 1))
	}
	buffer.Values[idx] = b.NewArray()

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, false, err
	}
	return row + chunk.Len(), true, nil
}

func (r *rowNumberTransformation) Close() error {
	return nil
}
