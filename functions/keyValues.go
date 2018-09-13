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

var (
	keyValuesExceptDefaultValue = []string{"_time", "_value"}
)

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

	*ns = *s

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
	return nil
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
