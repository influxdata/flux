package experimental

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	TableKind = "table"
)

type TableOpSpec struct {
	Rows values.Array
}

func init() {
	tableSignature := runtime.MustLookupBuiltinType("experimental", "table")
	runtime.RegisterPackageValue("experimental", "table", flux.MustValue(flux.FunctionValue(TableKind, createTableOpSpec, tableSignature)))
	flux.RegisterOpSpec(TableKind, newTableOp)
	plan.RegisterProcedureSpec(TableKind, newTableProcedure, TableKind)
	execute.RegisterSource(TableKind, createTableSource)
}

func createTableOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(TableOpSpec)

	if rows, err := args.GetRequired("rows"); err != nil {
		return nil, err
	} else {
		if rows.Type().Nature() != semantic.Array {
			return nil, errors.Newf(codes.Invalid, "row data must be an array of records, got %s", rows.Type())
		}
		spec.Rows = rows.Array()
	}

	return spec, nil
}

func newTableOp() flux.OperationSpec {
	return new(TableOpSpec)
}

func (s *TableOpSpec) Kind() flux.OperationKind {
	return TableKind
}

type TableProcedureSpec struct {
	plan.DefaultCost
	Rows values.Array
}

func newTableProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TableOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &TableProcedureSpec{
		Rows: spec.Rows,
	}, nil
}

func (s *TableProcedureSpec) Kind() plan.ProcedureKind {
	return TableKind
}

func (s *TableProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(TableProcedureSpec)
	*ns = *s
	return ns
}

func createTableSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*TableProcedureSpec)
	return &tableSource{
		id:   id,
		mem:  a.Allocator(),
		rows: spec.Rows,
	}, nil
}

type tableSource struct {
	id   execute.DatasetID
	mem  *memory.Allocator
	rows values.Array
	ts   []execute.Transformation
}

func (s *tableSource) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *tableSource) Run(ctx context.Context) {
	tbl, err := buildTable(s.rows, s.mem)
	if err == nil {
		for _, t := range s.ts {
			t.Process(s.id, tbl)
		}
	}

	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}

func buildTable(rows values.Array, mem *memory.Allocator) (flux.Table, error) {
	builder := execute.NewColListTableBuilder(execute.NewGroupKey(nil, nil), mem)
	if rows.Len() == 0 {
		return builder.Table()
	}
	first := rows.Get(0)
	if first.Type().Nature() != semantic.Object {
		return nil, errors.New(codes.Internal, "rows should have been a list of records")
	}
	typ := first.Type()

	l, _ := typ.NumProperties()
	colMap := make(map[string]int, l)
	for i := 0; i < l; i++ {
		rp, _ := typ.RowProperty(i)
		pt, _ := rp.TypeOf()
		ctyp := flux.ColumnType(pt)
		if ctyp == flux.TInvalid {
			return nil, errors.Newf(codes.Invalid, "cannot represent the type %v as column data", pt)
		}
		j, _ := builder.AddCol(flux.ColMeta{
			Label: rp.Name(),
			Type:  ctyp,
		})
		colMap[rp.Name()] = j
	}
	cols := builder.Cols()

	var err error
	rows.Range(func(i int, row values.Value) {
		if err != nil {
			return
		}
		r := row.Object()
		r.Range(func(col string, v values.Value) {
			if err != nil {
				return
			}
			j := colMap[col]
			if flux.ColumnType(v.Type()) != cols[j].Type {
				err = errors.New(codes.Internal, "records do not have a consistent type")
				return
			}
			builder.AppendValue(colMap[col], v)
		})
	})
	if err != nil {
		return nil, err
	}
	return builder.Table()
}
