package universe

import (
	"context"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const RenameKind = "rename"
const DropKind = "drop"
const KeepKind = "keep"
const DuplicateKind = "duplicate"

type RenameOpSpec struct {
	Columns map[string]string            `json:"columns"`
	Fn      interpreter.ResolvedFunction `json:"fn"`
}

type DropOpSpec struct {
	Columns   []string                     `json:"columns"`
	Predicate interpreter.ResolvedFunction `json:"fn"`
}

type KeepOpSpec struct {
	Columns   []string                     `json:"columns"`
	Predicate interpreter.ResolvedFunction `json:"fn"`
}

type DuplicateOpSpec struct {
	Column string `json:"columns"`
	As     string `json:"as"`
}

// The base kind for SchemaMutations
const SchemaMutationKind = "SchemaMutation"

// A list of all operations which should map to the SchemaMutationProcedure
// Added to dynamically upon calls to `Register()`
var SchemaMutationOps = []flux.OperationKind{}

// A MutationRegistrar contains information needed
// to register a type of Operation Spec
// that will be converted into a SchemaMutator
// and embedded in a SchemaMutationProcedureSpec.
// Operations with a corresponding MutationRegistrar
// should not have their own ProcedureSpec.
type MutationRegistrar struct {
	Kind   flux.OperationKind
	Type   semantic.MonoType
	Create flux.CreateOperationSpec
	New    flux.NewOperationSpec
}

func (m MutationRegistrar) Register() {
	t := runtime.MustLookupBuiltinType("universe", string(m.Kind))
	runtime.RegisterPackageValue("universe", string(m.Kind), flux.MustValue(flux.FunctionValue(string(m.Kind), m.Create, t)))
	flux.RegisterOpSpec(m.Kind, m.New)

	// Add to list of SchemaMutations which should map to a
	// SchemaMutationProcedureSpec
	SchemaMutationOps = append(SchemaMutationOps, m.Kind)
}

// A list of all MutationRegistrars to register.
// To register a new mutation, add an entry to this list.
var Registrars = []MutationRegistrar{
	{
		Kind:   RenameKind,
		Type:   runtime.MustLookupBuiltinType("universe", "rename"),
		Create: createRenameOpSpec,
		New:    newRenameOp,
	},
	{
		Kind:   DropKind,
		Type:   runtime.MustLookupBuiltinType("universe", "drop"),
		Create: createDropOpSpec,
		New:    newDropOp,
	},
	{
		Kind:   KeepKind,
		Type:   runtime.MustLookupBuiltinType("universe", "keep"),
		Create: createKeepOpSpec,
		New:    newKeepOp,
	},
	{
		Kind:   DuplicateKind,
		Type:   runtime.MustLookupBuiltinType("universe", "duplicate"),
		Create: createDuplicateOpSpec,
		New:    newDuplicateOp,
	},
}

func init() {
	for _, r := range Registrars {
		r.Register()
	}

	plan.RegisterProcedureSpec(SchemaMutationKind, newDualImplSpec(newSchemaMutationProcedure), SchemaMutationOps...)
	execute.RegisterTransformation(SchemaMutationKind, createDualImplTf(createSchemaMutationTransformation, createDeprecatedSchemaMutationTransformation))
}

func createRenameOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	var cols values.Object
	if c, ok, err := args.GetObject("columns"); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var renameFn interpreter.ResolvedFunction
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		if fn, err := interpreter.ResolveFunction(f); err != nil {
			return nil, err
		} else {
			renameFn = fn
		}
	}

	if cols == nil && renameFn.Fn == nil {
		return nil, errors.New(codes.Invalid, "rename error: neither column list nor map function provided")
	}

	if cols != nil && renameFn.Fn != nil {
		return nil, errors.New(codes.Invalid, "rename error: both column list and map function provided")
	}

	spec := &RenameOpSpec{
		Fn: renameFn,
	}

	if cols != nil {
		var err error
		renameCols := make(map[string]string, cols.Len())
		// Check types of object values manually
		cols.Range(func(name string, v values.Value) {
			if err != nil {
				return
			}
			if v.Type().Nature() != semantic.String {
				err = errors.Newf(codes.Invalid, "rename error: columns object contains non-string value of type %s", v.Type())
				return
			}
			renameCols[name] = v.Str()
		})
		if err != nil {
			return nil, err
		}
		spec.Columns = renameCols
	}

	return spec, nil
}

func createDropOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	var cols values.Array
	if c, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var dropPredicate interpreter.ResolvedFunction
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}

		dropPredicate = fn
	}

	if cols == nil && dropPredicate.Fn == nil {
		return nil, errors.New(codes.Invalid, "drop error: neither column list nor predicate function provided")
	}

	if cols != nil && dropPredicate.Fn != nil {
		return nil, errors.New(codes.Invalid, "drop error: both column list and predicate provided")
	}

	var dropCols []string
	var err error
	if cols != nil {
		dropCols, err = interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
	}

	return &DropOpSpec{
		Columns:   dropCols,
		Predicate: dropPredicate,
	}, nil
}

func createKeepOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	var cols values.Array
	if c, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var keepPredicate interpreter.ResolvedFunction
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}

		keepPredicate = fn
	}

	if cols == nil && keepPredicate.Fn == nil {
		return nil, errors.New(codes.Invalid, "keep error: neither column list nor predicate function provided")
	}

	if cols != nil && keepPredicate.Fn != nil {
		return nil, errors.New(codes.Invalid, "keep error: both column list and predicate provided")
	}

	var keepCols []string
	var err error
	if cols != nil {
		keepCols, err = interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
	}

	return &KeepOpSpec{
		Columns:   keepCols,
		Predicate: keepPredicate,
	}, nil
}

func createDuplicateOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	col, err := args.GetRequiredString("column")
	if err != nil {
		return nil, err
	}

	newName, err := args.GetRequiredString("as")
	if err != nil {
		return nil, err
	}

	return &DuplicateOpSpec{
		Column: col,
		As:     newName,
	}, nil
}

func newRenameOp() flux.OperationSpec {
	return new(RenameOpSpec)
}

func (s *RenameOpSpec) Kind() flux.OperationKind {
	return RenameKind
}

func newDropOp() flux.OperationSpec {
	return new(DropOpSpec)
}

func (s *DropOpSpec) Kind() flux.OperationKind {
	return DropKind
}

func newKeepOp() flux.OperationSpec {
	return new(KeepOpSpec)
}

func (s *KeepOpSpec) Kind() flux.OperationKind {
	return KeepKind
}

func newDuplicateOp() flux.OperationSpec {
	return new(DuplicateOpSpec)
}

func (s *DuplicateOpSpec) Kind() flux.OperationKind {
	return DuplicateKind
}

func (s *RenameOpSpec) Copy() SchemaMutation {
	newCols := make(map[string]string, len(s.Columns))
	for k, v := range s.Columns {
		newCols[k] = v
	}

	return &RenameOpSpec{
		Columns: newCols,
		Fn:      s.Fn.Copy(),
	}
}

func (s *DropOpSpec) Copy() SchemaMutation {
	newCols := make([]string, len(s.Columns))
	copy(newCols, s.Columns)

	return &DropOpSpec{
		Columns:   newCols,
		Predicate: s.Predicate.Copy(),
	}
}

func (s *KeepOpSpec) Copy() SchemaMutation {
	newCols := make([]string, len(s.Columns))
	copy(newCols, s.Columns)

	return &KeepOpSpec{
		Columns:   newCols,
		Predicate: s.Predicate.Copy(),
	}
}

func (s *DuplicateOpSpec) Copy() SchemaMutation {
	return &DuplicateOpSpec{
		Column: s.Column,
		As:     s.As,
	}
}

func (s *RenameOpSpec) Mutator() (SchemaMutator, error) {
	m, err := NewRenameMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *DropOpSpec) Mutator() (SchemaMutator, error) {
	m, err := NewDropKeepMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *KeepOpSpec) Mutator() (SchemaMutator, error) {
	m, err := NewDropKeepMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *DuplicateOpSpec) Mutator() (SchemaMutator, error) {
	m, err := NewDuplicateMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type SchemaMutationProcedureSpec struct {
	plan.DefaultCost
	Mutations []SchemaMutation
}

func (s *SchemaMutationProcedureSpec) Kind() plan.ProcedureKind {
	return SchemaMutationKind
}

func (s *SchemaMutationProcedureSpec) Copy() plan.ProcedureSpec {
	newMutations := make([]SchemaMutation, len(s.Mutations))
	for i, m := range s.Mutations {
		newMutations[i] = m.Copy()
	}

	return &SchemaMutationProcedureSpec{
		Mutations: newMutations,
	}
}

func newSchemaMutationProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(SchemaMutation)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T doesn't implement SchemaMutation", qs)
	}

	return &SchemaMutationProcedureSpec{
		Mutations: []SchemaMutation{s},
	}, nil
}

type schemaMutationTransformation struct {
	execute.ExecutionNode
	d        execute.Dataset
	cache    table.BuilderCache
	ctx      context.Context
	mutators []SchemaMutator
}

func createSchemaMutationTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SchemaMutationProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewSchemaMutationTransformation(a.Context(), s, id, a.Allocator())
}

func NewSchemaMutationTransformation(ctx context.Context, spec *SchemaMutationProcedureSpec, id execute.DatasetID, mem *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	mutators := make([]SchemaMutator, len(spec.Mutations))
	for i, mutation := range spec.Mutations {
		m, err := mutation.Mutator()
		if err != nil {
			return nil, nil, err
		}
		mutators[i] = m
	}

	t := &schemaMutationTransformation{
		cache: table.BuilderCache{
			New: func(key flux.GroupKey) table.Builder {
				return table.NewBufferedBuilder(key, mem)
			},
		},
		mutators: mutators,
		ctx:      ctx,
	}
	t.d = table.NewDataset(id, &t.cache)
	return t, t.d, nil
}

func (t *schemaMutationTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	ctx := NewBuilderContext(tbl)
	for _, m := range t.mutators {
		if err := m.Mutate(t.ctx, ctx); err != nil {
			return err
		}
	}

	mutTable, err := t.mutateTable(tbl, ctx)
	if err != nil {
		tbl.Done()
		return err
	}
	builder, _ := table.GetBufferedBuilder(mutTable.Key(), &t.cache)
	return builder.AppendTable(mutTable)
}

func (t *schemaMutationTransformation) mutateTable(in flux.Table, ctx *BuilderContext) (flux.Table, error) {
	// Check the schema for columns with the same name.
	cols := ctx.Cols()
	for i, c := range cols {
		for j := range cols[:i] {
			if cols[j].Label == c.Label {
				return nil, errors.Newf(codes.FailedPrecondition, "column %d and %d have the same name (%q) which is not allowed", j, i, c.Label)
			}
		}
	}

	return &mutateTable{
		in:  in,
		ctx: ctx,
	}, nil
}

func (t *schemaMutationTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *schemaMutationTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *schemaMutationTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *schemaMutationTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type mutateTable struct {
	in  flux.Table
	ctx *BuilderContext
}

func (m *mutateTable) Key() flux.GroupKey   { return m.ctx.Key() }
func (m *mutateTable) Cols() []flux.ColMeta { return m.ctx.Cols() }
func (m *mutateTable) Empty() bool          { return m.in.Empty() }
func (m *mutateTable) Done()                { m.in.Done() }

func (m *mutateTable) Do(f func(flux.ColReader) error) error {
	return m.in.Do(func(cr flux.ColReader) error {
		indices := m.ctx.ColMap()
		buffer := &arrow.TableBuffer{
			GroupKey: m.ctx.Key(),
			Columns:  m.ctx.Cols(),
			Values:   make([]array.Interface, len(indices)),
		}
		for j, idx := range indices {
			buffer.Values[j] = table.Values(cr, idx)
			buffer.Values[j].Retain()
		}
		return f(buffer)
	})
}
