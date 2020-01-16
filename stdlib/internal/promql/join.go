package promql

import (
	"context"
	"sort"
	"sync"

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

const joinKind = "internal/promql.join"

func init() {
	signature := semantic.MustLookupBuiltinType("internal/promql", "join")
	flux.RegisterPackageValue("internal/promql", "join", flux.MustValue(flux.FunctionValue("join", createJoinOpSpec, signature)))
	flux.RegisterOpSpec(joinKind, newJoinOp)
	plan.RegisterProcedureSpec(joinKind, newMergeJoinProcedure, joinKind)
	execute.RegisterTransformation(joinKind, createMergeJoinTransformation)
}

type JoinOpSpec struct {
	Left  flux.OperationID             `json:"left"`
	Right flux.OperationID             `json:"right"`
	Fn    interpreter.ResolvedFunction `json:"fn"`

	l, r *flux.TableObject
}

func (s *JoinOpSpec) IDer(ider flux.IDer) {
	s.Left = ider.ID(s.l)
	s.Right = ider.ID(s.r)
}

func createJoinOpSpec(args flux.Arguments, p *flux.Administration) (flux.OperationSpec, error) {
	l, err := args.GetRequiredObject("left")
	if err != nil {
		return nil, err
	}

	// TODO(josh): The type system should ensure that this
	// assertion is redundant. Unfortunately it does not.
	// Remove this check when type inference is fixed.
	//
	left, ok := l.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'left' must be a table stream")
	}
	p.AddParent(left)

	r, err := args.GetRequiredObject("right")
	if err != nil {
		return nil, err
	}

	// Same comment as above. The type system should ensure
	// that the folowing cast never panics.
	//
	right, ok := r.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'right' must be a table stream")
	}
	p.AddParent(right)

	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	fn, err := interpreter.ResolveFunction(f)
	if err != nil {
		return nil, err
	}

	return &JoinOpSpec{
		Fn: fn,
		l:  left,
		r:  right,
	}, nil
}

func newJoinOp() flux.OperationSpec {
	return new(JoinOpSpec)
}

func (s *JoinOpSpec) Kind() flux.OperationKind {
	return joinKind
}

type MergeJoinProcedureSpec struct {
	plan.DefaultCost

	Fn interpreter.ResolvedFunction `json:"fn"`
}

func newMergeJoinProcedure(spec flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := spec.(*JoinOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return &MergeJoinProcedureSpec{Fn: s.Fn}, nil
}

func (s *MergeJoinProcedureSpec) Kind() plan.ProcedureKind {
	return joinKind
}
func (s *MergeJoinProcedureSpec) Copy() plan.ProcedureSpec {
	return &MergeJoinProcedureSpec{Fn: s.Fn.Copy()}
}

func createMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MergeJoinProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	parents := a.Parents()

	c := NewMergeJoinCache(a.Context(), a.Allocator(), s.Fn, parents[0], parents[1])
	d := execute.NewDataset(id, mode, c)
	t := NewMergeJoinTransformation(d, c)
	return t, d, nil
}

type mergeJoinTransformation struct {
	mu    sync.Mutex
	d     execute.Dataset
	cache *mergeJoinCache
	done  bool
}

func NewMergeJoinTransformation(d execute.Dataset, cache *mergeJoinCache) *mergeJoinTransformation {
	return &mergeJoinTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *mergeJoinTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	columns := tbl.Cols()

	timeCol := execute.ColIdx(execute.DefaultTimeColLabel, columns)
	if timeCol == -1 {
		return errors.New(codes.Invalid, "no _time column found")
	}

	var readers []flux.ColReader
	if err := tbl.Do(func(cr flux.ColReader) error {
		cr.Retain()
		readers = append(readers, cr)
		return nil
	}); err != nil {
		return err
	}

	t.cache.insert(id, tbl.Key(), NewRowIterator(columns, readers, timeCol))
	return nil
}

func (t *mergeJoinTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return errors.New(codes.Unimplemented)
}

func (t *mergeJoinTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.d.UpdateWatermark(mark)
}

func (t *mergeJoinTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.d.UpdateProcessingTime(pt)
}

func (t *mergeJoinTransformation) Finish(id execute.DatasetID, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err != nil || t.done {
		t.d.Finish(err)
		t.cache.clean()
	}

	t.done = true
}

func NewMergeJoinCache(ctx context.Context, alloc *memory.Allocator, fn interpreter.ResolvedFunction, left, right execute.DatasetID) *mergeJoinCache {
	return &mergeJoinCache{
		left:  left,
		right: right,
		fn:    newRowJoinFn(fn.Fn, compiler.ToScope(fn.Scope)),
		data:  execute.NewGroupLookup(),
		ctx:   ctx,
		alloc: alloc,
	}
}

type mergeJoinCache struct {
	left, right execute.DatasetID
	fn          *rowJoinFn

	data *execute.GroupLookup
	spec plan.TriggerSpec

	ctx   context.Context
	alloc *memory.Allocator
}

type cacheEntry struct {
	l, r *RowIterator
}

func (e *cacheEntry) ready() bool {
	return e.l != nil && e.r != nil && e.l.len != 0 && e.r.len != 0
}

func (c *mergeJoinCache) insert(id execute.DatasetID, key flux.GroupKey, iter *RowIterator) {
	if entry, ok := c.data.Lookup(key); ok {
		switch id {
		case c.left:
			entry.(*cacheEntry).l = iter
		case c.right:
			entry.(*cacheEntry).r = iter
		}
	} else {
		switch id {
		case c.left:
			c.data.Set(key, &cacheEntry{l: iter})
		case c.right:
			c.data.Set(key, &cacheEntry{r: iter})
		}
	}
}

func (c *mergeJoinCache) delete(key flux.GroupKey) {
	if entry, ok := c.data.Delete(key); ok {
		t := entry.(*cacheEntry)
		if t.l != nil {
			for _, reader := range t.l.readers {
				reader.Release()
			}
		}
		if t.r != nil {
			for _, reader := range t.r.readers {
				reader.Release()
			}
		}
	}
}

func (c *mergeJoinCache) clean() {
	var keys []flux.GroupKey
	c.data.Range(func(key flux.GroupKey, value interface{}) {
		keys = append(keys, key)
	})
	for _, key := range keys {
		c.delete(key)
	}
}

func (c *mergeJoinCache) Table(key flux.GroupKey) (flux.Table, error) {
	entry, ok := c.data.Lookup(key)
	if !ok {
		return nil, errors.Newf(codes.Internal, "no entry for group key %v in cache", key)
	}
	t := entry.(*cacheEntry)
	if t.l == nil || t.r == nil {
		return nil, errors.Newf(codes.Internal, "no entry for group key %v in cache", key)
	}
	return c.join(key, t.l, t.r)
}

func (c *mergeJoinCache) ForEach(f func(flux.GroupKey)) {
	c.data.Range(func(key flux.GroupKey, value interface{}) {
		if value.(*cacheEntry).ready() {
			f(key)
		}
	})
}

func (c *mergeJoinCache) ForEachWithContext(f func(flux.GroupKey, execute.Trigger, execute.TableContext)) {
	c.data.Range(func(key flux.GroupKey, value interface{}) {
		if value.(*cacheEntry).ready() {
			f(key, execute.NewTriggerFromSpec(c.spec), execute.TableContext{
				Key: key,
			})
		}
	})
}

func (c *mergeJoinCache) DiscardTable(key flux.GroupKey) {
	c.delete(key)
}

func (c *mergeJoinCache) ExpireTable(key flux.GroupKey) {
	c.delete(key)
}

func (c *mergeJoinCache) SetTriggerSpec(spec plan.TriggerSpec) {
	c.spec = spec
}

func (c *mergeJoinCache) join(key flux.GroupKey, a, b *RowIterator) (flux.Table, error) {
	// Compile row fn for the input rows
	if err := c.fn.Prepare(a.columns, b.columns); err != nil {
		return nil, err
	}

	// Create a table builder for the output of the join
	builder := execute.NewColListTableBuilder(key, c.alloc)

	firstRow := true
	i, j := 0, 0

NEXT:
	ta := a.time(i)
	tb := b.time(j)

	if ta == -1 || tb == -1 {
		goto DONE
	}

	if ta < tb {
		i++
		goto NEXT
	}

	if ta > tb {
		j++
		goto NEXT
	}

	// There may be multiple rows of b that match with single row of a.
	// The following loop joins all such rows.
	//
	// Note there may be multiple rows of a that match with a single
	// row of b. This is accounted for as the current index of b (j)
	// is reset after the loop, while the current index of a (i) is
	// incremented by one.
	//
	for k := j; ta == b.time(k); k++ {

		// Evaluate fn over both input rows
		obj, err := c.fn.Eval(c.ctx, a.record(i), b.record(k))
		if err != nil {
			return nil, err
		}

		// Build schema if this is the first row being joined
		if firstRow {
			if err := buildSchema(builder, obj); err != nil {
				return nil, err
			}
			firstRow = false
		}

		// Check fn does not update the group key values.
		// TODO(josh): This should be caught during planning.
		// Remove this when the planner is made schema aware.
		if ok := objContainsKey(obj, key); !ok {
			return nil, errors.New(codes.Invalid, "argument 'fn' may not modify group key")
		}

		// The record obtained from calling fn may be added to output
		if err := appendRowToBuilder(builder, obj); err != nil {
			return nil, err
		}
	}
	i++
	goto NEXT
DONE:
	return builder.Table()
}

// objContainsKey checks if an object contains a specific group key
func objContainsKey(obj values.Object, key flux.GroupKey) bool {
	for _, col := range key.Cols() {
		if v, ok := obj.Get(col.Label); !ok || !v.Equal(key.LabelValue(col.Label)) {
			return false
		}
	}
	return true
}

// buildSchema adds a schema defined by an object to an empty builder
func buildSchema(builder *execute.ColListTableBuilder, obj values.Object) error {
	schema := make([]flux.ColMeta, 0, obj.Len())
	obj.Range(func(name string, v values.Value) {
		schema = append(schema, flux.ColMeta{
			Label: name,
			Type:  execute.ConvertFromKind(v.Type().Nature()),
		})
	})
	sort.Slice(schema, func(i, j int) bool {
		return schema[i].Label < schema[j].Label
	})
	for _, col := range schema {
		if _, err := builder.AddCol(col); err != nil {
			return err
		}
	}
	return nil
}

func appendRowToBuilder(builder *execute.ColListTableBuilder, obj values.Object) error {
	var err error
	obj.Range(func(name string, v values.Value) {
		idx := execute.ColIdx(name, builder.Cols())
		if idx < 0 {
			err = errors.Newf(codes.NotFound, "column %s not found", name)
			return
		}
		if err = builder.AppendValue(idx, v); err != nil {
			return
		}
	})
	return err
}

func NewRowIterator(columns []flux.ColMeta, readers []flux.ColReader, timeCol int) *RowIterator {
	offsets, l := make([]int, len(readers)), 0
	for i, r := range readers {
		offsets[i] = l
		l += r.Len()
	}
	return &RowIterator{
		row:     make(map[string]values.Value),
		len:     l,
		columns: columns,
		readers: readers,
		offsets: offsets,
		timeCol: timeCol,
	}
}

// RowIterator iterates over the rows of several column readers
type RowIterator struct {
	len int
	row map[string]values.Value

	timeCol int
	offsets []int
	readers []flux.ColReader
	columns []flux.ColMeta
}

// time returns the time at index idx
func (iter *RowIterator) time(idx int) int64 {
	for i := len(iter.readers) - 1; i >= 0; i-- {
		o := iter.offsets[i]
		r := iter.readers[i]
		if idx >= o {
			if idx-o >= r.Len() {
				return -1
			}
			return r.Times(iter.timeCol).Value(idx - o)
		}
	}
	return -1
}

// record returns the row at index idx
func (iter *RowIterator) record(idx int) map[string]values.Value {
	for k := range iter.row {
		delete(iter.row, k)
	}
	for i := len(iter.readers) - 1; i >= 0; i-- {
		o := iter.offsets[i]
		r := iter.readers[i]
		if idx >= o {
			for j, col := range r.Cols() {
				iter.row[col.Label] = execute.ValueForRow(r, idx-o, j)
			}
			break
		}
	}
	return iter.row
}

// rowJoinFn is equivalent to the lambda function (a, b) => ...
// Parameters a and b as well as the return value are all record types.
type rowJoinFn struct {
	fn         *semantic.FunctionExpression
	scope      compiler.Scope
	cache      *compiler.CompilationCache
	preparedFn compiler.Func
}

func newRowJoinFn(fn *semantic.FunctionExpression, scope compiler.Scope) *rowJoinFn {
	return &rowJoinFn{
		fn:    fn,
		scope: scope,
		cache: compiler.NewCompilationCache(fn, scope),
	}
}

func (fn *rowJoinFn) Prepare(left, right []flux.ColMeta) error {
	l := make(map[string]semantic.Nature, len(left))
	for _, col := range left {
		l[col.Label] = execute.ConvertToKind(col.Type)
	}

	r := make(map[string]semantic.Nature, len(right))
	for _, col := range right {
		r[col.Label] = execute.ConvertToKind(col.Type)
	}

	f, err := fn.cache.Compile(semantic.NewObjectType(
	// TODO (algow): determine the correct type
	//map[string]semantic.MonoType{
	//	"left":  semantic.NewObjectType(l),
	//	"right": semantic.NewObjectType(r),
	//},
	))
	if err != nil {
		return err
	}
	fn.preparedFn = f
	return nil
}

func (fn *rowJoinFn) Eval(ctx context.Context, left, right map[string]values.Value) (values.Object, error) {
	obj, err := fn.preparedFn.Eval(ctx, values.NewObjectWithValues(map[string]values.Value{
		"left":  values.NewObjectWithValues(left),
		"right": values.NewObjectWithValues(right),
	}))
	if err != nil {
		return nil, err
	}
	return obj.Object(), nil
}
