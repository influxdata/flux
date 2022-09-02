package experimental

import (
	"math"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const (
	DiffKind       = "experimental.diff"
	DefaultEpsilon = 1e-6
	DiffColumn     = "_diff"
)

func init() {
	diffSignature := runtime.MustLookupBuiltinType("experimental", "diff")

	runtime.RegisterPackageValue("experimental", "diff", flux.MustValue(flux.FunctionValue(DiffKind, createDiffOpSpec, diffSignature)))
	plan.RegisterProcedureSpec(DiffKind, newDiffProcedure, DiffKind)
	execute.RegisterTransformation(DiffKind, createDiffTransformation)
}

func createDiffOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	t, ok := args.Get("want")
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'want' not present")
	}
	p, ok := t.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "want input to diff is not a table object")
	}
	a.AddParent(p)

	t, ok = args.Get("got")
	if !ok {
		return nil, errors.New(codes.Invalid, "argument 'got' not present")
	}
	p, ok = t.(*flux.TableObject)
	if !ok {
		return nil, errors.New(codes.Invalid, "got input to diff is not a table object")
	}
	a.AddParent(p)

	return &DiffOpSpec{}, nil
}

type DiffOpSpec struct{}

func (s *DiffOpSpec) Kind() flux.OperationKind {
	return DiffKind
}

func newDiffProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*DiffOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &DiffProcedureSpec{}, nil
}

type DiffProcedureSpec struct {
	plan.DefaultCost
}

func (s *DiffProcedureSpec) Kind() plan.ProcedureKind {
	return DiffKind
}

func (s *DiffProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

func createDiffTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	if len(a.Parents()) != 2 {
		return nil, nil, errors.New(codes.Internal, "diff should have exactly 2 parents")
	}

	pspec, ok := spec.(*DiffProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", pspec)
	}

	wantID, gotID := a.Parents()[0], a.Parents()[1]
	return NewDiffTransformation(id, pspec, wantID, gotID, a.Allocator())
}

type diffTransformation struct {
	d        *execute.TransportDataset
	mem      memory.Allocator
	finished int
	err      error
	mu       sync.Mutex

	inputs        [2]*execute.RandomAccessGroupLookup
	wantID, gotID execute.DatasetID
}

func NewDiffTransformation(id execute.DatasetID, spec *DiffProcedureSpec, wantID, gotID execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &diffTransformation{
		d:   execute.NewTransportDataset(id, mem),
		mem: mem,
		inputs: [2]*execute.RandomAccessGroupLookup{
			execute.NewRandomAccessGroupLookup(),
			execute.NewRandomAccessGroupLookup(),
		},
		wantID: wantID,
		gotID:  gotID,
	}
	return execute.NewTransformationFromTransport(tr), tr.d, nil
}

func (d *diffTransformation) ProcessMessage(m execute.Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case execute.FinishMsg:
		d.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case execute.ProcessChunkMsg:
		return d.processChunk(m.SrcDatasetID(), m.TableChunk())
	case execute.FlushKeyMsg:
		return d.flushKey(m.SrcDatasetID(), m.Key())
	case execute.ProcessMsg:
		panic("unreachable")
	}
	return nil
}

func (d *diffTransformation) getInputs(id execute.DatasetID) *execute.RandomAccessGroupLookup {
	inputs := d.inputs[0]
	if d.gotID == id {
		inputs = d.inputs[1]
	}
	return inputs
}

func (d *diffTransformation) processChunk(id execute.DatasetID, chunk table.Chunk) error {
	// No lock needed for the inputs since we will only access data from one dataset.
	inputs := d.getInputs(id)
	chunks := inputs.LookupOrCreate(chunk.Key(), func() interface{} {
		return new([]table.Chunk)
	}).(*[]table.Chunk)
	chunk.Retain()
	*chunks = append(*chunks, chunk)
	return nil
}

func (d *diffTransformation) flushKey(id execute.DatasetID, key flux.GroupKey) error {
	inputs := d.getInputs(id)

	if state, ok := inputs.Delete(key); ok {
		d.mu.Lock()
		defer d.mu.Unlock()

		// Store the table chunks.
		chunks := state.(*[]table.Chunk)
		return d.computeDiffIfReady(id, key, *chunks)
	}
	return nil
}

func (d *diffTransformation) computeDiffIfReady(id execute.DatasetID, key flux.GroupKey, chunks []table.Chunk) error {
	state := d.d.LookupOrCreate(key, func() interface{} {
		return &diffTransformationState{}
	}).(*diffTransformationState)
	if d.gotID == id {
		state.got = chunks
	} else {
		state.want = chunks
	}
	state.finished++

	if state.finished < 2 {
		return nil
	}
	d.d.Delete(key)
	return d.computeDiff(key, state)
}

func (d *diffTransformation) computeDiff(key flux.GroupKey, state *diffTransformationState) error {
	defer state.Release()

	// Consolidate the chunks from want and got into a single table chunk.
	want, err := d.consolidate(state.want)
	if err != nil {
		return err
	}
	defer want.Release()

	got, err := d.consolidate(state.got)
	if err != nil {
		return err
	}
	defer got.Release()

	// Perform the diff on the consolidated chunks.
	diff, ok, err := d.diff(key, want, got)
	if err != nil || !ok {
		return err
	}

	if err := d.d.Process(diff); err != nil {
		return err
	}
	return d.d.FlushKey(key)
}

func (d *diffTransformation) consolidate(chunks []table.Chunk) (table.Chunk, error) {
	if len(chunks) == 0 {
		return table.Chunk{}, nil
	} else if len(chunks) == 1 {
		chunks[0].Retain()
		return chunks[0], nil
	}

	key := chunks[0].Key()
	builder := table.NewBufferedBuilder(key, d.mem)

	n := 0
	for _, chunk := range chunks {
		buf := chunk.Buffer()
		if err := builder.AppendBuffer(&buf); err != nil {
			return table.Chunk{}, err
		}
		n += buf.Len()
	}

	tbl, err := builder.Table()
	if err != nil {
		return table.Chunk{}, err
	}

	builders := make([]array.Builder, len(tbl.Cols()))
	for i, col := range tbl.Cols() {
		builders[i] = arrow.NewBuilder(col.Type, d.mem)
		builders[i].Resize(n)
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		for i := range cr.Cols() {
			arr := table.Values(cr, i)
			arrowutil.CopyTo(builders[i], arr)
		}
		return nil
	}); err != nil {
		return table.Chunk{}, err
	}

	buf := arrow.TableBuffer{
		GroupKey: chunks[0].Key(),
		Columns:  tbl.Cols(),
		Values:   make([]array.Array, len(builders)),
	}
	for i, b := range builders {
		buf.Values[i] = b.NewArray()
	}
	return table.ChunkFromBuffer(buf), nil
}

type diffSchema struct {
	cols            []flux.ColMeta
	offset          int
	want, got       table.Chunk
	wantIdx, gotIdx []int
}

func (d *diffSchema) equal(t *diffTransformation, i, j int) bool {
	for idx, col := range d.cols[d.offset:] {
		// Retrieve the arrays.
		var wantCol, gotCol array.Array
		if d.wantIdx[idx] >= 0 {
			wantCol = d.want.Values(d.wantIdx[idx])
		}

		if d.gotIdx[idx] >= 0 {
			gotCol = d.got.Values(d.gotIdx[idx])
		}

		// If one of the above is not present, then we're equal
		// if the other side is null. We do not have to check if both
		// are null because, if they were both null, neither entry would exist.
		if wantCol == nil {
			if gotCol.IsValid(j) {
				return false
			}
			continue
		} else if gotCol == nil {
			if wantCol.IsValid(i) {
				return false
			}
			continue
		}

		// Both columns are present, but we may still have to deal with nulls.
		if wantCol.IsValid(i) != gotCol.IsValid(j) {
			return false
		} else if wantCol.IsNull(i) {
			continue
		}

		// Both sides are valid because otherwise we would have skipped this column.
		// Compare the actual values.
		switch col.Type {
		case flux.TFloat:
			want, got := wantCol.(*array.Float).Value(i), gotCol.(*array.Float).Value(j)
			if math.IsNaN(want) && math.IsNaN(got) {
				// treat NaNs as equal so go to next column
				continue
			}
			if math.Abs(want-got) > DefaultEpsilon {
				return false
			}
		case flux.TInt:
			want, got := wantCol.(*array.Int), gotCol.(*array.Int)
			if want.Value(i) != got.Value(j) {
				return false
			}
		case flux.TUInt:
			want, got := wantCol.(*array.Uint), gotCol.(*array.Uint)
			if want.Value(i) != got.Value(j) {
				return false
			}
		case flux.TString:
			want, got := wantCol.(*array.String), gotCol.(*array.String)
			if want.Value(i) != got.Value(j) {
				return false
			}
		case flux.TBool:
			want, got := wantCol.(*array.Boolean), gotCol.(*array.Boolean)
			if want.Value(i) != got.Value(j) {
				return false
			}
		case flux.TTime:
			want, got := wantCol.(*array.Int), gotCol.(*array.Int)
			if want.Value(i) != got.Value(j) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (d *diffSchema) appendRow(builders []array.Builder, which, i int) {
	chunk, idxs := d.want, d.wantIdx
	if which > 0 {
		chunk, idxs = d.got, d.gotIdx
	}

	for j, idx := range idxs {
		builder := builders[j]
		if idx < 0 {
			builder.AppendNull()
			continue
		}

		arr := chunk.Values(idx)
		if arr.IsNull(i) {
			builder.AppendNull()
			continue
		}

		switch b := builder.(type) {
		case *array.FloatBuilder:
			b.Append(arr.(*array.Float).Value(i))
		case *array.IntBuilder:
			b.Append(arr.(*array.Int).Value(i))
		case *array.UintBuilder:
			b.Append(arr.(*array.Uint).Value(i))
		case *array.StringBuilder:
			b.Append(arr.(*array.String).Value(i))
		case *array.BooleanBuilder:
			b.Append(arr.(*array.Boolean).Value(i))
		default:
			b.AppendNull()
		}
	}
}

// createDiffSchema computes the output schema for the diff and maps the output schema
// to the columns in the original tables.
func (d *diffTransformation) createDiffSchema(key flux.GroupKey, want, got table.Chunk) (*diffSchema, error) {
	schema := &diffSchema{
		want: want,
		got:  got,
	}

	// Add the diff column first to place it in position zero.
	schema.cols = append(schema.cols, flux.ColMeta{
		Label: DiffColumn,
		Type:  flux.TString,
	})

	// While we could overwrite the diff column, that isn't really possible if
	// it's in the group key because it would meaningfully change the narrow property.
	if key.HasCol(DiffColumn) {
		return nil, errors.New(codes.FailedPrecondition, "group key cannot contain _diff column")
	}
	// Append all columns present in the key.
	schema.cols = append(schema.cols, key.Cols()...)
	// Mark where the comparison columns will start.
	schema.offset = len(schema.cols)

	// Add all columns from the want side of the table.
	for i, col := range want.Cols() {
		// Drop _diff columns as we will overwrite it.
		if col.Label == DiffColumn {
			continue
		}

		if !execute.HasCol(col.Label, schema.cols) {
			schema.cols = append(schema.cols, col)
			schema.wantIdx = append(schema.wantIdx, i)

			gotIdx := execute.ColIdx(col.Label, got.Cols())
			if gotIdx >= 0 {
				if gotType := got.Col(gotIdx).Type; gotType != col.Type {
					return nil, errors.Newf(codes.FailedPrecondition, "column %q has different types %s != %s", col.Label, col.Type, gotType)
				}
			}
			schema.gotIdx = append(schema.gotIdx, execute.ColIdx(col.Label, got.Cols()))
		}
	}

	// Add any missing columns from the got side. If one side has a column and
	// the other doesn't, this essentially means that all columns are going to be
	// unequal.
	for i, col := range got.Cols() {
		// Drop _diff columns as we will overwrite it.
		if col.Label == DiffColumn {
			continue
		}

		if !execute.HasCol(col.Label, schema.cols) {
			schema.cols = append(schema.cols, col)
			schema.wantIdx = append(schema.wantIdx, -1)
			schema.gotIdx = append(schema.gotIdx, i)
		}
	}
	return schema, nil
}

// diff computes a diff from want and got.
//
// If the two tables are identical, this will return false to indicate that no diff
// was computed. If the tables are not identical, a table chunk with the differences
// between the two tables will be created. This diff will also include context for the
// rows that are the same. Presently, this context is for the entire table so all
// rows will be represented in some way.
func (d *diffTransformation) diff(key flux.GroupKey, want, got table.Chunk) (table.Chunk, bool, error) {
	// Compute a schema to determine the output columns and map those columns
	// to their locations in the original tables.
	schema, err := d.createDiffSchema(key, want, got)
	if err != nil {
		return table.Chunk{}, false, err
	}

	// Check if this was a perfect match. The lengths have to be the same.
	if want.Len() == got.Len() {
		isPerfectMatch := true
		for i, n := 0, want.Len(); i < n; i++ {
			if !schema.equal(d, i, i) {
				isPerfectMatch = false
				break
			}
		}

		if isPerfectMatch {
			return table.Chunk{}, false, nil
		}
	}

	// At this point, we are producing a diff in some manner and
	// we expect to find some difference.
	diff := array.NewStringBuilder(d.mem)
	builders := make([]array.Builder, len(schema.wantIdx))
	for i, col := range schema.cols[schema.offset:] {
		builders[i] = arrow.NewBuilder(col.Type, d.mem)
	}

	// Compute the lcs (longest common subsequence) table.
	lcs := d.lcs(schema)
	tracer := diffTrace{
		lcs:      lcs,
		schema:   schema,
		diff:     diff,
		builders: builders,
	}
	// Trace the table from the last element.
	tracer.trace(len(lcs[0])-1, len(lcs)-1)

	// Create the diff table.
	buf := arrow.TableBuffer{
		GroupKey: key,
		Columns:  schema.cols,
		Values:   make([]array.Array, 0, len(schema.cols)),
	}
	// `diff.NewArray()` will move the array out of `diff` and thereby zero the length so we read it here for later use
	diffLen := diff.Len()
	buf.Values = append(buf.Values, diff.NewArray())
	for i, col := range key.Cols() {
		arr := arrow.Repeat(col.Type, key.Value(i), diffLen, d.mem)
		buf.Values = append(buf.Values, arr)
	}
	for _, b := range builders {
		buf.Values = append(buf.Values, b.NewArray())
	}
	return table.ChunkFromBuffer(buf), true, nil
}

type diffTrace struct {
	lcs      [][]diffLcsEntry
	schema   *diffSchema
	diff     *array.StringBuilder
	builders []array.Builder
}

func (d *diffTrace) trace(row, col int) {
	parent := d.lcs[col][row].parent
	if row > 0 && col > 0 && parent == diffParentMatch {
		d.trace(row-1, col-1)
		d.diff.Append("")
		d.schema.appendRow(d.builders, 0, row-1)
	} else if row > 0 && (col == 0 || parent == diffParentLeft) {
		d.trace(row-1, col)
		d.diff.Append("-")
		d.schema.appendRow(d.builders, 0, row-1)
	} else if col > 0 && (row == 0 || parent == diffParentTop || parent == diffParentEither) {
		d.trace(row, col-1)
		d.diff.Append("+")
		d.schema.appendRow(d.builders, 1, col-1)
	}
}

type diffLcsParent int

const (
	diffParentEither diffLcsParent = iota
	diffParentLeft
	diffParentTop
	diffParentMatch
)

type diffLcsEntry struct {
	parent diffLcsParent
	length int
}

// lcs computes a longest common subsequence table.
//
// This constructs a table like the one present here: https://en.wikipedia.org/wiki/Longest_common_subsequence_problem#Traceback_approach.
//
// The table is constructed with the row dimension corresponding to the want
// and the column dimension corresponding to the got.
//
// The table itself is accessed with table[col][row]. The table itself
// is also 1 indexed instead of zero where zero is the base case. That means the first
// element in the table chunks corresponds to index 1 in this table.
//
// See the example in the wikipedia article for details.
func (d *diffTransformation) lcs(schema *diffSchema) [][]diffLcsEntry {
	// We store the lcs entries in the table with the row being the want
	// and the columns being the got.
	lcsTable := make([][]diffLcsEntry, schema.got.Len()+1)
	for i := range lcsTable {
		lcsTable[i] = make([]diffLcsEntry, schema.want.Len()+1)
	}

	// This algorithm uses the naive approach to computing this table.
	for col := range lcsTable {
		// Zero index is the empty set so skip it.
		if col == 0 {
			continue
		}

		for row := range lcsTable[col] {
			if row == 0 {
				continue
			}

			match := schema.equal(d, row-1, col-1)
			if match {
				lcsTable[col][row] = diffLcsEntry{
					parent: diffParentMatch,
					length: lcsTable[col-1][row-1].length + 1,
				}
			} else {
				top := lcsTable[col-1][row].length
				left := lcsTable[col][row-1].length
				if top == left {
					lcsTable[col][row] = diffLcsEntry{
						parent: diffParentEither,
						length: top,
					}
				} else if top > left {
					lcsTable[col][row] = diffLcsEntry{
						parent: diffParentTop,
						length: top,
					}
				} else {
					lcsTable[col][row] = diffLcsEntry{
						parent: diffParentLeft,
						length: left,
					}
				}
			}
		}
	}
	return lcsTable
}

func (d *diffTransformation) Finish(id execute.DatasetID, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.err != nil {
		err = d.err
	}

	inputs := d.getInputs(id)
	if err == nil {
		// If no error occurred, we flush all keys
		err = inputs.Range(func(key flux.GroupKey, value interface{}) error {
			chunks := value.(*[]table.Chunk)
			return d.computeDiffIfReady(id, key, *chunks)
		})
	}

	if err != nil {
		_ = inputs.Range(func(key flux.GroupKey, value interface{}) error {
			chunks := value.(*[]table.Chunk)
			for _, chunk := range *chunks {
				chunk.Release()
			}
			return nil
		})
	}

	// This parent is finished. Mark it so.
	d.finished++
	inputs.Clear()

	// Store any error for future iterations.
	d.err = err

	// We are now handling global state.
	if d.finished < 2 {
		return
	}
	d.finish(err)
}

func (d *diffTransformation) finish(err error) {
	// Flush all keys that do not have a partner table.
	if err == nil {
		err = d.d.Range(func(key flux.GroupKey, value interface{}) error {
			state := value.(*diffTransformationState)
			return d.computeDiff(key, state)
		})
	}

	// If an error occurred, just clear the values stored in the global state.
	if err != nil {
		_ = d.d.Range(func(key flux.GroupKey, value interface{}) error {
			state := value.(*diffTransformationState)
			state.Release()
			return nil
		})
	}
	d.d.Finish(err)
}

type diffTransformationState struct {
	want, got []table.Chunk
	finished  int
}

func (d *diffTransformationState) Release() {
	for _, chunk := range d.want {
		chunk.Release()
	}
	for _, chunk := range d.got {
		chunk.Release()
	}
	d.want, d.got = nil, nil
}
