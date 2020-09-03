package aggregate

import (
	"context"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
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
	Fill    values.Value
}

func init() {
	runtime.RegisterPackageValue(pkgpath, "table", flux.MustValue(flux.FunctionValue(
		"table",
		createTableOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "table"),
	)))
	plan.RegisterProcedureSpec(TableKind, newTableProcedure, TableKind)
	execute.RegisterTransformation(TableKind, createTableTransformation)
	runtime.RegisterPackageValue(pkgpath, "null", fillNull)
	runtime.RegisterPackageValue(pkgpath, "none", fillNone)
}

func tableColumnsFromObject(object values.Object) (columns []TableColumn, err error) {
	columns = make([]TableColumn, 0, object.Len())
	object.Range(func(as string, v values.Value) {
		if err != nil {
			return
		} else if v.Type().Nature() != semantic.Object {
			err = errors.Newf(codes.Invalid, "aggregate for column %q is not an aggregate object", as)
			return
		}

		with := v.Object()
		column, ok := with.Get("column")
		if !ok {
			err = errors.Newf(codes.Invalid, "aggregate for column %q is missing a \"column\" property", as)
			return
		} else if got := column.Type().Nature(); got != semantic.String {
			err = errors.Newf(codes.Invalid, "aggregate object \"column\" property must be a string: got %s", got)
			return
		}

		fill, ok := with.Get("fill")
		if !ok {
			fill = fillNull
		}

		ac := TableColumn{
			Column: column.Str(),
			As:     as,
			Fill:   fill,
		}

		resolveFn := func(name string, with values.Object) (interpreter.ResolvedFunction, error) {
			fn, ok := with.Get(name)
			if !ok {
				return interpreter.ResolvedFunction{}, errors.Newf(codes.Invalid, "aggregate object is missing the %q function", name)
			}
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

		columns = append(columns, ac)
	})
	return columns, err
}

func createTableOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(TableOpSpec)

	columnsArg, err := args.GetRequiredObject("columns")
	if err != nil {
		return nil, err
	}

	columns, err := tableColumnsFromObject(columnsArg.Object())
	if err != nil {
		return nil, err
	}
	spec.Columns = columns

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
	execute.ExecutionNode
	d    *execute.PassthroughDataset
	spec *TableProcedureSpec
	ctx  context.Context
	mem  memory.Allocator
}

func NewTableTransformation(ctx context.Context, spec *TableProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &tableTransformation{
		d:    execute.NewPassthroughDataset(id),
		spec: spec,
		ctx:  ctx,
		mem:  mem,
	}
	return t, t.d, nil
}

func (t *tableTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *tableTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if err := t.validateInputTable(tbl); err != nil {
		return err
	}

	// Prepare each of the columns.
	columns, err := t.prepare(tbl.Cols(), 1)
	if err != nil {
		return err
	}

	// Iterate through each buffer to calculate the state.
	state := make([]values.Value, len(columns))
	if err := tbl.Do(func(cr flux.ColReader) error {
		for i, c := range columns {
			if err := c.Eval(t.ctx, cr, &state[i]); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Write the current state into the builders.
	for i, c := range columns {
		if err := c.Write(t.ctx, state[i]); err != nil {
			return err
		}
	}

	outTable, err := t.buildTable(tbl.Key(), columns)
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

func (t *tableTransformation) validateInputTable(tbl flux.Table) error {
	if err := t.validateGroupKey(tbl.Key()); err != nil {
		return err
	}
	return t.validateInputColumns(tbl.Key(), tbl.Cols())
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

func (t *tableTransformation) prepare(cols []flux.ColMeta, n int) ([]*columnState, error) {
	columns := make([]*columnState, len(t.spec.Columns))
	for i, c := range t.spec.Columns {
		j := execute.ColIdx(c.Column, cols)
		if j < 0 {
			return nil, errors.Newf(codes.FailedPrecondition, "missing column: %s", c.Column)
		}
		arrayType := semantic.NewArrayType(flux.SemanticType(cols[j].Type))

		cs := &columnState{Label: c.As, Column: j}
		if err := cs.compileInitFunc(c, arrayType); err != nil {
			return nil, err
		}
		if err := cs.compileReduceFunc(c, arrayType); err != nil {
			return nil, err
		}
		if err := cs.compileComputeFunc(c, n, t.mem); err != nil {
			return nil, err
		}
		columns[i] = cs
	}
	return columns, nil
}

func (t *tableTransformation) buildTable(key flux.GroupKey, columns []*columnState) (flux.Table, error) {
	// Construct the table schema.
	buffer := &arrow.TableBuffer{
		GroupKey: key,
		Columns:  make([]flux.ColMeta, len(key.Cols())+len(columns)),
	}
	buffer.Values = make([]array.Interface, len(buffer.Columns))

	// Create columns for the output values.
	offset := len(key.Cols())
	for i, cs := range columns {
		buffer.Columns[i+offset] = flux.ColMeta{
			Label: cs.Label,
			Type:  cs.Type,
		}
		buffer.Values[i] = cs.Builder.NewArray()
	}
	n := buffer.Values[0].Len()

	// Copy over the group key columns.
	for i, c := range key.Cols() {
		buffer.Columns[i] = c
		buffer.Values[i] = arrow.Repeat(key.Value(i), n, t.mem)
	}

	if err := buffer.Validate(); err != nil {
		buffer.Release()
		return nil, err
	}
	return table.FromBuffer(buffer), nil
}

type fillValue struct {
	values.Value
	null, none bool
}

var (
	fillNull = fillValue{Value: values.Null, null: true}
	fillNone = fillValue{Value: values.Null, none: true}
)

type columnState struct {
	Label  string
	Column int
	Type   flux.ColType
	Init   struct {
		Fn    compiler.Func
		Input semantic.MonoType
	}
	Reduce struct {
		Fn    compiler.Func
		Input semantic.MonoType
	}
	Compute struct {
		Fn    compiler.Func
		Input semantic.MonoType
	}
	Builder array.Builder
	Fill    struct {
		Null  bool
		None  bool
		Value values.Value
	}
}

func (cs *columnState) Eval(ctx context.Context, cr flux.ColReader, state *values.Value) (err error) {
	colType := cr.Cols()[cs.Column].Type
	arr := arrowutil.NewArrayValue(
		table.Values(cr, cs.Column),
		colType,
	)
	if *state == nil {
		input := values.NewObject(cs.Init.Input)
		input.Set("values", arr)

		*state, err = cs.Init.Fn.Eval(ctx, input)
		return err
	}

	input := values.NewObject(cs.Reduce.Input)
	input.Set("values", arr)
	input.Set("state", *state)

	*state, err = cs.Reduce.Fn.Eval(ctx, input)
	return err
}

// Write will compute the final value and write the value
// into the builder.
func (cs *columnState) Write(ctx context.Context, state values.Value) error {
	if state == nil {
		if cs.Fill.Null {
			cs.Builder.AppendNull()
			return nil
		} else if cs.Fill.None {
			// Do not append anything.
			return nil
		}
		return arrow.AppendValue(cs.Builder, cs.Fill.Value)
	}

	input := values.NewObject(cs.Compute.Input)
	input.Set("state", state)

	v, err := cs.Compute.Fn.Eval(ctx, input)
	if err != nil {
		return err
	}
	return arrow.AppendValue(cs.Builder, v)
}

func (cs *columnState) compileInitFunc(c TableColumn, arrayType semantic.MonoType) error {
	cs.Init.Input = semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("values"), Value: arrayType},
	})

	scope := compiler.ToScope(c.Init.Scope)
	fn, err := compiler.Compile(scope, c.Init.Fn, cs.Init.Input)
	if err != nil {
		return errors.Wrap(err, codes.Inherit, "error compiling aggregate init function")
	}
	cs.Init.Fn = fn
	return nil
}

func (cs *columnState) compileReduceFunc(c TableColumn, arrayType semantic.MonoType) error {
	cs.Reduce.Input = semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("values"), Value: arrayType},
		{Key: []byte("state"), Value: cs.Init.Fn.Type()},
	})
	scope := compiler.ToScope(c.Reduce.Scope)
	fn, err := compiler.Compile(scope, c.Reduce.Fn, cs.Reduce.Input)
	if err != nil {
		return errors.Wrap(err, codes.Inherit, "error compiling aggregate reduce function")
	}
	cs.Reduce.Fn = fn
	return nil
}

func (cs *columnState) compileComputeFunc(c TableColumn, n int, mem memory.Allocator) error {
	cs.Compute.Input = semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("state"), Value: cs.Reduce.Fn.Type()},
	})
	scope := compiler.ToScope(c.Compute.Scope)
	fn, err := compiler.Compile(scope, c.Compute.Fn, cs.Compute.Input)
	if err != nil {
		return errors.Wrap(err, codes.Inherit, "error compiling aggregate compute function")
	}
	cs.Compute.Fn = fn

	cs.Type = flux.ColumnType(cs.Compute.Fn.Type())
	if cs.Type == flux.TInvalid {
		return errors.Newf(codes.FailedPrecondition, "invalid output column type: %s", cs.Compute.Fn.Type())
	}
	cs.Builder = arrow.NewBuilder(cs.Type, mem)
	cs.Builder.Resize(n)

	if fv, ok := c.Fill.(fillValue); ok {
		cs.Fill.Null = fv.null
		cs.Fill.None = fv.none
	} else {
		if typ := flux.ColumnType(c.Fill.Type()); typ != cs.Type {
			return errors.Newf(codes.FailedPrecondition, "fill type does not match the output type: %s != %s", typ, cs.Type)
		}
		cs.Fill.Value = c.Fill
	}
	return nil
}
