package universe

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/bitutil"
	arrowmem "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FilterKind = "filter"

type FilterOpSpec struct {
	Fn      interpreter.ResolvedFunction `json:"fn"`
	OnEmpty string                       `json:"onEmpty,omitempty"`
}

func init() {
	filterSignature := runtime.MustLookupBuiltinType("universe", "filter")

	runtime.RegisterPackageValue("universe", FilterKind, flux.MustValue(flux.FunctionValue(FilterKind, createFilterOpSpec, filterSignature)))
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

	onEmpty, ok, err := args.GetString("onEmpty")
	if err != nil {
		return nil, err
	} else if ok {
		// Check that the string is ok.
		switch onEmpty {
		case "keep", "drop":
		default:
			return nil, errors.Newf(codes.Invalid, "onEmpty must be keep or drop, was %q", onEmpty)
		}
	}

	fn, err := interpreter.ResolveFunction(f)
	if err != nil {
		return nil, err
	}

	return &FilterOpSpec{
		Fn:      fn,
		OnEmpty: onEmpty,
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
	Fn              interpreter.ResolvedFunction
	KeepEmptyTables bool
}

func newFilterProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FilterOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	onEmpty := spec.OnEmpty
	if onEmpty == "" {
		onEmpty = "drop"
	}

	return &FilterProcedureSpec{
		Fn:              spec.Fn,
		KeepEmptyTables: onEmpty == "keep",
	}, nil
}

func (s *FilterProcedureSpec) Kind() plan.ProcedureKind {
	return FilterKind
}
func (s *FilterProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FilterProcedureSpec)
	ns.Fn = s.Fn.Copy()
	ns.KeepEmptyTables = s.KeepEmptyTables
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *FilterProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createFilterTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FilterProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d, err := NewFilterTransformation(a.Context(), s, id, a.Allocator())
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

func (s *FilterProcedureSpec) PlanDetails() string {
	if expr, ok := s.Fn.Fn.GetFunctionBodyExpression(); ok {
		return fmt.Sprintf("%v", semantic.Formatted(expr))
	}
	return "<non-Expression>"
}

type filterTransformation struct {
	d               *execute.PassthroughDataset
	ctx             context.Context
	fn              *execute.RowPredicateFn
	keepEmptyTables bool
	alloc           *memory.Allocator
}

func NewFilterTransformation(ctx context.Context, spec *FilterProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	t := &filterTransformation{
		d:               execute.NewPassthroughDataset(id),
		fn:              fn,
		ctx:             ctx,
		keepEmptyTables: spec.KeepEmptyTables,
		alloc:           alloc,
	}
	return t, t.d, nil
}

func (t *filterTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *filterTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the function for the column types.
	cols := tbl.Cols()
	fn, err := t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// Retrieve the inferred input type for the function.
	// If all of the inferred inputs are part of the group
	// key, we can evaluate a record with only the group key.
	if t.canFilterByKey(fn, tbl) {
		return t.filterByKey(tbl)
	}

	// Prefill the columns that can be inferred from the group key.
	// Retrieve the input type from the function and record the indices
	// that need to be obtained from the columns.
	record := values.NewObject(fn.InputType())
	indices := make([]int, 0, len(tbl.Cols())-len(tbl.Key().Cols()))
	for j, c := range tbl.Cols() {
		if idx := execute.ColIdx(c.Label, tbl.Key().Cols()); idx >= 0 {
			record.Set(c.Label, tbl.Key().Value(idx))
			continue
		}
		indices = append(indices, j)
	}

	// Filter the table and pass in the indices we have to read.
	table, err := t.filterTable(fn, tbl, record, indices)
	if err != nil {
		return err
	} else if table.Empty() && !t.keepEmptyTables {
		// Drop the table.
		return nil
	}
	return t.d.Process(table)
}

func (t *filterTransformation) canFilterByKey(fn *execute.RowPredicatePreparedFn, tbl flux.Table) bool {
	inType := fn.InferredInputType()
	nargs, err := inType.NumProperties()
	if err != nil {
		panic(err)
	}

	for i := 0; i < nargs; i++ {
		prop, err := inType.RowProperty(i)
		if err != nil {
			panic(err)
		}

		// Determine if this key is even valid. If it is not
		// in the table at all, we don't care if it is missing
		// since it will always be missing.
		label := prop.Name()
		if execute.ColIdx(label, tbl.Cols()) < 0 {
			continue
		}

		// Look for a column with this name in the group key.
		if execute.ColIdx(label, tbl.Key().Cols()) < 0 {
			// If we cannot find this referenced column in the group
			// key, then it is provided by the table and we need to
			// evaluate each row individually.
			return false
		}
	}

	// All referenced keys were part of the group key.
	return true
}

func (t *filterTransformation) filterByKey(tbl flux.Table) error {
	key := tbl.Key()
	cols := key.Cols()
	fn, err := t.fn.Prepare(cols)
	if err != nil {
		return err
	}

	record, err := values.BuildObjectWithSize(len(cols), func(set values.ObjectSetter) error {
		for j, c := range cols {
			set(c.Label, key.Value(j))
		}
		return nil
	})
	if err != nil {
		return err
	}

	v, err := fn.Eval(t.ctx, record)
	if err != nil {
		return err
	}

	if !v {
		tbl.Done()
		if !t.keepEmptyTables {
			return nil
		}
		// If we are supposed to keep empty tables, produce
		// an empty table with this group key and send it
		// to the next transformation to process it.
		tbl = execute.NewEmptyTable(tbl.Key(), tbl.Cols())
	}
	return t.d.Process(tbl)
}

func (t *filterTransformation) filterTable(fn *execute.RowPredicatePreparedFn, in flux.Table, record values.Object, indices []int) (flux.Table, error) {
	return table.StreamWithContext(t.ctx, in.Key(), in.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return in.Do(func(cr flux.ColReader) error {
			bitset, err := t.filter(fn, cr, record, indices)
			if err != nil {
				return err
			}
			defer bitset.Release()

			n := bitutil.CountSetBits(bitset.Buf(), 0, bitset.Len())
			if n == 0 {
				return nil
			}

			// Produce arrays for each column.
			vs := make([]array.Interface, len(w.Cols()))
			for j, col := range w.Cols() {
				arr := table.Values(cr, j)
				if in.Key().HasCol(col.Label) {
					vs[j] = arrow.Slice(arr, 0, int64(n))
					continue
				}
				vs[j] = arrowutil.Filter(arr, bitset.Bytes(), t.alloc)
			}
			return w.Write(vs)
		})
	})
}

func (t *filterTransformation) filter(fn *execute.RowPredicatePreparedFn, cr flux.ColReader, record values.Object, indices []int) (*arrowmem.Buffer, error) {
	cols, l := cr.Cols(), cr.Len()
	bitset := arrowmem.NewResizableBuffer(t.alloc)
	bitset.Resize(l)
	for i := 0; i < l; i++ {
		for _, j := range indices {
			record.Set(cols[j].Label, execute.ValueForRow(cr, i, j))
		}

		val, err := fn.Eval(t.ctx, record)
		if err != nil {
			bitset.Release()
			return nil, errors.Wrap(err, codes.Inherit, "failed to evaluate filter function")
		}
		bitutil.SetBitTo(bitset.Buf(), i, val)
	}
	return bitset, nil
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

func (RemoveTrivialFilterRule) Rewrite(ctx context.Context, filterNode plan.Node) (plan.Node, bool, error) {
	filterSpec := filterNode.ProcedureSpec().(*FilterProcedureSpec)
	if filterSpec.Fn.Fn == nil ||
		filterSpec.Fn.Fn.Block == nil ||
		filterSpec.Fn.Fn.Block.Body == nil {
		return filterNode, false, nil
	}

	if bodyExpr, ok := filterSpec.Fn.Fn.GetFunctionBodyExpression(); !ok {
		// Not an expression.
		return filterNode, false, nil
	} else if expr, ok := bodyExpr.(*semantic.BooleanLiteral); !ok || !expr.Value {
		// Either not a boolean at all, or evaluates to false.
		return filterNode, false, nil
	}

	anyNode := filterNode.Predecessors()[0]
	return anyNode, true, nil
}

// MergeFiltersRule merges Filter nodes whose body is a single return to create one Filter node.
type MergeFiltersRule struct{}

func (MergeFiltersRule) Name() string {
	return "MergeFiltersRule"
}

func (MergeFiltersRule) Pattern() plan.Pattern {
	return plan.Pat(FilterKind, plan.Pat(FilterKind, plan.Any()))
}

func (MergeFiltersRule) Rewrite(ctx context.Context, filterNode plan.Node) (plan.Node, bool, error) {
	// conditions
	filterSpec1 := filterNode.ProcedureSpec().(*FilterProcedureSpec)
	bodyExpr1, ok := filterSpec1.Fn.Fn.GetFunctionBodyExpression()
	if !ok {
		// Not an expression.
		return filterNode, false, nil
	}
	filterSpec2 := filterNode.Predecessors()[0].ProcedureSpec().(*FilterProcedureSpec)
	bodyExpr2, ok := filterSpec2.Fn.Fn.GetFunctionBodyExpression()
	if !ok {
		// Not an expression.
		return filterNode, false, nil
	}
	//checks if the fields of KeepEmptyTables are different and only allows merge if 1) they are the same 2) keep is the Predecessors field
	if filterSpec1.KeepEmptyTables != filterSpec2.KeepEmptyTables && !filterSpec2.KeepEmptyTables {
		return filterNode, false, nil
	}

	// created an instance of LogicalExpression to 'and' two different arguments
	expr := &semantic.LogicalExpression{Left: bodyExpr1, Operator: ast.AndOperator, Right: bodyExpr2}
	// set a new variables that converted the single body statement to a return type that can used with expr
	ret := filterSpec2.Fn.Fn.Block.Body[0].(*semantic.ReturnStatement)
	ret.Argument = expr
	// return the pred node
	anyNode := filterNode.Predecessors()[0]
	return anyNode, true, nil
}
