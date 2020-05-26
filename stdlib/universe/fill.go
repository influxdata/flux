package universe

import (
	"context"
	"strconv"

	"github.com/apache/arrow/go/arrow/array"
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

const FillKind = "fill"

type FillOpSpec struct {
	Column      string `json:"column"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	UsePrevious bool   `json:"use_previous"`
}

func init() {
	fillSignature := runtime.MustLookupBuiltinType("universe", "fill")

	runtime.RegisterPackageValue("universe", FillKind, flux.MustValue(flux.FunctionValue(FillKind, createFillOpSpec, fillSignature)))
	flux.RegisterOpSpec(FillKind, newFillOp)
	plan.RegisterProcedureSpec(FillKind, newFillProcedure, FillKind)
	execute.RegisterTransformation(FillKind, createFillTransformation)
}

func createFillOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(FillOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}

	val, valOk := args.Get("value")
	if valOk {
		typ := val.Type()
		spec.Type = typ.Nature().String()
		switch typ.Nature() {
		case semantic.Bool:
			spec.Value = strconv.FormatBool(val.Bool())
		case semantic.Int:
			spec.Value = strconv.FormatInt(val.Int(), 10)
		case semantic.UInt:
			spec.Value = strconv.FormatUint(val.UInt(), 10)
		case semantic.Float:
			spec.Value = strconv.FormatFloat(val.Float(), 'f', -1, 64)
		case semantic.String:
			spec.Value = val.Str()
		case semantic.Time:
			spec.Value = val.Time().String()
		default:
			return nil, errors.New(codes.Invalid, "value type for fill must be a valid primitive type (bool, int, uint, float, string, time)")
		}

	}

	usePrevious, prevOk, err := args.GetBool("usePrevious")
	if err != nil {
		return nil, err
	}
	if prevOk == valOk {
		return nil, errors.New(codes.Invalid, "fill requires exactly one of value or usePrevious")
	}

	if prevOk {
		spec.UsePrevious = usePrevious
	}

	return spec, nil
}

func newFillOp() flux.OperationSpec {
	return new(FillOpSpec)
}

func (s *FillOpSpec) Kind() flux.OperationKind {
	return FillKind
}

type FillProcedureSpec struct {
	plan.DefaultCost
	Column      string
	Value       values.Value
	UsePrevious bool
}

func newFillProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FillOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	pspec := &FillProcedureSpec{
		Column:      spec.Column,
		UsePrevious: spec.UsePrevious,
	}
	if !spec.UsePrevious {
		switch spec.Type {
		case "bool":
			v, err := strconv.ParseBool(spec.Value)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "int":
			v, err := strconv.ParseInt(spec.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "uint":
			v, err := strconv.ParseUint(spec.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "float":
			v, err := strconv.ParseFloat(spec.Value, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "string":
			pspec.Value = values.New(spec.Value)
		case "time":
			v, err := values.ParseTime(spec.Value)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		default:
			return nil, errors.New(codes.Internal, "unknown type in fill op-spec")
		}
	}

	return pspec, nil
}

func (s *FillProcedureSpec) Kind() plan.ProcedureKind {
	return FillKind
}
func (s *FillProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FillProcedureSpec)

	*ns = *s

	return ns
}

func createFillTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FillProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d := NewFillTransformation(a.Context(), s, id, a.Allocator())
	return t, d, nil
}

type fillTransformation struct {
	d     *execute.PassthroughDataset
	ctx   context.Context
	spec  *FillProcedureSpec
	alloc *memory.Allocator
}

func NewFillTransformation(ctx context.Context, spec *FillProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (*fillTransformation, execute.Dataset) {
	t := &fillTransformation{
		d:     execute.NewPassthroughDataset(id),
		ctx:   ctx,
		spec:  spec,
		alloc: alloc,
	}
	return t, t.d
}

func (t *fillTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *fillTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	idx := execute.ColIdx(t.spec.Column, tbl.Cols())
	if idx < 0 {
		return errors.Newf(codes.FailedPrecondition, "fill column not found: %s", t.spec.Column)
	}

	key := tbl.Key()
	if idx := execute.ColIdx(t.spec.Column, tbl.Key().Cols()); idx >= 0 {
		var err error
		gkb := execute.NewGroupKeyBuilder(tbl.Key())
		gkb.SetKeyValue(t.spec.Column, values.New(t.spec.Value))
		key, err = gkb.Build()
		if err != nil {
			return err
		}
	}

	prevNonNull := t.spec.Value
	if !t.spec.UsePrevious {
		if tbl.Cols()[idx].Type != flux.ColumnType(prevNonNull.Type()) {
			return errors.Newf(codes.FailedPrecondition, "fill column type mismatch: %s/%s", tbl.Cols()[idx].Type.String(), flux.ColumnType(prevNonNull.Type()).String())
		}
	}

	table, err := table.StreamWithContext(t.ctx, key, tbl.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return tbl.Do(func(cr flux.ColReader) error {
			l := cr.Len()
			if l == 0 {
				return nil
			}
			vs := make([]array.Interface, len(w.Cols()))
			for j, col := range w.Cols() {
				if j != idx {
					vs[j] = table.Values(cr, j)
					vs[j].Retain()
				} else {
					if t.spec.UsePrevious {
						prevNonNull = execute.ValueForRow(cr, 0, idx)
					}
					b := arrow.NewBuilder(col.Type, t.alloc)
					b.Reserve(l)
					hasNull := false
					for i := 0; i < l; i++ {
						v := execute.ValueForRow(cr, i, idx)
						if v.IsNull() {
							if err := arrow.AppendValue(b, prevNonNull); err != nil {
								return err
							}
							hasNull = true
						} else {
							if err := arrow.AppendValue(b, v); err != nil {
								return err
							}
							if t.spec.UsePrevious {
								prevNonNull = v
							}
						}
					}
					if hasNull {
						vs[j] = b.NewArray()
					} else {
						b.Release()
						vs[j] = table.Values(cr, j)
						vs[j].Retain()
					}
				}
			}
			return w.Write(vs)
		})
	})
	if err != nil {
		return err
	}
	return t.d.Process(table)
}

func (t *fillTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *fillTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *fillTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
