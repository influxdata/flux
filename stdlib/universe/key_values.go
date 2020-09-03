package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const KeyValuesKind = "keyValues"

type KeyValuesOpSpec struct {
	KeyColumns  []string                     `json:"keyColumns"`
	PredicateFn interpreter.ResolvedFunction `json:"fn"`
}

func init() {
	keyValuesSignature := runtime.MustLookupBuiltinType("universe", "keyValues")

	runtime.RegisterPackageValue("universe", KeyValuesKind, flux.MustValue(flux.FunctionValue(KeyValuesKind, createKeyValuesOpSpec, keyValuesSignature)))
	flux.RegisterOpSpec(KeyValuesKind, newKeyValuesOp)
	plan.RegisterProcedureSpec(KeyValuesKind, newKeyValuesProcedure, KeyValuesKind)
	execute.RegisterTransformation(KeyValuesKind, createKeyValuesTransformation)
}

func createKeyValuesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KeyValuesOpSpec)

	if c, ok, err := args.GetArray("keyColumns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.KeyColumns, err = interpreter.ToStringArray(c)
		if err != nil {
			return nil, err
		}
	}

	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.PredicateFn = fn
	}

	if spec.KeyColumns == nil && spec.PredicateFn.Fn == nil {
		return nil, errors.New(codes.Invalid, "neither column list nor predicate function provided")
	}

	if spec.KeyColumns != nil && spec.PredicateFn.Fn != nil {
		return nil, errors.New(codes.Invalid, "must provide exactly one of keyColumns list or predicate function")
	}

	return spec, nil
}

func newKeyValuesOp() flux.OperationSpec {
	return new(KeyValuesOpSpec)
}

func (s *KeyValuesOpSpec) Kind() flux.OperationKind {
	return KeyValuesKind
}

type KeyValuesProcedureSpec struct {
	plan.DefaultCost
	KeyColumns []string                     `json:"keyColumns"`
	Predicate  interpreter.ResolvedFunction `json:"fn"`
}

func newKeyValuesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KeyValuesOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &KeyValuesProcedureSpec{
		KeyColumns: spec.KeyColumns,
		Predicate:  spec.PredicateFn,
	}, nil
}

func (s *KeyValuesProcedureSpec) Kind() plan.ProcedureKind {
	return KeyValuesKind
}

func (s *KeyValuesProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KeyValuesProcedureSpec)
	ns.KeyColumns = make([]string, len(s.KeyColumns))
	copy(ns.KeyColumns, s.KeyColumns)
	ns.Predicate = s.Predicate.Copy()
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *KeyValuesProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type keyValuesTransformation struct {
	execute.ExecutionNode
	d        execute.Dataset
	cache    execute.TableBuilderCache
	spec     *KeyValuesProcedureSpec
	distinct bool
}

func createKeyValuesTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KeyValuesProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewKeyValuesTransformation(d, cache, s)
	return t, d, nil
}

func NewKeyValuesTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KeyValuesProcedureSpec) *keyValuesTransformation {
	return &keyValuesTransformation{
		d:        d,
		cache:    cache,
		spec:     spec,
		distinct: true,
	}
}

func (t *keyValuesTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *keyValuesTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.Internal, "distinct found duplicate table with key: %v", tbl.Key())
	}

	// TODO: use fn to populate t.spec.keyColumns

	var matchingCols bool
	var keyColType flux.ColType

	keyColumns := make([]struct {
		name string
		typ  flux.ColType
		idx  int
	}, 0, len(t.spec.KeyColumns))

	for _, c := range t.spec.KeyColumns {
		idx := execute.ColIdx(c, tbl.Cols())
		if idx >= 0 {
			matchingCols = true
			keyColType = tbl.Cols()[idx].Type
			keyColumns = append(keyColumns, struct {
				name string
				typ  flux.ColType
				idx  int
			}{
				name: c,
				typ:  tbl.Cols()[idx].Type,
				idx:  idx,
			})
		}
	}

	if !matchingCols {
		columnNames := make([]string, len(tbl.Cols()))
		for i, column := range tbl.Cols() {
			columnNames[i] = column.Label
		}
		return errors.Newf(codes.FailedPrecondition, "received table with columns %v not having key columns %v", columnNames, t.spec.KeyColumns)
	}

	for _, c := range keyColumns {
		if c.typ != keyColType {
			return errors.New(codes.FailedPrecondition, "keyColumns must all be the same type")
		}
	}

	err := execute.AddTableKeyCols(tbl.Key(), builder)
	if err != nil {
		return err
	}
	keyColIdx, err := builder.AddCol(flux.ColMeta{
		Label: "_key",
		Type:  flux.TString,
	})
	if err != nil {
		return err
	}
	valueColIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  keyColType,
	})
	if err != nil {
		return err
	}

	var (
		boolDistinct = map[struct {
			string
			bool
		}]bool{
			{"", false}: false,
		}
		intDistinct = map[struct {
			string
			int64
		}]bool{
			{"", 0}: false,
		}
		uintDistinct = map[struct {
			string
			uint64
		}]bool{
			{"", 0}: false,
		}
		floatDistinct = map[struct {
			string
			float64
		}]bool{
			{"", 0}: false,
		}
		timeDistinct = map[struct {
			string
			execute.Time
		}]bool{
			{"", 0}: false,
		}
		stringDistinct = map[[2]string]bool{
			{"", ""}: false,
		}
		nullDistinct = false
	)

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			// Check distinct
			for _, c := range keyColumns {
				switch keyColType {
				case flux.TBool:
					vs := cr.Bools(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := vs.Value(i)
							if boolDistinct[struct {
								string
								bool
							}{c.name, v}] {
								continue
							}
							boolDistinct[struct {
								string
								bool
							}{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}

					if vs.IsValid(i) {
						v := vs.Value(i)
						if err := builder.AppendBool(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				case flux.TInt:
					vs := cr.Ints(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := vs.Value(i)
							if intDistinct[struct {
								string
								int64
							}{c.name, v}] {
								continue
							}
							intDistinct[struct {
								string
								int64
							}{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}

					if vs.IsValid(i) {
						v := vs.Value(i)
						if err := builder.AppendInt(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				case flux.TUInt:
					vs := cr.UInts(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := vs.Value(i)
							if uintDistinct[struct {
								string
								uint64
							}{c.name, v}] {
								continue
							}
							uintDistinct[struct {
								string
								uint64
							}{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}

					if vs.IsValid(i) {
						v := vs.Value(i)
						if err := builder.AppendUInt(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				case flux.TFloat:
					vs := cr.Floats(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := vs.Value(i)
							if floatDistinct[struct {
								string
								float64
							}{c.name, v}] {
								continue
							}
							floatDistinct[struct {
								string
								float64
							}{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}
					if vs.IsValid(i) {
						v := vs.Value(i)
						if err := builder.AppendFloat(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				case flux.TString:
					vs := cr.Strings(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := vs.ValueString(i)
							if stringDistinct[[2]string{c.name, v}] {
								continue
							}
							stringDistinct[[2]string{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}
					if vs.IsValid(i) {
						v := vs.ValueString(i)
						if err := builder.AppendString(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				case flux.TTime:
					vs := cr.Times(c.idx)
					if t.distinct {
						if vs.IsNull(i) {
							if nullDistinct {
								continue
							}
							nullDistinct = true
						} else {
							v := execute.Time(vs.Value(i))
							if timeDistinct[struct {
								string
								execute.Time
							}{c.name, v}] {
								continue
							}
							timeDistinct[struct {
								string
								execute.Time
							}{c.name, v}] = true
						}
					}
					if err := builder.AppendString(keyColIdx, c.name); err != nil {
						return err
					}

					if vs.IsValid(i) {
						v := execute.Time(vs.Value(i))
						if err := builder.AppendTime(valueColIdx, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(valueColIdx); err != nil {
							return err
						}
					}
				}
				if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (t *keyValuesTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *keyValuesTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *keyValuesTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
