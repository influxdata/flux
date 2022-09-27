package join

import (
	"context"
	"fmt"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// MergeJoinTransformation performs a sort-merge-join on two table streams.
// It assumes that input tables are sorted by the columns in the `on` parameter,
// and that we're performing an equijoin. The planner should ensure that both of
// these assumptions are true.
type MergeJoinTransformation struct {
	ctx         context.Context
	on          []ColumnPair
	as          *JoinFn
	left, right execute.DatasetID
	method      string
	d           *execute.TransportDataset
	mu          sync.Mutex
	mem         memory.Allocator

	// leftSchema and rightSchema keep track of a union of all the schemas
	// the join transformation has seen from each side. These are only used
	// when a group key on one side of a join does not exist on the other side
	// of a join. In that case, this information will be substituted in place of the
	// schema tracked by the `joinState` for that group key on that side, which is
	// necessary for compilation of the `as` function. (See the call to `Prepare`
	// inside of `joinState.join()` later in this file).
	leftSchema, rightSchema []flux.ColMeta

	leftFinished,
	rightFinished bool
}

func NewMergeJoinTransformation(
	ctx context.Context,
	id execute.DatasetID,
	s plan.ProcedureSpec,
	leftID execute.DatasetID,
	rightID execute.DatasetID,
	mem memory.Allocator,
) (*MergeJoinTransformation, error) {
	spec, ok := s.(*SortMergeJoinProcedureSpec)
	if !ok {
		return nil, errors.New(codes.Internal, "unsupported join spec - not a sortMergeJoin")
	}
	return &MergeJoinTransformation{
		ctx:    ctx,
		on:     spec.On,
		as:     NewJoinFn(spec.As),
		left:   leftID,
		right:  rightID,
		method: spec.Method,
		d:      execute.NewTransportDataset(id, mem),
		mem:    mem,
	}, nil
}

func (t *MergeJoinTransformation) Dataset() *execute.TransportDataset {
	return t.d
}

func (t *MergeJoinTransformation) ProcessMessage(m execute.Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer m.Ack()

	switch m := m.(type) {
	case execute.ProcessChunkMsg:
		chunk := m.TableChunk()
		id := m.SrcDatasetID()
		state, _ := t.d.Lookup(chunk.Key())
		s, ok, err := t.processChunk(chunk, state, id)
		if err != nil {
			return err
		}
		if ok {
			t.d.Set(chunk.Key(), s)
		}
	case execute.FlushKeyMsg:
		id := m.SrcDatasetID()
		key := m.Key()
		state, _ := t.d.Lookup(key)
		s, _ := state.(*joinState)
		if id == t.left {
			s.left.done = true
		} else if id == t.right {
			s.right.done = true
		}

		if s.finished() {
			if err := t.flush(s); err != nil {
				return err
			}
		}
	case execute.FinishMsg:
		err := m.Error()
		if err != nil {
			t.d.Finish(err)
			return nil
		}

		id := m.SrcDatasetID()
		if id == t.left {
			t.leftFinished = true
		} else if id == t.right {
			t.rightFinished = true
		}

		if t.isFinished() {
			err = t.d.Range(func(key flux.GroupKey, value interface{}) error {
				s, ok := value.(*joinState)
				if !ok {
					return errors.New(codes.Internal, "received bad joinState")
				}
				return t.flush(s)
			})
			t.d.Finish(err)
		}
	}
	return nil
}

func (t *MergeJoinTransformation) initState(state interface{}) (*joinState, bool) {
	if state != nil {
		s, ok := state.(*joinState)
		return s, ok
	}
	s := joinState{}
	return &s, true
}

// processChunk is where the bulk of the work for the join transformation happens.
//
// First, it initializes the state if it hasn't been initialized already. Then it sets the schema
// for whatever side of join state we're currently operating on.
//
// After that, it calls mergeJoin() on the passed-in chunk, which will do as much as it can to produce
// the joined output tables and update the state object accordingly.
func (t *MergeJoinTransformation) processChunk(chunk table.Chunk, state interface{}, id execute.DatasetID) (*joinState, bool, error) {
	s, ok := t.initState(state)
	if !ok {
		return nil, false, errors.New(codes.Internal, "invalid join state")
	}

	if chunk.Len() == 0 {
		return s, true, nil
	}
	chunk.Retain()

	var isLeft bool
	if id == t.left {
		isLeft = true
		s.left.schema = schemaUnion(s.left.schema, chunk.Cols())
		t.leftSchema = schemaUnion(t.leftSchema, chunk.Cols())
	} else if id == t.right {
		isLeft = false
		s.right.schema = schemaUnion(s.right.schema, chunk.Cols())
		t.rightSchema = schemaUnion(t.rightSchema, chunk.Cols())
	} else {
		return s, true, errors.New(codes.Internal, "invalid chunk passed to join - dataset id is neither left nor right")
	}

	if err := t.mergeJoin(chunk, s, isLeft); err != nil {
		return s, true, err
	}

	return s, true, nil
}

// mergeJoin takes a table chunk, and attempts to produce joined output from it
// if possible. It will follow these steps:
//
//  1. Scan the rows in `chunk` for a complete join key. If it finds the end
//     of a join key, it returns that key, along with all of the rows with that
//     join key. If it can't find a complete join key (which means it has reached
//     the end of the chunk), it will store the rows with that join key in the
//     `chunks` field of `sideState`, and return a nil joinKey, which is the signal
//     to break the mergeJoin loop.
//
//  2. If scanKey finds a complete join key, we pass it and the returned joinRows
//     into `insert()`, which will attempt to find the appropriate place for it
//     in the joinState's `products` field (while maintaining sort order). If
//     `insert()` detects that some subset of the stored products can be joined,
//     it will return `true` and a position. The position is the index into
//     s.products up to which it is safe to join.
//
//  3. For every product in s.products[:i] (inclusive on both ends), `join()`
//     will attempt to produce a set of table chunks that contains the joined
//     cross-product of the rows on each side of the joinProduct.
//
//  4. Pass each joined chunk onto the next node in the transformation.
//
//  5. Repeat each of the previous steps until every row in `chunk` has been scanned.
func (t *MergeJoinTransformation) mergeJoin(chunk table.Chunk, s *joinState, isLeft bool) error {
	for {
		key, rows, err := s.scanKey(chunk, isLeft, t.on)
		if err != nil {
			return err
		}
		if key == nil {
			break
		}
		i, canJoin := s.insert(key, rows, isLeft)
		if canJoin {
			joined, err := s.join(t.ctx, t.method, t.as, i, t.mem, t.leftSchema, t.rightSchema)
			if err != nil {
				return err
			}

			for _, chunk := range joined {
				err := t.d.Process(chunk)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// flush produces results from whatever remining data there is in the joinState.
func (t *MergeJoinTransformation) flush(s *joinState) error {
	// Get whatever rows are stored in the left and right sideStates, as well as their joinKey.
	// If mergeJoin() has done its job properly, the stored rows in a given side should all have
	// the same join key.
	lkey, lrows := s.left.flush(t.mem)

	// If there were any stored rows, insert them into s.products
	if lkey != nil {
		_, _ = s.insert(lkey, lrows, true)
	}

	rkey, rrows := s.right.flush(t.mem)
	if rkey != nil {
		_, _ = s.insert(rkey, rrows, false)
	}

	// Join everything in s.products
	joined, err := s.join(t.ctx, t.method, t.as, len(s.products)-1, t.mem, t.leftSchema, t.rightSchema)
	if err != nil {
		return err
	}

	// Pass any output along to the next transformation
	for _, chunk := range joined {
		err = t.d.Process(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *MergeJoinTransformation) isFinished() bool {
	return t.leftFinished && t.rightFinished
}

type joinState struct {
	left, right sideState
	products    []joinProduct
}

// scanKey passes the chunk to the appropriate side of the transformation, sets
// the join key columns if they're not already set, and returns the output of scan()
func (s *joinState) scanKey(c table.Chunk, isLeft bool, on []ColumnPair) (*joinKey, joinRows, error) {
	if isLeft {
		if len(s.left.joinKeyCols) < 1 {
			if err := s.left.setJoinKeyCols(getJoinKeyCols(on, isLeft), c); err != nil {
				return nil, nil, errors.Newf(codes.Invalid,
					"cannot set join columns in left table stream: %s", err)
			}
		}
		key, rows := s.left.scan(c)
		return key, rows, nil
	} else {
		if len(s.right.joinKeyCols) < 1 {
			if err := s.right.setJoinKeyCols(getJoinKeyCols(on, isLeft), c); err != nil {
				return nil, nil, errors.Newf(codes.Invalid,
					"cannot set join columns in right table stream: %s", err)
			}
		}
		key, rows := s.right.scan(c)
		return key, rows, nil
	}
}

func getJoinKeyCols(on []ColumnPair, isLeft bool) []string {
	labels := make([]string, 0, len(on))
	for _, pair := range on {
		if isLeft {
			labels = append(labels, pair.Left)
		} else {
			labels = append(labels, pair.Right)
		}
	}
	return labels
}

// Inserts `rows` into s.products, while maintaining sort order. Returns a position and a bool.
// If the bool == true, that means it's safe to join all the items in s.products up to and including
// the returned position.
//
// Returns true under 2 circumstances:
//
//	(1) Inserting `rows` completes the left and right pair for a given product
//	(2) `rows` was inserted at an index greater than 0, and all of the products that
//	     come before it only have entries on the opposite side.
//
// If condition 1 is true, we can join everything up to and including the index where
// `rows` was inserted.
//
// If condition 2 is true and condition 1 is false, we can join up to, but not including,
// the index where rows was inserted.
func (s *joinState) insert(key *joinKey, rows joinRows, isLeft bool) (int, bool) {
	if len(s.products) == 0 {
		p := newJoinProduct(key, rows, isLeft)
		s.products = []joinProduct{p}
		return 0, false
	}

	found := false
	position := 0
	for i, product := range s.products {
		if product.key.equal(*key) {
			if isLeft {
				if product.left != nil {
					panic(fmt.Sprintf(
						"join - joinProduct already has left value for key %s",
						product.key.str(),
					))
				}
				product.left = rows
			} else {
				if product.right != nil {
					panic(fmt.Sprintf(
						"join - joinProduct already has right value for key %s",
						product.key.str(),
					))
				}
				product.right = rows
			}
			s.products[i] = product
			found = true
			position = i
			break
		} else if key.less(product.key) {
			newProducts := make([]joinProduct, 0, len(s.products)+1)
			newProducts = append(newProducts, s.products[:i]...)
			newProducts = append(newProducts, newJoinProduct(key, rows, isLeft))
			newProducts = append(newProducts, s.products[i:]...)
			s.products = newProducts
			found = true
			position = i
			break
		}
		position = i
	}

	if !found {
		s.products = append(s.products, newJoinProduct(key, rows, isLeft))
		position = len(s.products) - 1
	}

	if s.products[position].isDone() {
		return position, true
	}

	// The only condition where we would not emit a signal to join and flush
	// the list of products is if we received a bunch of `joinRows` all
	// from the same side.
	//
	// So if the previous product is only populated on one side, then it is either the first
	// item in the list, or all of the items before it are also populated on the same side.
	canJoin := false
	if position > 0 {
		position--
		prev := s.products[position]
		if isLeft {
			canJoin = prev.left.len() == 0 && prev.right.len() > 0
		} else {
			canJoin = prev.left.len() > 0 && prev.right.len() == 0
		}
	}
	return position, canJoin
}

func (s *joinState) join(
	ctx context.Context,
	method string,
	fn *JoinFn,
	joinable int,
	mem memory.Allocator,
	defaultLeft, defaultRight []flux.ColMeta,
) ([]table.Chunk, error) {
	var lschema, rschema []flux.ColMeta
	if len(s.left.schema) > 0 {
		lschema = s.left.schema
	} else {
		lschema = defaultLeft
	}
	if len(s.right.schema) > 0 {
		rschema = s.right.schema
	} else {
		rschema = defaultRight
	}
	err := fn.Prepare(ctx, lschema, rschema)
	if err != nil {
		return nil, err
	}
	joined := make([]table.Chunk, 0, joinable+1)
	for i := 0; i <= joinable; i++ {
		prod := s.products[i]
		p, ok, err := prod.evaluate(ctx, method, *fn, mem)
		if err != nil {
			return joined, err
		}
		if !ok {
			continue
		}
		joined = append(joined, p...)
	}
	s.products = s.products[joinable+1:]
	return joined, nil
}

func (s *joinState) finished() bool {
	return s.left.done && s.right.done
}

type sideState struct {
	schema      []flux.ColMeta
	joinKeyCols []flux.ColMeta
	currentKey  joinKey
	chunks      []table.Chunk
	keyStart    int
	keyEnd      int
	done        bool
}

func (s *sideState) setJoinKeyCols(labels []string, c table.Chunk) error {
	cols := make([]flux.ColMeta, 0, len(labels))
	for _, label := range labels {
		colIdx := c.Index(label)
		if colIdx < 0 {
			return errors.Newf(codes.Invalid, "table is missing column '%s'", label)
		}
		cols = append(cols, c.Col(colIdx))
	}
	s.joinKeyCols = cols
	return nil
}

// scan calls and handles the outputs of advance(). If advance reports that it
// found a complete join key, scan returns the key, as well as a collection of the rows that match
// that join key. Otherwise, it stores the rows it just scanned in s.chunks
func (s *sideState) scan(c table.Chunk) (*joinKey, joinRows) {
	key, complete := s.advance(c)
	if !complete {
		s.addChunk(getChunkSlice(c, s.keyStart, c.Len()))
		return nil, nil
	}
	rows := s.consumeRows(c)
	return key, rows
}

func (s *sideState) addChunk(c table.Chunk) {
	s.chunks = append(s.chunks, c)
}

// advance iterates over each row of c until it either finds a new join key or
// reaches the end of the chunk. Returns true if it finds the end of a join key,
// along with the key itself.
func (s *sideState) advance(c table.Chunk) (*joinKey, bool) {
	var startKey joinKey
	if len(s.chunks) > 0 {
		startKey = s.currentKey
		s.keyStart = s.keyEnd
	} else {
		startKey = joinKeyFromRow(s.joinKeyCols, c, s.keyEnd)
		s.keyStart = s.keyEnd
		s.keyEnd++
	}

	complete := false
	for ; s.keyEnd < c.Len(); s.keyEnd++ {
		key := joinKeyFromRow(s.joinKeyCols, c, s.keyEnd)

		if !key.equal(startKey) {
			complete = true
			break
		}
	}

	// We hit the end of the chunk without finding a new join key.
	// Reset keyEnd to 0 since the next thing we process will be a new chunk.
	if !complete {
		s.currentKey = startKey
		s.keyEnd = 0
	}

	return &startKey, complete
}

// collects rows within the indices set by `advance()` and returns them in one
// data structure. It will discard any exhausted chunks or rows.
//
// Any chunks stored in s.chunks should have the same join key.
func (s *sideState) consumeRows(c table.Chunk) joinRows {
	rows := make([]table.Chunk, 0, len(s.chunks)+1)
	if len(s.chunks) > 0 {
		rows = append(rows, s.chunks...)
	}
	rows = append(rows, getChunkSlice(c, s.keyStart, s.keyEnd))
	s.chunks = []table.Chunk{}

	return rows
}

// flush returns any stored table chunks and their join key. It should not be possible for
// the returned chunks to have multiple join keys.
func (s *sideState) flush(mem memory.Allocator) (*joinKey, joinRows) {
	if len(s.chunks) < 1 {
		return nil, nil
	}
	key := joinKeyFromRow(s.joinKeyCols, s.chunks[0], 0)
	rows := joinRows(s.chunks)
	s.chunks = s.chunks[:0]
	return &key, rows
}

// Convenience/utility function to get a zero-copy slice of a table chunk
func getChunkSlice(chunk table.Chunk, start, stop int) table.Chunk {
	buf := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, 0, chunk.NCols()),
	}
	for _, col := range chunk.Buffer().Values {
		arr := arrow.Slice(col, int64(start), int64(stop))
		buf.Values = append(buf.Values, arr)
	}
	return table.ChunkFromBuffer(buf)
}

// joinRows represents a collection of rows from the same input table that all
// have the same join key.
type joinRows []table.Chunk

func (r joinRows) Release() {
	for _, chunk := range r {
		chunk.Release()
	}
}

func (r joinRows) len() int {
	return len(r)
}

func (r joinRows) nrows() int {
	nrows := 0
	for _, chunk := range r {
		nrows += chunk.Len()
	}
	return nrows
}

func (r joinRows) getRow(i int, typ semantic.MonoType) values.Object {
	var obj values.Object
	for _, chunk := range r {
		if i > chunk.Len()-1 {
			i -= chunk.Len()
		} else {
			obj = rowFromChunk(chunk, i, typ)
			break
		}
	}
	return obj
}

// joinProduct represents a collection of rows from the left and right
// input tables that have the same join key, and can therefore be joined
// in the final output table.
type joinProduct struct {
	key         joinKey
	left, right joinRows
}

func (p *joinProduct) Release() {
	p.left.Release()
	p.right.Release()
}

func newJoinProduct(key *joinKey, rows joinRows, isLeft bool) joinProduct {
	p := joinProduct{
		key: *key,
	}
	if isLeft {
		p.left = rows
	} else {
		p.right = rows
	}
	return p
}

func (p *joinProduct) isDone() bool {
	return p.left.len() > 0 && p.right.len() > 0
}

// Returns the joined output of the product, if there is any. If the returned bool is `true`,
// the returned table chunk contains joined data that should be passed along to the next node.
func (p *joinProduct) evaluate(ctx context.Context, method string, fn JoinFn, mem memory.Allocator) ([]table.Chunk, bool, error) {
	return fn.Eval(ctx, p, method, mem)
}

func rowFromChunk(c table.Chunk, i int, mt semantic.MonoType) values.Object {
	obj := values.NewObject(mt)
	buf := c.Buffer()
	for j := 0; j < c.NCols(); j++ {
		col := c.Col(j).Label
		v := execute.ValueForRow(&buf, i, j)
		obj.Set(col, v)
	}
	obj.Range(func(name string, v values.Value) {
		if v == nil {
			obj.Set(name, values.Null)
		}
	})
	return obj
}

func schemaUnion(a, b []flux.ColMeta) []flux.ColMeta {
	for _, bcol := range b {
		found := false
		for _, acol := range a {
			if bcol == acol {
				found = true
				break
			}
		}
		if !found {
			a = append(a, bcol)
		}
	}
	return a
}
