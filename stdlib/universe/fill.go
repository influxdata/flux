package universe

import (
	"strconv"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
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
	fillSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"column":      semantic.String,
			"value":       semantic.Tvar(1),
			"usePrevious": semantic.Bool,
		},
		[]string{},
	)

	flux.RegisterPackageValue("universe", FillKind, flux.FunctionValue(FillKind, createFillOpSpec, fillSignature))
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
		switch typ {
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
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewFillTransformation(d, cache, s)
	return t, d, nil
}

type fillTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec *FillProcedureSpec
}

func NewFillTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *FillProcedureSpec) *fillTransformation {
	return &fillTransformation{
		d:     d,
		cache: cache,
		spec:  spec,
	}
}

func (t *fillTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *fillTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
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

	builder, created := t.cache.TableBuilder(key)
	if created {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return err
		}
	}
	idx := execute.ColIdx(t.spec.Column, builder.Cols())
	if idx < 0 {
		return errors.Newf(codes.FailedPrecondition, "fill column not found: %s", t.spec.Column)
	}

	prevNonNull := t.spec.Value
	if !t.spec.UsePrevious {
		if builder.Cols()[idx].Type != flux.ColumnType(prevNonNull.Type()) {
			return errors.Newf(codes.FailedPrecondition, "fill column type mismatch: %s/%s", builder.Cols()[idx].Type.String(), flux.ColumnType(prevNonNull.Type()).String())
		}
	}
	return tbl.Do(func(cr flux.ColReader) error {
		for j := range cr.Cols() {
			if j == idx {
				continue
			}
			if err := execute.AppendCol(j, j, cr, builder); err != nil {
				return err
			}
		}
		// Set new value
		l := cr.Len()

		if l > 0 {
			if t.spec.UsePrevious {
				prevNonNull = execute.ValueForRow(cr, 0, idx)
			}

			for i := 0; i < l; i++ {
				v := execute.ValueForRow(cr, i, idx)
				if v.IsNull() {
					if err := builder.AppendValue(idx, prevNonNull); err != nil {
						return err
					}
				} else {
					if err := builder.AppendValue(idx, v); err != nil {
						return err
					}
					if t.spec.UsePrevious {
						prevNonNull = v
					}

				}
			}
		}
		return nil
	})
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
