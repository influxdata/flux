package transformations

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const KeysKind = "keys"

var (
	keysExceptDefaultValue = []string{"_time", "_value"}
)

type KeysOpSpec struct {
	Except []string `json:"except"`
}

var keysSignature = flux.DefaultFunctionSignature()

func init() {
	keysSignature.Params["except"] = semantic.NewArrayType(semantic.String)

	flux.RegisterFunction(KeysKind, createKeysOpSpec, keysSignature)
	flux.RegisterOpSpec(KeysKind, newKeysOp)
	plan.RegisterProcedureSpec(KeysKind, newKeysProcedure, KeysKind)
	plan.RegisterRewriteRule(KeysPointLimitRewriteRule{})
	execute.RegisterTransformation(KeysKind, createKeysTransformation)
}

func createKeysOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KeysOpSpec)
	if array, ok, err := args.GetArray("except", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.Except, err = interpreter.ToStringArray(array)
		if err != nil {
			return nil, err
		}
	} else {
		spec.Except = keysExceptDefaultValue
	}

	return spec, nil
}

func newKeysOp() flux.OperationSpec {
	return new(KeysOpSpec)
}

func (s *KeysOpSpec) Kind() flux.OperationKind {
	return KeysKind
}

type KeysProcedureSpec struct {
	Except []string
}

func newKeysProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KeysOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &KeysProcedureSpec{
		Except: spec.Except,
	}, nil
}

func (s *KeysProcedureSpec) Kind() plan.ProcedureKind {
	return KeysKind
}

func (s *KeysProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KeysProcedureSpec)

	*ns = *s

	return ns
}

type KeysPointLimitRewriteRule struct {
}

func (r KeysPointLimitRewriteRule) Root() plan.ProcedureKind {
	return inputs.FromKind
}

func (r KeysPointLimitRewriteRule) Rewrite(pr *plan.Procedure, planner plan.PlanRewriter) error {
	fromSpec, ok := pr.Spec.(*inputs.FromProcedureSpec)
	if !ok {
		return nil
	}

	var keys *KeysProcedureSpec
	pr.DoChildren(func(child *plan.Procedure) {
		if d, ok := child.Spec.(*KeysProcedureSpec); ok {
			keys = d
		}
	})
	if keys == nil {
		return nil
	}

	if !fromSpec.LimitSet {
		fromSpec.LimitSet = true
		fromSpec.PointsLimit = -1
	}
	return nil
}

func createKeysTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KeysProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewKeysTransformation(d, cache, s)
	return t, d, nil
}

type keysTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	except []string
}

func NewKeysTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KeysProcedureSpec) *keysTransformation {
	var except []string
	if len(spec.Except) > 0 {
		except = append([]string{}, spec.Except...)
		sort.Strings(except)
	}

	return &keysTransformation{
		d:      d,
		cache:  cache,
		except: except,
	}
}

func (t *keysTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *keysTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("keys found duplicate table with key: %v", tbl.Key())
	}

	err := execute.AddTableKeyCols(tbl.Key(), builder)
	if err != nil {
		return err
	}
	colIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TString})
	if err != nil {
		return err
	}

	cols := tbl.Cols()
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Label < cols[j].Label
	})

	var i int
	if len(t.except) > 0 {
		var j int
		for i < len(cols) && j < len(t.except) {
			c := strings.Compare(cols[i].Label, t.except[j])
			if c < 0 {
				execute.AppendKeyValues(tbl.Key(), builder)
				builder.AppendString(colIdx, cols[i].Label)
				i++
			} else if c > 0 {
				j++
			} else {
				i++
				j++
			}
		}
	}

	// add remaining
	for ; i < len(cols); i++ {
		execute.AppendKeyValues(tbl.Key(), builder)
		builder.AppendString(colIdx, cols[i].Label)
	}

	// TODO: this is a hack
	return tbl.Do(func(flux.ColReader) error {
		return nil
	})
}

func (t *keysTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *keysTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *keysTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
