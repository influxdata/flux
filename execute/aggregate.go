package execute

import (
	"context"
	"sync"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/feature"
	fluxmemory "github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
)

// AggregateTransformation implements a transformation that aggregates
// the results from multiple TableView values and then outputs a Table
// with the same group key.
//
// This is similar to NarrowTransformation that it does not modify the group key,
// but different because it will only output a table when the key is flushed.
type AggregateTransformation interface {
	// Aggregate will process the table.Chunk with the state from the previous
	// time a table with this group key was invoked.
	//
	// If this group key has never been invoked before, the state will be nil.
	//
	// The transformation should return the new state and a boolean
	// value of true if the state was created or modified. If false is returned,
	// the new state will be discarded and any old state will be kept.
	//
	// It is ok for the transformation to modify the state if it is
	// a pointer. This is both allowed and recommended.
	Aggregate(chunk table.Chunk, state interface{}, mem memory.Allocator) (interface{}, bool, error)

	// Compute will signal the AggregateTransformation to compute
	// the aggregate for the given key from the provided state.
	//
	// The state will be the value that was returned from Aggregate.
	// If the Aggregate function never returned state, this function
	// will never be called.
	Compute(key flux.GroupKey, state interface{}, d *TransportDataset, mem memory.Allocator) error

	Closer
}

type aggregateTransformation struct {
	t AggregateTransformation
	d *TransportDataset
}

// NewAggregateTransformation constructs a Transformation and Dataset
// using the aggregateTransformation implementation.
func NewAggregateTransformation(id DatasetID, t AggregateTransformation, mem memory.Allocator) (Transformation, Dataset, error) {
	tr := &aggregateTransformation{
		t: t,
		d: NewTransportDataset(id, mem),
	}
	return NewTransformationFromTransport(tr), tr.d, nil
}

// ProcessMessage will process the incoming message.
func (t *aggregateTransformation) ProcessMessage(m Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case FinishMsg:
		t.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessChunkMsg:
		return t.processChunk(m.TableChunk())
	case FlushKeyMsg:
		return t.flushKey(m.Key())
	case ProcessMsg:
		panic("unreachable")
	}
	return nil
}

func (t *aggregateTransformation) processChunk(chunk table.Chunk) error {
	state, _ := t.d.Lookup(chunk.Key())
	if newState, ok, err := t.t.Aggregate(chunk, state, t.d.mem); err != nil {
		return err
	} else if ok {
		// Associate the newly returned state with the group key
		// if we were told to do so by Aggregate.
		t.d.Set(chunk.Key(), newState)
	}
	return nil
}

func (t *aggregateTransformation) computeFor(key flux.GroupKey, state interface{}) error {
	// Special code to allow aggregate parallel to work without copying this code.
	// Unwrap the state from the aggregate parallel state automatically.
	if s, ok := state.(*aggregateParallelState); ok {
		state = s.state
	}

	if err := t.t.Compute(key, state, t.d, t.d.mem); err != nil {
		return err
	}

	// If this state is disposable, we are done with it so invoke
	// the Close method.
	if v, ok := state.(Closer); ok {
		if err := v.Close(); err != nil {
			return err
		}
	}
	return t.d.FlushKey(key)
}

func (t *aggregateTransformation) flushKey(key flux.GroupKey) error {
	// Remove the state for this key from the dataset.
	// If we find state associated with the key, compute the table.
	if state, ok := t.d.Delete(key); ok {
		return t.computeFor(key, state)
	}
	return nil
}

// Finish is implemented to remain compatible with legacy upstreams.
func (t *aggregateTransformation) Finish(id DatasetID, err error) {
	if err == nil {
		err = t.d.Range(func(key flux.GroupKey, value interface{}) error {
			return t.computeFor(key, value)
		})
	}

	// If an error occurred, close all items in the global state.
	// This happens automatically in computeFor for the successful case,
	// but an error can cause computeFor not to be called on all states.
	if err != nil {
		_ = t.d.Range(func(key flux.GroupKey, value interface{}) error {
			if v, ok := value.(Closer); ok {
				_ = v.Close()
			}
			return nil
		})
	}

	err = Close(err, t.t)
	t.d.Finish(err)
}

func (t *aggregateTransformation) OperationType() string {
	return OperationType(t.t)
}

// AggregateParallelTransformation is an AggregateTransformation that is capable of
// processing chunks from within the same group key in parallel.
//
// The thing that differentiates this from a normal AggregateTransformation is having
// multiple parents and the capability to merge two existing states into a single one
// that will be passed to Compute.
type AggregateParallelTransformation interface {
	AggregateTransformation

	// Merge will take two existing states produced by the Aggregate method and merge them
	// into a single state.
	Merge(into, from interface{}, mem memory.Allocator) (interface{}, error)
}

type aggregateParallelTransformation struct {
	aggregateTransformation

	merge    func(into, from interface{}, mem memory.Allocator) (interface{}, error)
	parents  map[DatasetID]*RandomAccessGroupLookup
	finished int
	err      error
	mu       sync.RWMutex
}

// NewAggregateParallelTransformation constructs a Transformation and Dataset
// using the AggregateParallelTransformation implementation.
func NewAggregateParallelTransformation(id DatasetID, parents []DatasetID, t AggregateParallelTransformation, mem memory.Allocator) (Transformation, Dataset, error) {
	if len(parents) == 1 {
		return NewAggregateTransformation(id, t, mem)
	}

	tr := &aggregateParallelTransformation{
		aggregateTransformation: aggregateTransformation{
			t: t,
			d: NewTransportDataset(id, mem),
		},
		merge:   t.Merge,
		parents: make(map[DatasetID]*RandomAccessGroupLookup, len(parents)),
	}
	for _, parent := range parents {
		tr.parents[parent] = NewRandomAccessGroupLookup()
	}
	return NewTransformationFromTransport(tr), tr.d, nil
}

// ProcessMessage will process the incoming message.
func (t *aggregateParallelTransformation) ProcessMessage(m Message) error {
	defer m.Ack()

	// Finish messages always go through.
	if m, ok := m.(FinishMsg); ok {
		t.Finish(m.SrcDatasetID(), m.Error())
		return nil
	}

	// Do not invoke the other methods if an error has previously happened.
	if t.hasError() {
		return nil
	}

	switch m := m.(type) {
	case ProcessChunkMsg:
		return t.processChunk(m.SrcDatasetID(), m.TableChunk())
	case FlushKeyMsg:
		return t.flushKey(m.SrcDatasetID(), m.Key())
	case ProcessMsg:
		panic("unreachable")
	}
	return nil
}

// hasError returns true if an error has happened from some dataset.
func (t *aggregateParallelTransformation) hasError() bool {
	t.mu.Lock()
	hasErr := t.err != nil
	t.mu.Unlock()
	return hasErr
}

func (t *aggregateParallelTransformation) processChunk(parent DatasetID, chunk table.Chunk) error {
	d := t.parents[parent]

	state, _ := d.Lookup(chunk.Key())
	if newState, ok, err := t.t.Aggregate(chunk, state, t.d.mem); err != nil {
		return err
	} else if ok {
		// Associate the newly returned state with the group key
		// if we were told to do so by Aggregate.
		d.Set(chunk.Key(), newState)
	}
	return nil
}

func (t *aggregateParallelTransformation) mergeState(key flux.GroupKey, from interface{}) error {
	merged := t.d.LookupOrCreate(key, func() interface{} {
		return &aggregateParallelState{}
	}).(*aggregateParallelState)

	if merged.state == nil {
		// No existing state for this merge. Just take the passed in
		// state. This works fine even if both sides are nil.
		merged.state = from
	} else if from != nil {
		// We have merged state and new state.
		// Merge these together with the merge function.
		mergedState, err := t.merge(merged.state, from, t.d.mem)
		if err != nil {
			return err
		}
		merged.state = mergedState

		// If the from state was disposable, close it now.
		if v, ok := from.(Closer); ok {
			if err := v.Close(); err != nil {
				return err
			}
		}
	}

	// Increment the count regardless of what happens.
	merged.count++
	// Are we done with this group key? Flush the key if we are.
	if done := merged.count == len(t.parents); done {
		return t.aggregateTransformation.flushKey(key)
	}
	return nil
}

func (t *aggregateParallelTransformation) flushKey(parent DatasetID, key flux.GroupKey) error {
	d := t.parents[parent]

	// Remove the state for this key from the dataset.
	// If we find state associated with the key, compute the table.
	if state, ok := d.Delete(key); ok {
		// We will be interacting with shared state so lock the mutex.
		t.mu.Lock()
		defer t.mu.Unlock()

		// Merge this state with the others.
		return t.mergeState(key, state)
	}
	return nil
}

// Finish is implemented to remain compatible with legacy upstreams.
func (t *aggregateParallelTransformation) Finish(id DatasetID, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Restore the error if one happened previously.
	// This allows remaining datasets to know that an error
	// was already reported.
	//
	// This really only applies to multiple parent situations
	// since this method is only called once with a single
	// parent so there's nothing to restore.
	if t.err != nil {
		err = t.err
	}

	// Retrieve the parent dataset.
	d := t.parents[id]
	if err == nil {
		// If no error occurred, we need to flush all keys from the parent
		// dataset into the global one.
		err = d.Range(func(key flux.GroupKey, value interface{}) error {
			return t.mergeState(key, value)
		})
	}

	// If an error occurred, possibly in the previous section,
	// we need to clear the state from this parent.
	// If no error happened, then mergeState should have already done
	// this for us. Calling close multiple times should be fine.
	if err != nil {
		_ = d.Range(func(key flux.GroupKey, value interface{}) error {
			if v, ok := value.(Closer); ok {
				_ = v.Close()
			}
			return nil
		})
	}

	// This parent is finished. Mark it so.
	t.finished++
	d.Clear()

	// Store any error for future iterations.
	t.err = err

	// From this point on, we're handling global state instead
	// of just state for the specific parent.
	if t.finished < len(t.parents) {
		// Do not continue until we get a finish from all parents.
		return
	}
	t.aggregateTransformation.Finish(id, t.err)
}

func (t *aggregateParallelTransformation) OperationType() string {
	return OperationType(t.t)
}

type aggregateParallelState struct {
	state interface{}
	count int
}

func (s *aggregateParallelState) Close() error {
	if c, ok := s.state.(Closer); ok {
		return c.Close()
	}
	return nil
}

type SimpleAggregateConfig struct {
	plan.DefaultCost
	Columns []string `json:"columns"`
}

var DefaultSimpleAggregateConfig = SimpleAggregateConfig{
	Columns: []string{DefaultValueColLabel},
}

func (c SimpleAggregateConfig) Copy() SimpleAggregateConfig {
	nc := c
	if c.Columns != nil {
		nc.Columns = make([]string, len(c.Columns))
		copy(nc.Columns, c.Columns)
	}
	return nc
}

func (c *SimpleAggregateConfig) ReadArgs(args flux.Arguments) error {
	if col, ok, err := args.GetString("column"); err != nil {
		return err
	} else if ok {
		c.Columns = []string{col}
	} else {
		c.Columns = DefaultSimpleAggregateConfig.Columns
	}
	return nil
}

func NewSimpleAggregateTransformation(ctx context.Context, id DatasetID, agg SimpleAggregate, config SimpleAggregateConfig, mem memory.Allocator) (Transformation, Dataset, error) {
	if feature.AggregateTransformationTransport().Enabled(ctx) {
		tr := &simpleAggregateTransformation2{
			agg:    agg,
			config: config,
		}
		return NewAggregateTransformation(id, tr, mem)
	}

	alloc, ok := mem.(*fluxmemory.ResourceAllocator)
	if !ok {
		alloc = &fluxmemory.ResourceAllocator{
			Allocator: mem,
		}
	}
	cache := NewTableBuilderCache(alloc)
	d := NewDataset(id, DiscardingMode, cache)
	return &simpleAggregateTransformation{
		d:      d,
		cache:  cache,
		agg:    agg,
		config: config,
	}, d, nil
}

type simpleAggregateTransformation struct {
	ExecutionNode
	d     Dataset
	cache TableBuilderCache
	agg   SimpleAggregate

	config SimpleAggregateConfig
}

func (t *simpleAggregateTransformation) RetractTable(id DatasetID, key flux.GroupKey) error {
	// TODO(nathanielc): Store intermediate state for retractions
	return t.d.RetractTable(key)
}

func (t *simpleAggregateTransformation) Process(id DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "aggregate found duplicate table with key: %v", tbl.Key())
	}

	if err := AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}

	builderColMap := make([]int, len(t.config.Columns))
	tableColMap := make([]int, len(t.config.Columns))
	aggregates := make([]ValueFunc, len(t.config.Columns))
	cols := tbl.Cols()
	for j, label := range t.config.Columns {
		idx := -1
		for bj, bc := range cols {
			if bc.Label == label {
				idx = bj
				break
			}
		}
		if idx < 0 {
			return errors.Newf(codes.FailedPrecondition, "column %q does not exist", label)
		}
		c := cols[idx]
		if tbl.Key().HasCol(c.Label) {
			return errors.New(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
		}
		var vf ValueFunc
		switch c.Type {
		case flux.TBool:
			vf = t.agg.NewBoolAgg()
		case flux.TInt:
			vf = t.agg.NewIntAgg()
		case flux.TUInt:
			vf = t.agg.NewUIntAgg()
		case flux.TFloat:
			vf = t.agg.NewFloatAgg()
		case flux.TString:
			vf = t.agg.NewStringAgg()
		}
		if vf == nil {
			return errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", c.Type)
		}
		aggregates[j] = vf

		var err error
		builderColMap[j], err = builder.AddCol(flux.ColMeta{
			Label: c.Label,
			Type:  vf.Type(),
		})
		if err != nil {
			return err
		}
		tableColMap[j] = idx
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		for j := range t.config.Columns {
			vf := aggregates[j]

			tj := tableColMap[j]
			c := tbl.Cols()[tj]

			switch c.Type {
			case flux.TBool:
				vf.(DoBoolAgg).DoBool(cr.Bools(tj))
			case flux.TInt:
				vf.(DoIntAgg).DoInt(cr.Ints(tj))
			case flux.TUInt:
				vf.(DoUIntAgg).DoUInt(cr.UInts(tj))
			case flux.TFloat:
				vf.(DoFloatAgg).DoFloat(cr.Floats(tj))
			case flux.TString:
				vf.(DoStringAgg).DoString(cr.Strings(tj))
			default:
				return errors.Newf(codes.Invalid, "unsupported aggregate type %v", c.Type)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for j, vf := range aggregates {
		bj := builderColMap[j]

		// If the value is null, append a null to the column.
		if vf.IsNull() {
			if err := builder.AppendNil(bj); err != nil {
				return err
			}
			if vf, ok := vf.(Closer); ok {
				if err := vf.Close(); err != nil {
					return err
				}
			}
			continue
		}

		// Append aggregated value
		switch vf.Type() {
		case flux.TBool:
			v := vf.(BoolValueFunc).ValueBool()
			if err := builder.AppendBool(bj, v); err != nil {
				return err
			}
		case flux.TInt:
			v := vf.(IntValueFunc).ValueInt()
			if err := builder.AppendInt(bj, v); err != nil {
				return err
			}
		case flux.TUInt:
			v := vf.(UIntValueFunc).ValueUInt()
			if err := builder.AppendUInt(bj, v); err != nil {
				return err
			}
		case flux.TFloat:
			v := vf.(FloatValueFunc).ValueFloat()
			if err := builder.AppendFloat(bj, v); err != nil {
				return err
			}
		case flux.TString:
			v := vf.(StringValueFunc).ValueString()
			if err := builder.AppendString(bj, v); err != nil {
				return err
			}
		}
		if vf, ok := vf.(Closer); ok {
			if err := vf.Close(); err != nil {
				return err
			}
		}
	}

	return AppendKeyValues(tbl.Key(), builder)
}

func (t *simpleAggregateTransformation) UpdateWatermark(id DatasetID, mark Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *simpleAggregateTransformation) UpdateProcessingTime(id DatasetID, pt Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *simpleAggregateTransformation) Finish(id DatasetID, err error) {
	t.d.Finish(err)
}

type simpleAggregateTransformation2 struct {
	agg    SimpleAggregate
	config SimpleAggregateConfig
}

type aggregateState struct {
	// inType is the column type of the input for this aggregate.
	inType flux.ColType

	// agg holds the aggregate function and associated state to produce a value.
	agg ValueFunc
}

func (s *aggregateState) Close() error {
	if v, ok := s.agg.(Closer); ok {
		return v.Close()
	}
	return nil
}

type aggregateStateList []aggregateState

func (a aggregateStateList) Close() (err error) {
	for i := range a {
		err = Close(err, &a[i])
	}
	return err
}

func (t *simpleAggregateTransformation2) initializeState(chunk table.Chunk, current interface{}) (aggregateStateList, error) {
	if current != nil {
		return current.(aggregateStateList), nil
	}

	state := make(aggregateStateList, len(t.config.Columns))
	for i, label := range t.config.Columns {
		j := chunk.Index(label)
		if j < 0 {
			return nil, errors.Newf(codes.FailedPrecondition, "column %q does not exist", label)
		} else if chunk.Key().HasCol(label) {
			return nil, errors.New(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
		}

		var vf ValueFunc
		col := chunk.Col(j)
		switch col.Type {
		case flux.TBool:
			vf = t.agg.NewBoolAgg()
		case flux.TInt:
			vf = t.agg.NewIntAgg()
		case flux.TUInt:
			vf = t.agg.NewUIntAgg()
		case flux.TFloat:
			vf = t.agg.NewFloatAgg()
		case flux.TString:
			vf = t.agg.NewStringAgg()
		default:
			return nil, errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", col.Type)
		}

		if vf == nil {
			return nil, errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", col.Type)
		}
		state[i].agg, state[i].inType = vf, col.Type
	}
	return state, nil
}

func (t *simpleAggregateTransformation2) Aggregate(chunk table.Chunk, state interface{}, mem memory.Allocator) (interface{}, bool, error) {
	aggregates, err := t.initializeState(chunk, state)
	if err != nil {
		return nil, false, err
	}

	for j, label := range t.config.Columns {
		idx := chunk.Index(label)
		if idx < 0 {
			return nil, false, errors.Newf(codes.FailedPrecondition, "column %q does not exist", label)
		}

		c := chunk.Col(idx)
		if inType := aggregates[j].inType; inType != c.Type {
			return nil, false, errors.Newf(codes.FailedPrecondition, "aggregate type conflict: %s != %s", c.Type, inType)
		}

		agg := aggregates[j].agg
		switch c.Type {
		case flux.TBool:
			agg.(DoBoolAgg).DoBool(chunk.Bools(idx))
		case flux.TInt:
			agg.(DoIntAgg).DoInt(chunk.Ints(idx))
		case flux.TUInt:
			agg.(DoUIntAgg).DoUInt(chunk.Uints(idx))
		case flux.TFloat:
			agg.(DoFloatAgg).DoFloat(chunk.Floats(idx))
		case flux.TString:
			agg.(DoStringAgg).DoString(chunk.Strings(idx))
		default:
			// This error should be impossible because loadState should have
			// already caught invalid input types and we have already verified
			// that the input type matches the type for this chunk.
			return nil, false, errors.Newf(codes.Internal, "aggregate of type %s not supported", c.Type)
		}
	}
	return aggregates, true, nil
}

func (t *simpleAggregateTransformation2) Compute(key flux.GroupKey, state interface{}, d *TransportDataset, mem memory.Allocator) error {
	aggregates := state.(aggregateStateList)
	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  make([]flux.ColMeta, 0, len(key.Cols())+len(aggregates)),
	}
	buffer.Columns = append(buffer.Columns, key.Cols()...)
	for i, s := range aggregates {
		buffer.Columns = append(buffer.Columns, flux.ColMeta{
			Label: t.config.Columns[i],
			Type:  s.agg.Type(),
		})
	}

	buffer.Values = make([]array.Array, len(key.Cols()), len(buffer.Columns))
	for j := range key.Cols() {
		buffer.Values[j] = arrow.Repeat(key.Cols()[j].Type, key.Value(j), 1, mem)
	}

	for _, s := range aggregates {
		var arr array.Array
		isNull := s.agg.IsNull()
		switch s.agg.Type() {
		case flux.TBool:
			v := s.agg.(BoolValueFunc).ValueBool()
			arr = array.BooleanRepeat(v, isNull, 1, mem)
		case flux.TInt:
			v := s.agg.(IntValueFunc).ValueInt()
			arr = array.IntRepeat(v, isNull, 1, mem)
		case flux.TUInt:
			v := s.agg.(UIntValueFunc).ValueUInt()
			arr = array.UintRepeat(v, isNull, 1, mem)
		case flux.TFloat:
			v := s.agg.(FloatValueFunc).ValueFloat()
			arr = array.FloatRepeat(v, isNull, 1, mem)
		case flux.TString:
			v := s.agg.(StringValueFunc).ValueString()
			arr = array.StringRepeat(v, 1, mem)
		}
		buffer.Values = append(buffer.Values, arr)
	}

	if err := buffer.Validate(); err != nil {
		return err
	}

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (t *simpleAggregateTransformation2) Close() error {
	if closer, ok := t.agg.(Closer); ok {
		return closer.Close()
	}
	return nil
}

type SimpleAggregate interface {
	NewBoolAgg() DoBoolAgg
	NewIntAgg() DoIntAgg
	NewUIntAgg() DoUIntAgg
	NewFloatAgg() DoFloatAgg
	NewStringAgg() DoStringAgg
}

type ValueFunc interface {
	Type() flux.ColType
	IsNull() bool
}
type DoBoolAgg interface {
	ValueFunc
	DoBool(*array.Boolean)
}
type DoFloatAgg interface {
	ValueFunc
	DoFloat(*array.Float)
}
type DoIntAgg interface {
	ValueFunc
	DoInt(*array.Int)
}
type DoUIntAgg interface {
	ValueFunc
	DoUInt(*array.Uint)
}
type DoStringAgg interface {
	ValueFunc
	DoString(*array.String)
}

type BoolValueFunc interface {
	ValueBool() bool
}
type FloatValueFunc interface {
	ValueFloat() float64
}
type IntValueFunc interface {
	ValueInt() int64
}
type UIntValueFunc interface {
	ValueUInt() uint64
}
type StringValueFunc interface {
	ValueString() string
}
