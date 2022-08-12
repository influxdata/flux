package universe

import (
	"context"
	"sort"

	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/dataset"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const GroupKind = "group"

const (
	groupModeBy     = "by"
	groupModeExcept = "except"
)

type GroupOpSpec struct {
	Mode    string   `json:"mode"`
	Columns []string `json:"columns"`
}

func init() {
	groupSignature := runtime.MustLookupBuiltinType("universe", "group")

	runtime.RegisterPackageValue("universe", GroupKind, flux.MustValue(flux.FunctionValue(GroupKind, createGroupOpSpec, groupSignature)))
	plan.RegisterProcedureSpec(GroupKind, newGroupProcedure, GroupKind)
	plan.RegisterLogicalRules(MergeGroupRule{})
	execute.RegisterTransformation(GroupKind, createGroupTransformation)
}

func createGroupOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(GroupOpSpec)

	if mode, ok, err := args.GetString("mode"); err != nil {
		return nil, err
	} else if ok {
		if _, err := validateGroupMode(mode); err != nil {
			return nil, err
		}

		spec.Mode = mode
	} else {
		spec.Mode = groupModeBy
	}

	if columns, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.Columns, err = interpreter.ToStringArray(columns)
		if err != nil {
			return nil, err
		}
	} else {
		spec.Columns = []string{}
	}

	return spec, nil
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

func (s *GroupOpSpec) Kind() flux.OperationKind {
	return GroupKind
}

type GroupProcedureSpec struct {
	plan.DefaultCost
	GroupMode flux.GroupMode
	GroupKeys []string
}

func (s *GroupProcedureSpec) PassThroughAttribute(attrKey string) bool {
	switch attrKey {
	case plan.ParallelRunKey:
		return true
	}
	return false
}

func newGroupProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*GroupOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	mode, err := validateGroupMode(spec.Mode)
	if err != nil {
		return nil, err
	}

	p := &GroupProcedureSpec{
		GroupMode: mode,
		GroupKeys: spec.Columns,
	}
	return p, nil
}

func (s *GroupProcedureSpec) Kind() plan.ProcedureKind {
	return GroupKind
}
func (s *GroupProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(GroupProcedureSpec)

	ns.GroupMode = s.GroupMode

	ns.GroupKeys = make([]string, len(s.GroupKeys))
	copy(ns.GroupKeys, s.GroupKeys)

	return ns
}

func createGroupTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*GroupProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d, err := NewGroupTransformation(a.Context(), s, id, a.Allocator())
	if err != nil {
		return nil, nil, errors.Newf(codes.Internal, "could not create group transformation: %s", err)
	}
	return t, d, nil
}

type groupTransformationAdapter struct {
	t *groupTransformation
}

func (a *groupTransformationAdapter) Process(chunk table.Chunk, d *execute.TransportDataset, mem arrowmem.Allocator) error {
	// Determine the group key of this table if the grouped columns
	// are all part of the group key.
	if key, ok, err := a.t.getTableKey(chunk.Key(), chunk.Cols()); err != nil {
		return err
	} else if ok {
		buffer := arrow.TableBuffer{
			GroupKey: key,
			Columns:  chunk.Cols(),
			Values:   make([]array.Array, chunk.NCols()),
		}
		for j := range buffer.Values {
			buffer.Values[j] = chunk.Values(j)
			buffer.Values[j].Retain()
		}
		return d.Process(table.ChunkFromBuffer(buffer))
	}

	// We are grouping by something that is not within the group key,
	// so we have to determine which row goes in which column.
	// TODO(jsternberg): This can probably be optimized for memory, but
	// not going to do that at the moment.
	return a.t.groupChunkByRow(chunk, d, mem)
}

func (a *groupTransformationAdapter) Close() error { return nil }

type groupTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache table.BuilderCache
	mem   memory.Allocator

	mode flux.GroupMode
	keys []string
}

func NewGroupTransformation(ctx context.Context, spec *GroupProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &groupTransformation{
		cache: table.BuilderCache{
			New: func(key flux.GroupKey) table.Builder {
				return table.NewBufferedBuilder(key, mem)
			},
		},
		mem:  mem,
		mode: spec.GroupMode,
		keys: spec.GroupKeys,
	}
	t.d = dataset.New(id, &t.cache)
	sort.Strings(t.keys)
	if feature.GroupTransformationGroup().Enabled(ctx) {
		a := &groupTransformationAdapter{t: t}
		gt, d, err := execute.NewGroupTransformation(id, a, mem)
		if err != nil {
			return nil, nil, err
		}
		return gt, d, nil
	}
	return t, t.d, nil
}

func (t *groupTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *groupTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Determine the group key of this table if the grouped columns
	// are all part of the group key.
	if key, ok, err := t.getTableKey(tbl.Key(), tbl.Cols()); err != nil {
		return err
	} else if ok {
		ab, _ := table.GetBufferedBuilder(key, &t.cache)
		return t.appendTable(ab, tbl)
	}

	// We are grouping by something that is not within the group key,
	// so we have to determine which row goes in which column.
	// TODO(jsternberg): This can probably be optimized for memory, but
	// not going to do that at the moment.
	return t.groupByRow(tbl)
}

// getTableKey returns the table key if the entire table matches
// the same table key. If the entire table does not match the key,
// this will return false and no key will be returned.
func (t *groupTransformation) getTableKey(key flux.GroupKey, cols []flux.ColMeta) (flux.GroupKey, bool, error) {
	var indices []int
	switch t.mode {
	case flux.GroupModeBy:
		indices = make([]int, 0, len(t.keys))
		for _, label := range t.keys {
			if execute.ColIdx(label, cols) < 0 {
				// Skip past this label since it doesn't exist in the table.
				continue
			}

			// If this column is in the table but not part of the group key,
			// return false since this table cannot be easily categorized.
			idx := execute.ColIdx(label, key.Cols())
			if idx < 0 {
				return nil, false, nil
			}
			indices = append(indices, idx)
		}
	case flux.GroupModeExcept:
		indices = make([]int, 0, len(cols))
		for _, c := range cols {
			// If this string is part of except, then it is not included.
			if execute.ContainsStr(t.keys, c.Label) {
				continue
			}

			// If this column is not part of the group key, return false.
			idx := execute.ColIdx(c.Label, key.Cols())
			if idx < 0 {
				return nil, false, nil
			}
			indices = append(indices, idx)
		}
	default:
		panic(errors.Newf(codes.Internal, "unsupported group mode: %v", t.mode))
	}

	// Produce the group key from the indices.
	cols = make([]flux.ColMeta, len(indices))
	vs := make([]values.Value, len(indices))
	for j, idx := range indices {
		cols[j], vs[j] = key.Cols()[idx], key.Value(idx)
	}
	return execute.NewGroupKey(cols, vs), true, nil
}

func (t *groupTransformation) appendTable(ab *table.BufferedBuilder, tbl flux.Table) error {
	// Read the table and append each of the columns.
	return ab.AppendTable(tbl)
}

func (t *groupTransformation) groupChunkByRow(tbl table.Chunk, d *execute.TransportDataset, mem arrowmem.Allocator) error {
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
		return d.Process(table.ChunkFromBuffer(buf))
	})
}

// groupByRow will determine which table each row belongs to
// and to append them to that table.
func (t *groupTransformation) groupByRow(tbl flux.Table) error {
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
			return table.NewArrowBuilder(key, t.mem)
		},
	}
	if err := tbl.Do(func(cr flux.ColReader) error {
		for i, l := 0, cr.Len(); i < l; i++ {
			key := execute.GroupKeyForRowOn(i, cr, on)
			ab, created := table.GetArrowBuilder(key, &cache)
			if created {
				for _, c := range cr.Cols() {
					_, _ = ab.AddCol(c)
				}
			}
			for j := range cr.Cols() {
				if err := t.appendValueFromRow(ab.Builders[j], cr, i, j); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return cache.ForEach(func(key flux.GroupKey, builder table.Builder) error {
		tbl, err := builder.Table()
		if err != nil {
			return err
		}

		ab, _ := table.GetBufferedBuilder(key, &t.cache)
		return t.appendTable(ab, tbl)
	})
}

func (t *groupTransformation) appendValueFromRow(b array.Builder, cr flux.ColReader, i, j int) error {
	switch cr.Cols()[j].Type {
	case flux.TInt:
		b := b.(*array.IntBuilder)
		vs := cr.Ints(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TUInt:
		b := b.(*array.UintBuilder)
		vs := cr.UInts(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TFloat:
		b := b.(*array.FloatBuilder)
		vs := cr.Floats(j)
		if vs.IsNull(i) {
			b.AppendNull()
		} else {
			b.Append(vs.Value(i))
		}
	case flux.TString:
		b := b.(*array.StringBuilder)
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
		b := b.(*array.IntBuilder)
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

func (t *groupTransformation) UpdateWatermark(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateWatermark(ts)
}

func (t *groupTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *groupTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

// MergeGroupRule merges two group operations and keeps only the last one
type MergeGroupRule struct{}

func (r MergeGroupRule) Name() string {
	return "MergeGroupRule"
}

// Pattern returns the pattern that matches `group |> group`
func (r MergeGroupRule) Pattern() plan.Pattern {
	return plan.MultiSuccessor(GroupKind, plan.SingleSuccessor(GroupKind))
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
