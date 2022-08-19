package experimental

import (
	"fmt"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const PreviewKind = "experimental.preview"

type PreviewOpSpec struct {
	NRows   int64
	NTables int64
}

func init() {
	previewSignature := runtime.MustLookupBuiltinType("experimental", "preview")

	runtime.RegisterPackageValue("experimental", "preview", flux.MustValue(flux.FunctionValue(PreviewKind, createPreviewOpSpec, previewSignature)))
	plan.RegisterProcedureSpec(PreviewKind, newPreviewProcedure, PreviewKind)
	execute.RegisterTransformation(PreviewKind, createPreviewTransformation)
}

func createPreviewOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(PreviewOpSpec)
	if nrows, ok, err := args.GetInt("nrows"); err != nil {
		return nil, err
	} else if ok {
		spec.NRows = nrows
	} else {
		spec.NRows = 5
	}

	if ntables, ok, err := args.GetInt("ntables"); err != nil {
		return nil, err
	} else if ok {
		spec.NTables = ntables
	} else {
		spec.NTables = 5
	}
	return spec, nil
}

func (s *PreviewOpSpec) Kind() flux.OperationKind {
	return PreviewKind
}

type PreviewProcedureSpec struct {
	plan.DefaultCost
	NRows   int64
	NTables int64
}

func newPreviewProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*PreviewOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	p := &PreviewProcedureSpec{
		NRows:   s.NRows,
		NTables: s.NTables,
	}
	return p, nil
}

func (s *PreviewProcedureSpec) Kind() plan.ProcedureKind {
	return PreviewKind
}
func (s *PreviewProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

func createPreviewTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*PreviewProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	return NewPreviewTransformation(id, s, a.Allocator())
}

type previewTransformation struct {
	nrows   int64
	ntables int64
}

func NewPreviewTransformation(id execute.DatasetID, spec *PreviewProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &previewTransformation{
		nrows:   spec.NRows,
		ntables: spec.NTables,
	}
	return execute.NewNarrowStateTransformation[any](id, tr, mem)
}

func (t *previewTransformation) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	n, ok := state.(int64)
	if !ok {
		if t.ntables == 0 {
			return nil, false, nil
		}
		t.ntables--
		n = t.nrows
	}

	if int64(chunk.Len()) <= n {
		chunk.Retain()
		if err := d.Process(chunk); err != nil {
			return nil, false, err
		}
		n -= int64(chunk.Len())
		return n, true, nil
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, chunk.NCols()),
	}
	for i := range chunk.Cols() {
		buffer.Values[i] = arrow.Slice(chunk.Values(i), 0, n)
	}

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, false, err
	}
	return 0, true, nil
}

func (t *previewTransformation) Close() error {
	return nil
}
