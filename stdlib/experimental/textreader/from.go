package textreader

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
	TextReaderKind = "textreader.from"
)

type TextReaderOpSpec struct {
	Txt    string
	Header values.Array
}

func init() {
	textReaderFromSignature := runtime.MustLookupBuiltinType("textreader", "from")
	runtime.RegisterPackageValue("textreader", "from", flux.MustValue(flux.FunctionValue(TextReaderKind, createFromOpSpec, textReaderFromSignature)))
	plan.RegisterProcedureSpec(TextReaderKind, newFromProcedure, TextReaderKind)
	execute.RegisterSource(TextReaderKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(TextReaderOpSpec)

	if txt, err := args.GetRequiredString("txt"); err != nil {
		return nil, err
	} else {
		spec.Txt = txt
	}

	if header, err := args.GetRequiredArrayAllowEmpty("header", semantic.String); err != nil {
		spec.Header = header
	}

	return spec, nil
}

func (s *TextReaderOpSpec) Kind() flux.OperationKind {
	return TextReaderKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	Txt    string
	Header values.Array
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TextReaderOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Txt: spec.Txt,
		Header: spec.Header,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return TextReaderKind
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
		txt: spec.Txt,
		header: spec.Header,
	}, nil
}

type tableSource struct {
	execute.ExecutionNode
	id   execute.DatasetID
	mem  *memory.Allocator
	txt string
	header values.Array
	ts   execute.TransformationSet
}

func (s *tableSource) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *tableSource) Run(ctx context.Context) {
	tbl, err := buildTable(s.txt, s.header, s.mem)
	if err == nil {
		err = s.ts.Process(s.id, tbl)
	}

	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}

func buildTable(txt string, header values.Array, mem *memory.Allocator) (flux.Table, error) {

	l := header.Len()
	if l
	cols := make([]flux.ColMeta, 0, l)
	for i := 0; i < l; i++ {
		rp, err := typ.RecordProperty(i)
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
