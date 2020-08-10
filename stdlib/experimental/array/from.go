package array

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	FromKind = "experimental/array.from"
)

type FromOpSpec struct {
	Rows values.Array
}

func init() {
	fromSignature := runtime.MustLookupBuiltinType("experimental/array", "from")
	runtime.RegisterPackageValue("experimental/array", "from", flux.MustValue(flux.FunctionValue(FromKind, createFromOpSpec, fromSignature)))
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if rows, err := args.GetRequired("rows"); err != nil {
		return nil, err
	} else {
		if rows.Type().Nature() != semantic.Array {
			return nil, errors.Newf(codes.Invalid, "row data must be an array of records, got %s", rows.Type())
		}
		spec.Rows = rows.Array()
	}

	if spec.Rows.Len() == 0 {
		return nil, errors.New(codes.Invalid, "rows must be non-empty")
	}

	return spec, nil
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	Rows values.Array
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Rows: spec.Rows,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	*ns = *s
	return ns
}

func createFromSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*FromProcedureSpec)
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
	ts   execute.TransformationSet
}

func (s *tableSource) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *tableSource) Run(ctx context.Context) {
	tbl, err := buildTable(s.rows, s.mem)
	if err == nil {
		err = s.ts.Process(s.id, tbl)
	}

	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}

func buildTable(rows values.Array, mem *memory.Allocator) (flux.Table, error) {
	typ, err := rows.Type().ElemType()
	if err != nil {
		return nil, err
	} else if typ.Nature() != semantic.Object {
		return nil, errors.New(codes.Internal, "rows should have been a list of records")
	}

	l, err := typ.NumProperties()
	if err != nil {
		return nil, err
	}
	cols := make([]flux.ColMeta, 0, l)
	for i := 0; i < l; i++ {
		rp, err := typ.RowProperty(i)
		if err != nil {
			return nil, err
		}

		pt, err := rp.TypeOf()
		if err != nil {
			return nil, err
		}
		ctyp := flux.ColumnType(pt)
		if ctyp == flux.TInvalid {
			return nil, errors.Newf(codes.Invalid, "cannot represent the type %v as column data", pt)
		}
		cols = append(cols, flux.ColMeta{
			Label: rp.Name(),
			Type:  ctyp,
		})
	}

	key := execute.NewGroupKey(nil, nil)
	builder := table.NewArrowBuilder(key, mem)
	for _, col := range cols {
		i, err := builder.AddCol(col)
		if err != nil {
			return nil, err
		}
		builder.Builders[i].Resize(rows.Len())
	}

	if err := appendRows(builder, rows); err != nil {
		return nil, err
	}
	return builder.Table()
}

func appendRows(builder *table.ArrowBuilder, rows values.Array) (err error) {
	rows.Range(func(i int, row values.Value) {
		if err != nil {
			return
		}

		for j, col := range builder.Cols() {
			v, _ := row.Object().Get(col.Label)
			err = arrow.AppendValue(builder.Builders[j], v)
		}
	})
	return err
}
