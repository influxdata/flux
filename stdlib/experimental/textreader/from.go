package textreader

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
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
	Txt     string
	File    string
	ParseFn interpreter.ResolvedFunction `json:"parseFn"`
}

func init() {
	textReaderFromSignature := runtime.MustLookupBuiltinType("experimental/textreader", "from")
	runtime.RegisterPackageValue("experimental/textreader", "from", flux.MustValue(flux.FunctionValue(TextReaderKind, createFromOpSpec, textReaderFromSignature)))
	plan.RegisterProcedureSpec(TextReaderKind, newFromProcedure, TextReaderKind)
	execute.RegisterSource(TextReaderKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(TextReaderOpSpec)

	if csv, ok, err := args.GetString("txt"); err != nil {
		return nil, err
	} else if ok {
		spec.Txt = csv
	}

	if file, ok, err := args.GetString("file"); err != nil {
		return nil, err
	} else if ok {
		spec.File = file
	}

	if f, ok, err := args.GetFunction("parseFn"); err != nil {
		return nil, err
	} else if ok {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.ParseFn = fn
	} else {
		spec.ParseFn = interpreter.ResolvedFunction{
			Fn:    nil,
			Scope: nil,
		}
	}

	return spec, nil
}

func (s *TextReaderOpSpec) Kind() flux.OperationKind {
	return TextReaderKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	Txt     string
	File    string
	ParseFn interpreter.ResolvedFunction
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TextReaderOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Txt:     spec.Txt,
		File:    spec.File,
		ParseFn: spec.ParseFn,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return TextReaderKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	*ns = *s
	ns.ParseFn = s.ParseFn.Copy()
	return ns
}

func createFromSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*FromProcedureSpec)

	var getDataStream func() (io.ReadCloser, error)
	if spec.File != "" {
		getDataStream = func() (io.ReadCloser, error) {
			f, err := filesystem.OpenFile(a.Context(), spec.File)
			if err != nil {
				return nil, errors.Wrap(err, codes.Inherit, "csv.from() failed to read file")
			}
			return f, nil
		}
	} else { // if spec.File is empty then spec.CSV is not empty
		getDataStream = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(strings.NewReader(spec.Txt)), nil
		}
	}

	return &tableSource{
		id:            id,
		mem:           a.Allocator(),
		getDataStream: getDataStream,
		parseFn:       spec.ParseFn,
	}, nil
}

type tableSource struct {
	execute.ExecutionNode
	id            execute.DatasetID
	mem           *memory.Allocator
	getDataStream func() (io.ReadCloser, error)
	parseFn       interpreter.ResolvedFunction
	ts            execute.TransformationSet
}

func (s *tableSource) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *tableSource) Run(ctx context.Context) {
	tbl, err := buildTable(ctx, s, s.mem)
	if err == nil {
		err = s.ts.Process(s.id, tbl)
	}

	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}

func buildTable(ctx context.Context, s *tableSource, mem *memory.Allocator) (flux.Table, error) {

	params := semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("line"), Value: semantic.BasicString},
	})
	preparedFn, err := compiler.Compile(compiler.ToScope(s.parseFn.Scope), s.parseFn.Fn, params)
	if err != nil {
		return nil, err
	}

	input := values.NewObjectWithValues(map[string]values.Value{
		"line": values.New(""),
	})

	r, err := s.getDataStream()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(r)
	key := execute.NewGroupKey(nil, nil)
	builder := table.NewArrowBuilder(key, mem)
	for i := 0; scanner.Scan(); i++ {
		input.Set("line", values.NewString(scanner.Text()))
		row, err := preparedFn.Eval(ctx, input)
		if err != nil {
			return nil, err
		}

		typ := row.Type()
		if typ.Nature() != semantic.Object {
			return nil, errors.New(codes.Internal, "parseFn should return a record type")
		}
		if i > 0 {
			l, err := typ.NumProperties()
			if err != nil {
				return nil, err
			}
			for i := 0; i < l; i++ {
				rp, ctyp, err := getColumnType(typ, i)
				if err != nil {
					return nil, err
				}
				found := execute.ColIdx(rp.Name(), builder.Cols())
				if found < 0 || builder.Cols()[found].Type != ctyp {
					return nil, errors.New(codes.Internal, "current row on line ", i, " is different from first row.")
				}
			}
		} else {
			l, err := typ.NumProperties()
			if err != nil {
				return nil, err
			}
			cols := make([]flux.ColMeta, 0, l)
			for i := 0; i < l; i++ {
				rp, ctyp, err := getColumnType(typ, i)
				if err != nil {
					return nil, err
				}
				cols = append(cols, flux.ColMeta{
					Label: rp.Name(),
					Type:  ctyp,
				})
			}
			for _, col := range cols {
				_, err := builder.AddCol(col)
				if err != nil {
					return nil, err
				}
			}
		}

		for j, col := range builder.Cols() {
			v, _ := row.Object().Get(col.Label)
			err = arrow.AppendValue(builder.Builders[j], v)
		}
	}

	return builder.Table()
}

func getColumnType(typ semantic.MonoType, i int) (*semantic.RecordProperty, flux.ColType, error) {
	rp, err := typ.RecordProperty(i)
	if err != nil {
		return nil, flux.TInvalid, err
	}

	pt, err := rp.TypeOf()
	if err != nil {
		return nil, flux.TInvalid, err
	}
	ctyp := flux.ColumnType(pt)
	if ctyp == flux.TInvalid {
		return nil, flux.TInvalid, errors.Newf(codes.Invalid, "cannot represent the type %v as column data", pt)
	}
	return rp, ctyp, err
}
