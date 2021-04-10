package universe

import (
	"context"
	"sort"

	"github.com/apache/arrow/go/arrow/array"
	arrowmem "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/execkit"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const GroupKind = "group"

const (
	groupModeBy     = "by"
	groupModeExcept = "except"
)

type GroupMode flux.GroupMode

func (g *GroupMode) ReadArg(name string, arg values.Value, a *flux.Administration) error {
	mode, err := validateGroupMode(arg.Str())
	if err != nil {
		return err
	}
	*g = GroupMode(mode)
	return nil
}

func init() {
	execkit.RegisterTransformation(&GroupProcedureSpec{})
}

func validateGroupMode(mode string) (flux.GroupMode, error) {
	switch mode {
	case groupModeBy:
		return flux.GroupModeBy, nil
	case groupModeExcept:
		return flux.GroupModeExcept, nil
	default:
		return flux.GroupModeNone, errors.New(codes.Invalid, `invalid group mode: must be "by" or "except"`)
	}
}

type GroupProcedureSpec struct {
	GroupMode flux.GroupMode `flux:"mode"`
	GroupKeys []string       `flux:"columns"`
}

func (s *GroupProcedureSpec) CreateTransformation(id execute.DatasetID, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	t, d := NewGroupTransformation(s, id, a.Allocator())
	return t, d, nil
}

func (s *GroupProcedureSpec) ReadArgs(args flux.Arguments, a *flux.Administration) error {
	if err := a.AddParentFromArgs(args); err != nil {
		return err
	}

	if mode, ok, err := args.GetString("mode"); err != nil {
		return err
	} else if ok {
		m, err := validateGroupMode(mode)
		if err != nil {
			return err
		}
		s.GroupMode = m
	} else {
		s.GroupMode = flux.GroupModeBy
	}

	if columns, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return err
	} else if ok {
		s.GroupKeys, err = interpreter.ToStringArray(columns)
		if err != nil {
			return err
		}
	} else {
		s.GroupKeys = []string{}
	}

	return nil
}

func (s *GroupProcedureSpec) Kind() plan.ProcedureKind {
	return GroupKind
}

func (s *GroupProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.GroupKeys = make([]string, len(s.GroupKeys))
	copy(ns.GroupKeys, s.GroupKeys)
	return &ns
}

type groupTransformation struct {
	mode flux.GroupMode
	keys []string
}

func NewGroupTransformation(spec *GroupProcedureSpec, id execute.DatasetID, mem *memory.Allocator) (execute.Transformation, execute.Dataset) {
	sort.Strings(spec.GroupKeys)
	t, d, _ := execkit.NewGroupTransformation(id, &groupTransformation{
		mode: spec.GroupMode,
		keys: spec.GroupKeys,
	}, mem)
	return t, d
}

func (t *groupTransformation) Process(view table.View, d *execkit.Dataset, mem arrowmem.Allocator) error {
	// Determine the group key of this table if the grouped columns
	// are all part of the group key.
	if key, ok, err := t.getTableKey(view); err != nil {
		return err
	} else if ok {
		buffer := arrow.TableBuffer{
			GroupKey: key,
			Columns:  view.Cols(),
			Values:   make([]array.Interface, view.NCols()),
		}
		for j := range buffer.Values {
			buffer.Values[j] = view.Values(j)
		}
		return d.Process(table.ViewFromBuffer(buffer))
	}

	// We are grouping by something that is not within the group key,
	// so we have to determine which row goes in which column.
	// TODO(jsternberg): This can probably be optimized for memory, but
	// not going to do that at the moment.
	return t.groupByRow(view, d, mem)
}

// getTableKey returns the table key if the entire table matches
// the same table key. If the entire table does not match the key,
// this will return false and no key will be returned.
func (t *groupTransformation) getTableKey(tbl table.View) (flux.GroupKey, bool, error) {
	var indices []int
	switch t.mode {
	case flux.GroupModeBy:
		indices = make([]int, 0, len(t.keys))
		for _, label := range t.keys {
			if execute.ColIdx(label, tbl.Cols()) < 0 {
				// Skip past this label since it doesn't exist in the table.
				continue
			}

			// If this column is in the table but not part of the group key,
			// return false since this table cannot be easily categorized.
			idx := execute.ColIdx(label, tbl.Key().Cols())
			if idx < 0 {
				return nil, false, nil
			}
			indices = append(indices, idx)
		}
	case flux.GroupModeExcept:
		indices = make([]int, 0, len(tbl.Cols()))
		for _, c := range tbl.Cols() {
			// If this string is part of except, then it is not included.
			if execute.ContainsStr(t.keys, c.Label) {
				continue
			}

			// If this column is not part of the group key, return false.
			idx := execute.ColIdx(c.Label, tbl.Key().Cols())
			if idx < 0 {
				return nil, false, nil
			}
			indices = append(indices, idx)
		}
	default:
		panic(errors.Newf(codes.Internal, "unsupported group mode: %v", t.mode))
	}

	// Produce the group key from the indices.
	cols := make([]flux.ColMeta, len(indices))
	vs := make([]values.Value, len(indices))
	for j, idx := range indices {
		cols[j], vs[j] = tbl.Key().Cols()[idx], tbl.Key().Value(idx)
	}
	return execute.NewGroupKey(cols, vs), true, nil
}

// groupByRow will determine which table each row belongs to
// and to append them to that table.
func (t *groupTransformation) groupByRow(tbl table.View, d *execkit.Dataset, mem arrowmem.Allocator) error {
	var on map[string]bool
	switch t.mode {
	case flux.GroupModeBy:
		on = make(map[string]bool, len(t.keys))
		for _, key := range t.keys {
			on[key] = true
		}
	case flux.GroupModeExcept:
		on = make(map[string]bool, len(tbl.Cols()))
		for _, c := range tbl.Cols() {
			if !execute.ContainsStr(t.keys, c.Label) {
				on[c.Label] = true
			}
		}
	}

	// Construct a builder cache for the built tables.
	cache := table.BuilderCache{
		New: func(key flux.GroupKey) table.Builder {
			return table.NewArrowBuilder(key, mem)
		},
	}
	buffer := tbl.Buffer()
	for i, l := 0, buffer.Len(); i < l; i++ {
		key := execute.GroupKeyForRowOn(i, &buffer, on)
		ab, created := table.GetArrowBuilder(key, &cache)
		if created {
			for _, c := range buffer.Cols() {
				_, _ = ab.AddCol(c)
			}
		}
		for j := range buffer.Cols() {
			if err := t.appendValueFromRow(ab.Builders[j], &buffer, i, j); err != nil {
				return err
			}
		}
	}

	// Pass a view of each table we grouped to the downstream datasets.
	return cache.ForEach(func(key flux.GroupKey, builder table.Builder) error {
		buf, err := builder.(*table.ArrowBuilder).Buffer()
		if err != nil {
			return err
		}
		return d.Process(table.ViewFromBuffer(buf))
	})
}

func (t *groupTransformation) appendValueFromRow(b array.Builder, cr flux.ColReader, i, j int) error {
	switch cr.Cols()[j].Type {
	case flux.TInt:
		b := b.(*array.Int64Builder)
		vs := cr.Ints(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TUInt:
		b := b.(*array.Uint64Builder)
		vs := cr.UInts(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TFloat:
		b := b.(*array.Float64Builder)
		vs := cr.Floats(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TString:
		b := b.(*array.BinaryBuilder)
		vs := cr.Strings(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TBool:
		b := b.(*array.BooleanBuilder)
		vs := cr.Bools(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TTime:
		b := b.(*array.Int64Builder)
		vs := cr.Times(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	default:
		return errors.New(codes.Internal, "invalid builder type")
	}
	return nil
}

// `MergeGroupRule` merges two group operations and keeps only the last one
type MergeGroupRule struct{}

func (r MergeGroupRule) Name() string {
	return "MergeGroupRule"
}

// returns the pattern that matches `group |> group`
func (r MergeGroupRule) Pattern() plan.Pattern {
	return plan.Pat(GroupKind, plan.Pat(GroupKind, plan.Any()))
}

func (r MergeGroupRule) Rewrite(ctx context.Context, lastGroup plan.Node) (plan.Node, bool, error) {
	firstGroup := lastGroup.Predecessors()[0]
	lastSpec := lastGroup.ProcedureSpec().(*GroupProcedureSpec)

	if lastSpec.GroupMode != flux.GroupModeBy &&
		lastSpec.GroupMode != flux.GroupModeExcept {
		return lastGroup, false, nil
	}

	merged, err := plan.MergeToLogicalNode(lastGroup, firstGroup, lastSpec.Copy())
	if err != nil {
		return nil, false, err
	}

	return merged, true, nil
}
