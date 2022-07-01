package universe

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type BuilderContext struct {
	TableColumns []flux.ColMeta
	TableKey     flux.GroupKey
	ColIdxMap    []int
}

func NewBuilderContext(tbl flux.Table) *BuilderContext {
	colMap := make([]int, len(tbl.Cols()))
	for i := range tbl.Cols() {
		colMap[i] = i
	}

	cols := make([]flux.ColMeta, len(tbl.Cols()))
	copy(cols, tbl.Cols())

	return &BuilderContext{
		TableColumns: cols,
		TableKey:     tbl.Key(),
		ColIdxMap:    colMap,
	}
}

func (b *BuilderContext) Cols() []flux.ColMeta {
	return b.TableColumns
}

func (b *BuilderContext) Key() flux.GroupKey {
	return b.TableKey
}

func (b *BuilderContext) ColMap() []int {
	return b.ColIdxMap
}

type SchemaMutator interface {
	Mutate(ctx context.Context, bctx *BuilderContext) error
}

type SchemaMutation interface {
	Mutator() (SchemaMutator, error)
	Copy() SchemaMutation
}

func toStringSet(arr []string) map[string]bool {
	if arr == nil {
		return nil
	}
	set := make(map[string]bool, len(arr))
	for _, s := range arr {
		set[s] = true
	}

	return set
}

func checkCol(label string, cols []flux.ColMeta) error {
	if execute.ColIdx(label, cols) < 0 {
		return errors.Newf(codes.FailedPrecondition, `column "%s" doesn't exist`, label)
	}
	return nil
}

const schemaFnMutatorParamName = "column"

type schemaFnMutator struct {
	Fn    compiler.Func
	Input values.Object
}

func (m *schemaFnMutator) compile(fn interpreter.ResolvedFunction) error {
	in := semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte(schemaFnMutatorParamName), Value: semantic.BasicString},
	})
	preparedFn, err := compiler.Compile(compiler.ToScope(fn.Scope), fn.Fn, in)
	if err != nil {
		return err
	}

	m.Fn = preparedFn
	m.Input = values.NewObject(in)
	return nil
}

func (m *schemaFnMutator) eval(ctx context.Context, column string) (values.Value, error) {
	m.Input.Set(schemaFnMutatorParamName, values.NewString(column))
	v, err := m.Fn.Eval(ctx, m.Input)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type RenameMutator struct {
	schemaFnMutator
	Columns map[string]string
}

func NewRenameMutator(qs flux.OperationSpec) (*RenameMutator, error) {
	s, ok := qs.(*RenameOpSpec)

	m := &RenameMutator{}
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	if s.Columns != nil {
		m.Columns = s.Columns
	}

	if s.Fn.Fn != nil {
		if err := m.compile(s.Fn); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *RenameMutator) renameCol(ctx context.Context, col *flux.ColMeta) error {
	if col == nil {
		return errors.New(codes.FailedPrecondition, "rename error: cannot rename nil column")
	}
	if m.Columns != nil {
		if newName, ok := m.Columns[col.Label]; ok {
			col.Label = newName
		}
	} else if m.Fn != nil {
		newName, err := m.eval(ctx, col.Label)
		if err != nil {
			return err
		}
		col.Label = newName.Str()
	}
	return nil
}

func (m *RenameMutator) checkColumns(tableCols []flux.ColMeta) error {
	for c := range m.Columns {
		if err := checkCol(c, tableCols); err != nil {
			return errors.Wrap(err, codes.Inherit, "rename error")
		}
	}
	return nil
}

func (m *RenameMutator) Mutate(ctx context.Context, bctx *BuilderContext) error {
	if err := m.checkColumns(bctx.Cols()); err != nil {
		return err
	}

	keyCols := make([]flux.ColMeta, 0, len(bctx.Cols()))
	keyValues := make([]values.Value, 0, len(bctx.Cols()))

	for i := range bctx.Cols() {
		keyIdx := execute.ColIdx(bctx.TableColumns[i].Label, bctx.Key().Cols())
		keyed := keyIdx >= 0

		if err := m.renameCol(ctx, &bctx.TableColumns[i]); err != nil {
			return err
		}

		if keyed {
			keyCols = append(keyCols, bctx.TableColumns[i])
			keyValues = append(keyValues, bctx.Key().Value(keyIdx))
		}
	}

	bctx.TableKey = execute.NewGroupKey(keyCols, keyValues)

	return nil
}

type DropKeepMutator struct {
	schemaFnMutator
	KeepCols      map[string]bool
	DropCols      map[string]bool
	FlipPredicate bool
}

func NewDropKeepMutator(qs flux.OperationSpec) (*DropKeepMutator, error) {
	m := &DropKeepMutator{}

	switch s := qs.(type) {
	case *DropOpSpec:
		if s.Columns != nil {
			m.DropCols = toStringSet(s.Columns)
		}
		if s.Predicate.Fn != nil {
			if err := m.compile(s.Predicate); err != nil {
				return nil, err
			}
		}
	case *KeepOpSpec:
		if s.Columns != nil {
			m.KeepCols = toStringSet(s.Columns)
		}
		if s.Predicate.Fn != nil {
			if err := m.compile(s.Predicate); err != nil {
				return nil, err
			}
			m.FlipPredicate = true
		}
	default:
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return m, nil
}

func (m *DropKeepMutator) shouldDrop(ctx context.Context, col string) (bool, error) {
	v, err := m.eval(ctx, col)
	if err != nil {
		return false, err
	}
	shouldDrop := !v.IsNull() && v.Bool()
	if m.FlipPredicate {
		shouldDrop = !shouldDrop
	}
	return shouldDrop, nil
}

func (m *DropKeepMutator) shouldDropCol(ctx context.Context, col string) (bool, error) {
	if m.DropCols != nil {
		if _, exists := m.DropCols[col]; exists {
			return true, nil
		}
	} else if m.Fn != nil {
		return m.shouldDrop(ctx, col)
	}
	return false, nil
}

func (m *DropKeepMutator) keepToDropCols(cols []flux.ColMeta) {
	// If we have columns we want to keep, we can accomplish this by inverting the Cols map,
	// and storing it in Cols.
	//  With a keep operation, Cols may be changed with each call to `Mutate`, but
	// `Cols` will not be.
	if m.KeepCols != nil {
		exclusiveDropCols := make(map[string]bool, len(cols))
		for _, c := range cols {
			if _, ok := m.KeepCols[c.Label]; !ok {
				exclusiveDropCols[c.Label] = true
			}
		}
		m.DropCols = exclusiveDropCols
	}
}

func (m *DropKeepMutator) Mutate(ctx context.Context, bctx *BuilderContext) error {

	m.keepToDropCols(bctx.Cols())

	keyCols := make([]flux.ColMeta, 0, len(bctx.Cols()))
	keyValues := make([]values.Value, 0, len(bctx.Cols()))
	newCols := make([]flux.ColMeta, 0, len(bctx.Cols()))

	oldColMap := bctx.ColMap()
	newColMap := make([]int, 0, len(bctx.Cols()))

	for i, c := range bctx.Cols() {
		if shouldDrop, err := m.shouldDropCol(ctx, c.Label); err != nil {
			return err
		} else if shouldDrop {
			continue
		}

		keyIdx := execute.ColIdx(c.Label, bctx.Key().Cols())
		if keyIdx >= 0 {
			keyCols = append(keyCols, c)
			keyValues = append(keyValues, bctx.Key().Value(keyIdx))
		}
		newCols = append(newCols, c)
		newColMap = append(newColMap, oldColMap[i])
	}

	bctx.TableColumns = newCols
	bctx.TableKey = execute.NewGroupKey(keyCols, keyValues)
	bctx.ColIdxMap = newColMap

	return nil
}

type DuplicateMutator struct {
	Column string
	As     string
}

// TODO: figure out what we'd like to do with the context and dependencies here
func NewDuplicateMutator(qs flux.OperationSpec) (*DuplicateMutator, error) {
	s, ok := qs.(*DuplicateOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &DuplicateMutator{
		Column: s.Column,
		As:     s.As,
	}, nil
}

func (m *DuplicateMutator) Mutate(ctx context.Context, bctx *BuilderContext) error {
	fromIdx := execute.ColIdx(m.Column, bctx.Cols())
	if fromIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, `duplicate error: column "%s" doesn't exist`, m.Column)
	}

	newCol := duplicate(bctx.TableColumns[fromIdx], m.As)
	asIdx := execute.ColIdx(m.As, bctx.Cols())
	if asIdx < 0 {
		bctx.TableColumns = append(bctx.TableColumns, newCol)
		bctx.ColIdxMap = append(bctx.ColIdxMap, bctx.ColIdxMap[fromIdx])
		asIdx = len(bctx.TableColumns) - 1
	} else {
		bctx.TableColumns[asIdx] = newCol
		bctx.ColIdxMap[asIdx] = bctx.ColIdxMap[fromIdx]
	}
	asKeyIdx := execute.ColIdx(bctx.TableColumns[asIdx].Label, bctx.Key().Cols())
	if asKeyIdx >= 0 {
		newKeyCols := append(bctx.Key().Cols()[:0:0], bctx.Key().Cols()...)
		newKeyVals := append(bctx.Key().Values()[:0:0], bctx.Key().Values()...)
		fromKeyIdx := execute.ColIdx(m.Column, newKeyCols)
		if fromKeyIdx >= 0 {
			newKeyCols[asKeyIdx] = newCol
			newKeyVals[asKeyIdx] = newKeyVals[fromKeyIdx]
		} else {
			newKeyCols = append(newKeyCols[:asKeyIdx], newKeyCols[asKeyIdx+1:]...)
		}
		bctx.TableKey = execute.NewGroupKey(newKeyCols, newKeyVals)
	}

	return nil
}

func duplicate(col flux.ColMeta, dupName string) flux.ColMeta {
	return flux.ColMeta{
		Type:  col.Type,
		Label: dupName,
	}
}

// TODO: determine pushdown rules
/*
func (s *SchemaMutationProcedureSpec) PushDownRules() []plan.PushDownRule {
	return []plan.PushDownRule{{
		Root:    SchemaMutationKind,
		Through: nil,
		Match:   nil,
	}}
}

func (s *SchemaMutationProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	rootSpec := root.Spec.(*SchemaMutationProcedureSpec)
	rootSpec.Mutations = append(rootSpec.Mutations, s.Mutations...)
}
*/
