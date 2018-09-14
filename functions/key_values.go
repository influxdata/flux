package functions

import (
	"errors"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const KeyValuesKind = "keyValues"

type KeyValuesOpSpec struct {
	KeyCols     []string                     `json:"keyCols"`
	PredicateFn *semantic.FunctionExpression `json:"fn"`
}

var keyValuesSignature = flux.DefaultFunctionSignature()

func init() {
	keyValuesSignature.Params["keyCols"] = semantic.NewArrayType(semantic.String)
	keyValuesSignature.Params["fn"] = semantic.Function

	flux.RegisterFunction(KeyValuesKind, createKeyValuesOpSpec, keyValuesSignature)
	flux.RegisterOpSpec(KeyValuesKind, newKeyValuesOp)
	plan.RegisterProcedureSpec(KeyValuesKind, newKeyValuesProcedure, KeyValuesKind)
	execute.RegisterTransformation(KeyValuesKind, createKeyValuesTransformation)
}

func createKeyValuesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KeyValuesOpSpec)

	if c, ok, err := args.GetArray("keyCols", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.KeyCols, err = interpreter.ToStringArray(c)
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

	if spec.KeyCols == nil && spec.PredicateFn == nil {
		return nil, errors.New("neither column list nor predicate function provided")
	}

	if spec.KeyCols != nil && spec.PredicateFn != nil {
		return nil, errors.New("must provide exactly one of keyCol list or predicate function")
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
	KeyCols      []string                     `json:"keyCols"`
	Predicate *semantic.FunctionExpression    `json:"fn"`
}

func newKeyValuesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KeyValuesOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &KeyValuesProcedureSpec{
		KeyCols: spec.KeyCols,
		Predicate: spec.PredicateFn,
	}, nil
}

func (s *KeyValuesProcedureSpec) Kind() plan.ProcedureKind {
	return KeyValuesKind
}

func (s *KeyValuesProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KeyValuesProcedureSpec)
	ns.KeyCols = make([]string, len(s.KeyCols))
	copy(ns.KeyCols, s.KeyCols)
	ns.Predicate = s.Predicate.Copy().(*semantic.FunctionExpression)
	return ns
}

type keyValuesTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec *KeyValuesProcedureSpec
}
func createKeyValuesTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KeyValuesProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewKeyValuesTransformation(d, cache, s)
	return t, d, nil
}

func NewKeyValuesTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KeyValuesProcedureSpec) *keyValuesTransformation {
	return &keyValuesTransformation{
		d:      d,
		cache:  cache,
		spec: spec,
	}
}

func (t *keyValuesTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *keyValuesTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("distinct found duplicate table with key: %v", tbl.Key())
	}

	// TODO: use fn to populate t.spec.keyCols


	// we'll ignore keyCol values that just don't exist in the table.
	cols := tbl.Cols()
	i := 0
	keyColIndex := -1
	for keyColIndex < 0 && i < len(t.spec.KeyCols) {
		keyColIndex = execute.ColIdx(t.spec.KeyCols[i], cols)
		i++
	}
	if keyColIndex < 1 {
		return errors.New("no columns matched by keyCols parameter")
	}


	keyColIndices :=  make([]int, len(t.spec.KeyCols))
	keyColIndices[i-1] = keyColIndex
	keyColType := cols[keyColIndex].Type
	for j, v := range t.spec.KeyCols[i:] {
		keyColIndex = execute.ColIdx(v, cols)
		keyColIndices[i+j] = keyColIndex
		if keyColIndex < 0 {
			continue
		}
		if cols[keyColIndex].Type != keyColType {
			return errors.New("keyCols must all be the same type")
		}
	}



	execute.AddTableKeyCols(tbl.Key(), builder)
	keyColIdx := builder.AddCol(flux.ColMeta{
		Label: "_key",
		Type: flux.TString,
	})
	valueColIdx := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  keyColType,
	})


	var (
		boolDistinct   map[bool]bool
		intDistinct    map[int64]bool
		uintDistinct   map[uint64]bool
		floatDistinct  map[float64]bool
		stringDistinct map[string]bool
		timeDistinct   map[execute.Time]bool
	)
	switch keyColType {
	case flux.TBool:
		boolDistinct = make(map[bool]bool)
	case flux.TInt:
		intDistinct = make(map[int64]bool)
	case flux.TUInt:
		uintDistinct = make(map[uint64]bool)
	case flux.TFloat:
		floatDistinct = make(map[float64]bool)
	case flux.TString:
		stringDistinct = make(map[string]bool)
	case flux.TTime:
		timeDistinct = make(map[execute.Time]bool)
	}



	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			// Check distinct
			for j, rowIdx := range keyColIndices {
				if rowIdx < 0 {
					continue
				}
				switch keyColType {
				case flux.TBool:
					v := cr.Bools(rowIdx)[i]
					if boolDistinct[v] {
						continue
					}
					boolDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendBool(valueColIdx, v)
				case flux.TInt:
					v := cr.Ints(rowIdx)[i]
					if intDistinct[v] {
						continue
					}
					intDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendInt(valueColIdx, v)
				case flux.TUInt:
					v := cr.UInts(rowIdx)[i]
					if uintDistinct[v] {
						continue
					}
					uintDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendUInt(valueColIdx, v)
				case flux.TFloat:
					v := cr.Floats(rowIdx)[i]
					if floatDistinct[v] {
						continue
					}
					floatDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendFloat(valueColIdx, v)
				case flux.TString:
					v := cr.Strings(rowIdx)[i]
					if stringDistinct[v] {
						continue
					}
					stringDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendString(valueColIdx, v)
				case flux.TTime:
					v := cr.Times(rowIdx)[i]
					if timeDistinct[v] {
						continue
					}
					timeDistinct[v] = true
					builder.AppendString(keyColIdx, t.spec.KeyCols[j])
					builder.AppendTime(valueColIdx, v)
				}
				execute.AppendKeyValues(tbl.Key(), builder)
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
