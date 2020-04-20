package universe

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/bitutil"
	arrowmem "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FilterKind = "filter"

type FilterOpSpec struct {
	Fn      interpreter.ResolvedFunction `json:"fn"`
	OnEmpty string                       `json:"onEmpty,omitempty"`
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
			"onEmpty": semantic.String,
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
	body := s.Fn.Fn.Block.Body
	if expr, ok := body.(semantic.Expression); ok {
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
	fn, err := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	if err != nil {
		return nil, nil, err
	}

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

	// Otherwise, we have to filter the existing table.
	table, err := t.filterTable(tbl, record, properties)
	if err != nil {
		return err
	} else if table.Empty() && !t.keepEmptyTables {
		// Drop the table.
		return nil
	}
	return t.d.Process(table)
}

func (t *filterTransformation) filterTable(in flux.Table, record values.Object, properties map[string]semantic.Type) (flux.Table, error) {
	return table.StreamWithContext(t.ctx, in.Key(), in.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return in.Do(func(cr flux.ColReader) error {
			bitset, err := t.filter(cr, record, properties)
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

func (t *filterTransformation) filter(cr flux.ColReader, record values.Object, properties map[string]semantic.Type) (*arrowmem.Buffer, error) {
	cols, l := cr.Cols(), cr.Len()
	bitset := arrowmem.NewResizableBuffer(t.alloc)
	bitset.Resize(l)
	for i := 0; i < l; i++ {
		for j := 0; j < len(cols); j++ {
			label := cols[j].Label
			if _, ok := properties[label]; ok {
				record.Set(label, execute.ValueForRow(cr, i, j))
			}
		}

		val, err := t.fn.Eval(t.ctx, record)
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
	if boolean, ok := filterSpec.Fn.Fn.Block.Body.(*semantic.BooleanLiteral); !ok || !boolean.Value {
		return filterNode, false, nil
	}

	anyNode := filterNode.Predecessors()[0]
	return anyNode, true, nil
}
