package promql

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const EmptyTableKind = "emptyTable"

type EmptyTableOpSpec struct{}

func init() {
	emptyTableSignature := runtime.MustLookupBuiltinType("internal/promql", "emptyTable")
	runtime.RegisterPackageValue("internal/promql", "emptyTable", flux.MustValue(flux.FunctionValue(EmptyTableKind, createEmptyTableOpSpec, emptyTableSignature)))
	plan.RegisterProcedureSpec(EmptyTableKind, newEmptyTableProcedure, EmptyTableKind)
	execute.RegisterSource(EmptyTableKind, createEmptyTableSource)
}

func createEmptyTableOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	return new(EmptyTableOpSpec), nil
}

func (s *EmptyTableOpSpec) Kind() flux.OperationKind {
	return EmptyTableKind
}

type EmptyTableProcedureSpec struct {
	plan.DefaultCost
}

func newEmptyTableProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*EmptyTableOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return new(EmptyTableProcedureSpec), nil
}

func (s *EmptyTableProcedureSpec) Kind() plan.ProcedureKind {
	return EmptyTableKind
}

func (s *EmptyTableProcedureSpec) Copy() plan.ProcedureSpec {
	return new(EmptyTableProcedureSpec)
}

func createEmptyTableSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	_, ok := prSpec.(*EmptyTableProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	return &EmptyTableSource{id: dsid}, nil
}

type EmptyTableSource struct {
	execute.ExecutionNode
	id execute.DatasetID
	ts execute.TransformationSet
}

func (s *EmptyTableSource) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *EmptyTableSource) Run(ctx context.Context) {
	var err error
	var tbl flux.Table

	startCol := flux.ColMeta{Label: execute.DefaultStartColLabel, Type: flux.TTime}
	stopCol := flux.ColMeta{Label: execute.DefaultStopColLabel, Type: flux.TTime}
	timeCol := flux.ColMeta{Label: execute.DefaultTimeColLabel, Type: flux.TTime}
	valueCol := flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat}

	key := execute.NewGroupKey(
		[]flux.ColMeta{
			startCol,
			stopCol,
		},
		[]values.Value{
			values.NewTime(values.ConvertTime(time.Time{})),
			values.NewTime(values.ConvertTime(time.Time{})),
		},
	)

	builder := execute.NewColListTableBuilder(key, &memory.ResourceAllocator{})

	for _, c := range []flux.ColMeta{startCol, stopCol, timeCol, valueCol} {
		if _, err = builder.AddCol(c); err != nil {
			goto FINISH
		}
	}

	tbl, err = builder.Table()
	if err != nil {
		goto FINISH
	}

	s.ts.Process(s.id, tbl)

FINISH:
	for _, t := range s.ts {
		if err != nil {
			err = fmt.Errorf("error in promql.emptyTable(): %w", err)
		}
		t.Finish(s.id, err)
	}
}
