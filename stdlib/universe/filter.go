package universe

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FilterKind = "filter"

type FilterOpSpec struct {
	Fn interpreter.ResolvedFunction `json:"fn"`
}

func init() {
	filterSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{
					"r": semantic.Tvar(1),
				},
				Required: semantic.LabelSet{"r"},
				Return:   semantic.Bool,
			}),
		},
		[]string{"fn"},
	)

	flux.RegisterPackageValue("universe", FilterKind, flux.FunctionValue(FilterKind, createFilterOpSpec, filterSignature))
	flux.RegisterOpSpec(FilterKind, newFilterOp)
	plan.RegisterProcedureSpec(FilterKind, newFilterProcedure, FilterKind)
	execute.RegisterTransformation(FilterKind, createFilterTransformation)
	plan.RegisterPhysicalRules(
		RemoveTrivialFilterRule{},
	)
}

func createFilterOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	fn, err := interpreter.ResolveFunction(f)
	if err != nil {
		return nil, err
	}

	return &FilterOpSpec{
		Fn: fn,
	}, nil
}
func newFilterOp() flux.OperationSpec {
	return new(FilterOpSpec)
}

func (s *FilterOpSpec) Kind() flux.OperationKind {
	return FilterKind
}

type FilterProcedureSpec struct {
	plan.DefaultCost
	Fn interpreter.ResolvedFunction
}

func newFilterProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FilterOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FilterProcedureSpec{
		Fn: spec.Fn,
	}, nil
}

func (s *FilterProcedureSpec) Kind() plan.ProcedureKind {
	return FilterKind
}
func (s *FilterProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FilterProcedureSpec)
	ns.Fn = s.Fn.Copy()
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *FilterProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createFilterTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FilterProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	t, d, err := NewFilterTransformation(a.Context(), s, id, a.Allocator())
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type filterTransformation struct {
	d     *execute.PassthroughDataset
	ctx   context.Context
	fn    *execute.RowPredicateFn
	alloc *memory.Allocator
}

func NewFilterTransformation(ctx context.Context, spec *FilterProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn, err := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	if err != nil {
		return nil, nil, err
	}

	t := &filterTransformation{
		d:     execute.NewPassthroughDataset(id),
		fn:    fn,
		ctx:   ctx,
		alloc: alloc,
	}
	return t, t.d, nil
}

func (t *filterTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *filterTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the function for the column types.
	cols := tbl.Cols()
	if err := t.fn.Prepare(cols); err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// Copy out the properties so we can modify the map.
	properties := make(map[string]semantic.Type)
	for name, typ := range t.fn.InputType().Properties() {
		properties[name] = typ
	}

	// Iterate through the properties and prefill a record
	// with the values from the group key.
	record := values.NewObject()
	for name := range properties {
		if idx := execute.ColIdx(name, tbl.Key().Cols()); idx >= 0 {
			record.Set(name, tbl.Key().Value(idx))
			delete(properties, name)
		}
	}

	// If there are no remaining properties, then all
	// of the referenced values were in the group key
	// and we can perform the comparison once for the
	// entire table.
	if len(properties) == 0 {
		v, err := t.fn.Eval(t.ctx, record)
		if err != nil {
			return err
		}

		if !v {
			tbl.Done()
			tbl = execute.NewEmptyTable(tbl.Key(), tbl.Cols())
		}
		return t.d.Process(tbl)
	}

	// Otherwise, we have to filter the existing table.
	// TODO(jsternberg): I'm pretty sure this can be optimized
	// too, but we're getting fewer returns on optimizing this
	// area right now. In particular, I don't think it is
	// more efficient to construct the table as we are
	// processing. It is likely more efficient to perform
	// the comparisons on each row and specify if it matches
	// or not, then use that information to determine if using
	// slices or reconstructing the table would be faster.
	builder := execute.NewColListTableBuilder(tbl.Key(), t.alloc)
	defer builder.Release()

	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			for j := 0; j < len(cols); j++ {
				label := cols[j].Label
				if _, ok := properties[label]; ok {
					record.Set(label, execute.ValueForRow(cr, i, j))
				}
			}
			if pass, err := t.fn.Eval(t.ctx, record); err != nil {
				return errors.Wrap(err, codes.Inherit, "failed to evaluate filter function")
			} else if !pass {
				// No match, skipping
				continue
			}
			if err := execute.AppendRecord(i, cr, builder); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	table, err := builder.Table()
	if err != nil {
		return err
	}
	return t.d.Process(table)
}

func (t *filterTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *filterTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *filterTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

// RemoveTrivialFilterRule removes Filter nodes whose predicate always evaluates to true.
type RemoveTrivialFilterRule struct{}

func (RemoveTrivialFilterRule) Name() string {
	return "RemoveTrivialFilterRule"
}

func (RemoveTrivialFilterRule) Pattern() plan.Pattern {
	return plan.Pat(FilterKind, plan.Any())
}

func (RemoveTrivialFilterRule) Rewrite(filterNode plan.Node) (plan.Node, bool, error) {
	filterSpec := filterNode.ProcedureSpec().(*FilterProcedureSpec)
	if filterSpec.Fn.Fn == nil ||
		filterSpec.Fn.Fn.Block == nil ||
		filterSpec.Fn.Fn.Block.Body == nil {
		return filterNode, false, nil
	}
	if boolean, ok := filterSpec.Fn.Fn.Block.Body.(*semantic.BooleanLiteral); !ok || !boolean.Value {
		return filterNode, false, nil
	}

	anyNode := filterNode.Predecessors()[0]
	return anyNode, true, nil
}
