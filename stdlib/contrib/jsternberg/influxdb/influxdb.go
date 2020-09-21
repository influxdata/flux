package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const pkgpath = "contrib/jsternberg/influxdb"

const maskKind = pkgpath + "._mask"

func init() {
	runtime.RegisterPackageValue(pkgpath, "_mask", flux.MustValue(flux.FunctionValue(
		"_mask",
		createMaskOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "_mask"),
	)))
	plan.RegisterProcedureSpec(maskKind, newMaskProcedure, maskKind)
	execute.RegisterTransformation(maskKind, createMaskTransformation)
}

type maskOpSpec struct {
	Columns []string
}

func createMaskOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(maskOpSpec)

	columns, err := args.GetRequiredArray("columns", semantic.String)
	if err != nil {
		return nil, err
	}

	spec.Columns = make([]string, columns.Len())
	columns.Range(func(i int, v values.Value) {
		spec.Columns[i] = v.Str()
	})
	return spec, nil
}

func (a *maskOpSpec) Kind() flux.OperationKind {
	return maskKind
}

type maskProcedureSpec struct {
	plan.DefaultCost
	Columns []string
}

func newMaskProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*maskOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &maskProcedureSpec{
		Columns: spec.Columns,
	}, nil
}

func (s *maskProcedureSpec) Kind() plan.ProcedureKind {
	return maskKind
}

func (s *maskProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(maskProcedureSpec)
	ns.Columns = make([]string, len(s.Columns))
	copy(ns.Columns, s.Columns)
	return ns
}

func createMaskTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*maskProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return newMaskTransformation(s, id)
}

type maskTransformation struct {
	execute.ExecutionNode
	d    *execute.PassthroughDataset
	spec *maskProcedureSpec
}

func newMaskTransformation(spec *maskProcedureSpec, id execute.DatasetID) (execute.Transformation, execute.Dataset, error) {
	t := &maskTransformation{
		d:    execute.NewPassthroughDataset(id),
		spec: spec,
	}
	return t, t.d, nil
}

func (t *maskTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *maskTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	outTable := table.Mask(tbl, t.spec.Columns)
	return t.d.Process(outTable)
}

func (t *maskTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *maskTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *maskTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
