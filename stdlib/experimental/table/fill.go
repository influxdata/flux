package table

import (
	"context"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const (
	pkgpath  = "experimental/table"
	FillKind = pkgpath + ".fill"
)

type FillOpSpec struct{}

func init() {
	fillSignature := runtime.MustLookupBuiltinType(pkgpath, "fill")

	runtime.RegisterPackageValue(pkgpath, "fill", flux.MustValue(flux.FunctionValue(FillKind, createFillOpSpec, fillSignature)))
	plan.RegisterProcedureSpec(FillKind, newFillProcedure, FillKind)
	plan.RegisterLogicalRules(IdempotentTableFill{})
	execute.RegisterTransformation(FillKind, createFillTransformation)
}

func createFillOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	return &FillOpSpec{}, nil
}

func (s *FillOpSpec) Kind() flux.OperationKind {
	return FillKind
}

type FillProcedureSpec struct {
	plan.DefaultCost
}

func newFillProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*FillOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &FillProcedureSpec{}, nil
}

func (s *FillProcedureSpec) Kind() plan.ProcedureKind {
	return FillKind
}
func (s *FillProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

func createFillTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	_, ok := spec.(*FillProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return newFillTransformation(id, a.Allocator())
}

type fillTransformation struct {
	execute.ExecutionNode
	d   *execute.PassthroughDataset
	mem memory.Allocator
}

func newFillTransformation(id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &fillTransformation{
		d:   execute.NewPassthroughDataset(id),
		mem: mem,
	}
	return t, t.d, nil
}

func (t *fillTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if !tbl.Empty() {
		return t.d.Process(tbl)
	}

	buf := arrow.TableBuffer{
		GroupKey: tbl.Key(),
		Columns:  tbl.Cols(),
		Values:   make([]array.Interface, len(tbl.Cols())),
	}
	for i, col := range buf.Columns {
		b := arrow.NewBuilder(col.Type, t.mem)
		if idx := execute.ColIdx(col.Label, buf.Key().Cols()); idx >= 0 {
			if err := arrow.AppendValue(b, buf.Key().Value(idx)); err != nil {
				return err
			}
		} else {
			b.AppendNull()
		}
		buf.Values[i] = b.NewArray()
	}
	out := table.FromBuffer(&buf)
	return t.d.Process(out)
}

func (t *fillTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *fillTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *fillTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *fillTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type IdempotentTableFill struct{}

func (i IdempotentTableFill) Name() string {
	return pkgpath + ".IdempotentTableFill"
}

func (i IdempotentTableFill) Pattern() plan.Pattern {
	return plan.Pat(FillKind, plan.Pat(FillKind, plan.Any()))
}

func (i IdempotentTableFill) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	parentNode := node.Predecessors()[0]
	node.ClearPredecessors()
	return parentNode, true, nil
}
