package universe

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/v7/arrow/bitutil"
	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
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

func (s *FilterProcedureSpec) PassThroughAttribute(attrKey string) bool {
	switch attrKey {
	case plan.ParallelRunKey, plan.CollationKey:
		return true
	}
	return false
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

func NewFilterTransformation(ctx context.Context, spec *FilterProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	t := &filterTransformation{
		ctx:             ctx,
		fn:              fn,
		keepEmptyTables: spec.KeepEmptyTables,
	}
	return execute.NewNarrowTransformation(id, t, alloc)
}

type filterTransformation struct {
	ctx             context.Context
	fn              *execute.RowPredicateFn
	keepEmptyTables bool
}

func (t *filterTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem arrowmem.Allocator) error {
	// Prepare the function for the column types.
	cols := chunk.Cols()
	fn, err := t.fn.Prepare(t.ctx, cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// Prefill the columns that can be inferred from the group key.
	// Retrieve the input type from the function and record the indices
	// that need to be obtained from the columns.
	record := values.NewObject(fn.InputType())
	indices := make([]int, 0, len(chunk.Cols())-len(chunk.Key().Cols()))
	for j, c := range chunk.Cols() {
		if idx := execute.ColIdx(c.Label, chunk.Key().Cols()); idx >= 0 {
			record.Set(c.Label, chunk.Key().Value(idx))
			continue
		}
		indices = append(indices, j)
	}

	// Filter the table and pass in the indices we have to read.
	out, ok, err := t.filterChunk(fn, chunk, record, indices, mem)
	if err != nil || !ok {
		return err
	}
	return d.Process(out)
}

func (t *filterTransformation) filterChunk(fn *execute.RowPredicatePreparedFn, chunk table.Chunk, record values.Object, indices []int, mem arrowmem.Allocator) (table.Chunk, bool, error) {
	buffer := chunk.Buffer()
	bitset, err := t.filter(fn, &buffer, record, indices, mem)
	if err != nil {
		return table.Chunk{}, false, err
	}
	defer bitset.Release()

	n := bitutil.CountSetBits(bitset.Buf(), 0, bitset.Len())
	if n == 0 && !t.keepEmptyTables {
		// Drop this chunk if it is empty and we are not keeping empty tables.
		return table.Chunk{}, false, nil
	}

	// Produce arrays for each column.
	vs := make([]array.Array, len(chunk.Cols()))
	for j, col := range chunk.Cols() {
		arr := chunk.Values(j)
		if chunk.Key().HasCol(col.Label) {
			vs[j] = arrow.Slice(arr, 0, int64(n))
			continue
		}
		vs[j] = arrowutil.Filter(arr, bitset.Bytes(), mem)
	}

	return table.ChunkFromBuffer(arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   vs,
	}), true, nil
}

func (t *filterTransformation) filter(fn *execute.RowPredicatePreparedFn, cr flux.ColReader, record values.Object, indices []int, mem arrowmem.Allocator) (*arrowmem.Buffer, error) {
	cols, l := cr.Cols(), cr.Len()

	bitset := arrowmem.NewResizableBuffer(mem)
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

func (t *filterTransformation) Close() error { return nil }

// RemoveTrivialFilterRule removes Filter nodes whose predicate always evaluates to true.
type RemoveTrivialFilterRule struct{}

func (RemoveTrivialFilterRule) Name() string {
	return "RemoveTrivialFilterRule"
}

func (RemoveTrivialFilterRule) Pattern() plan.Pattern {

	return plan.MultiSuccessor(FilterKind, plan.AnySingleSuccessor())
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
	return plan.MultiSuccessor(FilterKind, plan.SingleSuccessor(FilterKind))
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
	// checks if the fields of KeepEmptyTables are different and only allows merge if 1) they are the same 2) keep is the Predecessors field
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
