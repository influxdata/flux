package aggregate

import (
	"context"

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
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const pkgpath = "contrib/jsternberg/aggregate"

const TableKind = pkgpath + ".table"

type TableOpSpec struct {
	Columns []TableColumn
}

type TableColumn struct {
	Column  string
	Init    interpreter.ResolvedFunction
	Reduce  interpreter.ResolvedFunction
	Compute interpreter.ResolvedFunction
	As      string
}

func init() {
	runtime.RegisterPackageValue(pkgpath, "table", flux.MustValue(flux.FunctionValue(
		"table",
		createTableOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "table"),
	)))
	plan.RegisterProcedureSpec(TableKind, newTableProcedure, TableKind)
	execute.RegisterTransformation(TableKind, createTableTransformation)
}

func createTableOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(TableOpSpec)

	columns, err := args.GetRequired("columns")
	if err != nil {
		return nil, err
	}

	asTargets := make(map[string]bool)
	spec.Columns = make([]TableColumn, columns.Array().Len())
	columns.Array().Range(func(i int, v values.Value) {
		if err != nil {
			return
		}

		object := v.Object()
		column, _ := object.Get("column")
		as, _ := object.Get("as")

		ac := TableColumn{
			Column: column.Str(),
			As:     as.Str(),
		}

		if asTargets[ac.As] {
			err = errors.Newf(codes.Invalid, "multiple columns with the same output name: %s", ac.As)
			return
		}
		asTargets[ac.As] = true

		with := func() values.Object {
			o, _ := object.Get("with")
			return o.Object()
		}()

		resolveFn := func(name string, with values.Object) (interpreter.ResolvedFunction, error) {
			fn, _ := with.Get(name)
			return interpreter.ResolveFunction(fn.Function())
		}

		ac.Init, err = resolveFn("init", with)
		if err != nil {
			return
		}

		ac.Reduce, err = resolveFn("reduce", with)
		if err != nil {
			return
		}

		ac.Compute, err = resolveFn("compute", with)
		if err != nil {
			return
		}

		spec.Columns[i] = ac
	})

	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (a *TableOpSpec) Kind() flux.OperationKind {
	return TableKind
}

type TableProcedureSpec struct {
	plan.DefaultCost
	Columns []TableColumn
}

func newTableProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TableOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &TableProcedureSpec{
		Columns: spec.Columns,
	}, nil
}

func (s *TableProcedureSpec) Kind() plan.ProcedureKind {
	return TableKind
}

func (s *TableProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(TableProcedureSpec)
	ns.Columns = make([]TableColumn, len(s.Columns))
	for i, c := range s.Columns {
		ns.Columns[i] = TableColumn{
			Column:  c.Column,
			Init:    c.Init.Copy(),
			Reduce:  c.Reduce.Copy(),
			Compute: c.Compute.Copy(),
			As:      c.As,
		}
	}
	return ns
}

func createTableTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*TableProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewTableTransformation(a.Context(), s, id, a.Allocator())
}

type tableTransformation struct {
	d     *execute.PassthroughDataset
	spec  *TableProcedureSpec
	ctx   context.Context
	alloc *memory.Allocator
}

func NewTableTransformation(ctx context.Context, spec *TableProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &tableTransformation{
		d:     execute.NewPassthroughDataset(id),
		spec:  spec,
		ctx:   ctx,
		alloc: alloc,
	}
	return t, t.d, nil
}

func (t *tableTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *tableTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if err := t.validateGroupKey(tbl.Key()); err != nil {
		return err
	}

	if err := t.validateInputColumns(tbl.Key(), tbl.Cols()); err != nil {
		return err
	}

	// Initialize the state for each column.
	var aggregateState []values.Value
	if err := tbl.Do(func(cr flux.ColReader) error {
		if aggregateState == nil {
			var err error
			aggregateState, err = t.initializeAggregateState(cr)
			return err
		}
		return t.processBuffer(cr, aggregateState)
	}); err != nil {
		return err
	}

	aggregates, err := t.computeFromState(aggregateState)
	if err != nil {
		return err
	}

	outTable, err := t.buildTable(tbl.Key(), aggregates)
	if err != nil {
		return err
	}
	return t.d.Process(outTable)
}

func (t *tableTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *tableTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *tableTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

// validateGroupKey will check if any of the labels in the
// group key conflict with column names to be created.
func (t *tableTransformation) validateGroupKey(key flux.GroupKey) error {
	for _, c := range t.spec.Columns {
		if key.HasCol(c.As) {
			return errors.Newf(codes.FailedPrecondition, "destination column %s is part of the group key for series %s and cannot be overridden", c.As, key.String())
		}
	}
	return nil
}

// validateInputColumns will ensure that all referenced input columns
// exist within the column reader. It does not verify types are valid.
func (t *tableTransformation) validateInputColumns(key flux.GroupKey, cols []flux.ColMeta) error {
	for _, c := range t.spec.Columns {
		if idx := execute.ColIdx(c.Column, cols); idx < 0 {
			return errors.Newf(codes.FailedPrecondition, "source column %s does not exist in series %s", c.Column, key.String())
		}
	}
	return nil
}

func (t *tableTransformation) initializeAggregateState(cr flux.ColReader) ([]values.Value, error) {
	state := make([]values.Value, len(t.spec.Columns))
	for i, c := range t.spec.Columns {
		j := execute.ColIdx(c.Column, cr.Cols())
		if j < 0 {
			return nil, errors.Newf(codes.FailedPrecondition, "missing column: %s", c.Column)
		}

		colType := cr.Cols()[j].Type
		inType := semantic.NewObjectType([]semantic.PropertyType{
			{Key: []byte("values"), Value: semantic.NewArrayType(flux.SemanticType(colType))},
		})
		scope := compiler.ToScope(c.Init.Scope)
		fn, err := compiler.Compile(scope, c.Init.Fn, inType)
		if err != nil {
			return nil, err
		}

		input := values.NewObject(inType)
		input.Set("values", arrowutil.NewArrayValue(
			table.Values(cr, j),
			colType,
		))

		columnState, err := fn.Eval(t.ctx, input)
		if err != nil {
			return nil, err
		}
		state[i] = columnState
	}
	return state, nil
}

func (t *tableTransformation) processBuffer(cr flux.ColReader, state []values.Value) error {
	for i, c := range t.spec.Columns {
		j := execute.ColIdx(c.Column, cr.Cols())
		if j < 0 {
			return errors.Newf(codes.FailedPrecondition, "missing column: %s", c.Column)
		}

		colType := cr.Cols()[j].Type
		inType := semantic.NewObjectType([]semantic.PropertyType{
			{Key: []byte("values"), Value: semantic.NewArrayType(flux.SemanticType(colType))},
			{Key: []byte("state"), Value: state[i].Type()},
		})
		scope := compiler.ToScope(c.Reduce.Scope)
		fn, err := compiler.Compile(scope, c.Reduce.Fn, inType)
		if err != nil {
			return err
		}

		input := values.NewObject(inType)
		input.Set("values", arrowutil.NewArrayValue(
			table.Values(cr, j),
			colType,
		))
		input.Set("state", state[i])

		columnState, err := fn.Eval(t.ctx, input)
		if err != nil {
			return err
		}
		state[i] = columnState
	}
	return nil
}

func (t *tableTransformation) computeFromState(state []values.Value) ([]values.Value, error) {
	aggregateValues := make([]values.Value, len(state))

	// Compute the final value for each column.
	for i, c := range t.spec.Columns {
		inType := semantic.NewObjectType([]semantic.PropertyType{
			{Key: []byte("state"), Value: state[i].Type()},
		})
		scope := compiler.ToScope(c.Compute.Scope)
		fn, err := compiler.Compile(scope, c.Compute.Fn, inType)
		if err != nil {
			return nil, err
		}

		input := values.NewObject(inType)
		input.Set("state", state[i])

		v, err := fn.Eval(t.ctx, input)
		if err != nil {
			return nil, err
		}
		aggregateValues[i] = v
	}
	return aggregateValues, nil
}

func (t *tableTransformation) buildTable(key flux.GroupKey, aggregates []values.Value) (flux.Table, error) {
	// Build the schema from the types in the group key and the
	// types from the computed aggregates.
	builder := table.NewArrowBuilder(key, t.alloc)
	for _, c := range key.Cols() {
		if _, err := builder.AddCol(c); err != nil {
			return nil, err
		}
	}

	for i, c := range t.spec.Columns {
		colType := flux.ColumnType(aggregates[i].Type())
		if _, err := builder.AddCol(flux.ColMeta{
			Label: c.As,
			Type:  colType,
		}); err != nil {
			return nil, err
		}
	}

	// Append a single value for each of the key columns.
	// We do not use slices here because, since it is always one
	// value, it is potentially more expensive to keep the existing
	// data as a slice than it is to copy a single data point.
	for j, c := range key.Cols() {
		idx := execute.ColIdx(c.Label, builder.Cols())
		if err := arrow.AppendValue(builder.Builders[idx], key.Value(j)); err != nil {
			return nil, err
		}
	}

	// Append a value for each of the values that were computed.
	for i, c := range t.spec.Columns {
		idx := execute.ColIdx(c.As, builder.Cols())
		if err := arrow.AppendValue(builder.Builders[idx], aggregates[i]); err != nil {
			return nil, err
		}
	}
	return builder.Table()
}
