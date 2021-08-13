package execute

import (
	"context"

	"github.com/apache/arrow/go/arrow/memory"
	uuid "github.com/gofrs/uuid"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

// Dataset represents the set of data produced by a transformation.
type Dataset interface {
	Node

	RetractTable(key flux.GroupKey) error
	UpdateProcessingTime(t Time) error
	UpdateWatermark(mark Time) error
	Finish(error)

	SetTriggerSpec(t plan.TriggerSpec)
}

// DatasetContext represents a Dataset with a context.Context attached.
type DatasetContext interface {
	Dataset
	WithContext(ctx context.Context)
}

// DataCache holds all working data for a transformation.
type DataCache interface {
	Table(flux.GroupKey) (flux.Table, error)

	ForEach(func(flux.GroupKey))
	ForEachWithContext(func(flux.GroupKey, Trigger, TableContext))

	DiscardTable(flux.GroupKey)
	ExpireTable(flux.GroupKey)

	SetTriggerSpec(t plan.TriggerSpec)
}

type AccumulationMode int

const (
	// DiscardingMode will discard the data associated with a group key
	// after it has been processed.
	DiscardingMode AccumulationMode = iota

	// AccumulatingMode will retain the data associated with a group key
	// after it has been processed. If it has already sent a table with
	// that group key to a downstream transformation, it will signal
	// to that transformation that the previous table should be retracted.
	//
	// This is not implemented at the moment.
	AccumulatingMode
)

type DatasetID uuid.UUID

func (id DatasetID) String() string {
	return uuid.UUID(id).String()
}

var ZeroDatasetID DatasetID

func (id DatasetID) IsZero() bool {
	return id == ZeroDatasetID
}

func DatasetIDFromNodeID(id plan.NodeID) DatasetID {
	return DatasetID(uuid.NewV5(uuid.UUID{}, string(id)))
}

type dataset struct {
	ctx context.Context
	id  DatasetID

	ts      TransformationSet
	accMode AccumulationMode

	watermark      Time
	processingTime Time

	cache DataCache
}

func NewDataset(id DatasetID, accMode AccumulationMode, cache DataCache) *dataset {
	return &dataset{
		ctx:     context.Background(),
		id:      id,
		accMode: accMode,
		cache:   cache,
	}
}

func (d *dataset) WithContext(ctx context.Context) {
	d.ctx = ctx
}

func (d *dataset) AddTransformation(t Transformation) {
	d.ts = append(d.ts, t)
}

func (d *dataset) SetTriggerSpec(spec plan.TriggerSpec) {
	d.cache.SetTriggerSpec(spec)
}

func (d *dataset) UpdateWatermark(mark Time) error {
	d.watermark = mark
	if err := d.evalTriggers(); err != nil {
		return err
	}
	return d.ts.UpdateWatermark(d.id, mark)
}

func (d *dataset) UpdateProcessingTime(time Time) error {
	d.processingTime = time
	if err := d.evalTriggers(); err != nil {
		return err
	}
	return d.ts.UpdateProcessingTime(d.id, time)
}

func (d *dataset) evalTriggers() (err error) {
	d.cache.ForEachWithContext(func(key flux.GroupKey, trigger Trigger, bc TableContext) {
		if err != nil {
			// Skip the rest once we have encountered an error
			return
		}

		if err = d.ctx.Err(); err != nil {
			return
		}

		c := TriggerContext{
			Table:                 bc,
			Watermark:             d.watermark,
			CurrentProcessingTime: d.processingTime,
		}

		if trigger.Triggered(c) {
			err = d.triggerTable(key)
		}
		if trigger.Finished() {
			d.expireTable(key)
		}
	})
	return err
}

func (d *dataset) triggerTable(key flux.GroupKey) error {
	b, err := d.cache.Table(key)
	if err != nil {
		return err
	}

	switch d.accMode {
	case DiscardingMode:
		if err := d.ts.Process(d.id, b); err != nil {
			return err
		}
		d.cache.DiscardTable(key)
	case AccumulatingMode:
		return errors.New(codes.Unimplemented)
	}
	return nil
}

func (d *dataset) expireTable(key flux.GroupKey) {
	d.cache.ExpireTable(key)
}

func (d *dataset) RetractTable(key flux.GroupKey) error {
	d.cache.DiscardTable(key)
	return d.ts.RetractTable(d.id, key)
}

func (d *dataset) Finish(err error) {
	if err == nil {
		// Only trigger tables we if we not finishing because of an error.
		d.cache.ForEach(func(bk flux.GroupKey) {
			if err != nil {
				return
			}

			if err = d.ctx.Err(); err != nil {
				return
			}

			err = d.triggerTable(bk)
			d.cache.ExpireTable(bk)
		})
	}
	d.ts.Finish(d.id, err)
}

// PassthroughDataset is a Dataset that will passthrough
// the processed data to the next Transformation.
type PassthroughDataset struct {
	id DatasetID
	ts TransformationSet
}

// NewPassthroughDataset constructs a new PassthroughDataset.
func NewPassthroughDataset(id DatasetID) *PassthroughDataset {
	return &PassthroughDataset{id: id}
}

func (d *PassthroughDataset) AddTransformation(t Transformation) {
	d.ts = append(d.ts, t)
}

func (d *PassthroughDataset) Process(tbl flux.Table) error {
	return d.ts.Process(d.id, tbl)
}

func (d *PassthroughDataset) RetractTable(key flux.GroupKey) error {
	return d.ts.RetractTable(d.id, key)
}

func (d *PassthroughDataset) UpdateProcessingTime(t Time) error {
	return d.ts.UpdateProcessingTime(d.id, t)
}

func (d *PassthroughDataset) UpdateWatermark(mark Time) error {
	return d.ts.UpdateWatermark(d.id, mark)
}

func (d *PassthroughDataset) Finish(err error) {
	d.ts.Finish(d.id, err)
}

func (d *PassthroughDataset) SetTriggerSpec(t plan.TriggerSpec) {
}

// TransportDataset holds data for a specific node and sends
// messages to downstream nodes using the Transport.
//
// This Dataset also implements a shim for execute.Dataset
// so it can be integrated with the existing execution engine.
// These methods are stubs and do not do anything.
type TransportDataset struct {
	id         DatasetID
	transports []Transport
	cache      *RandomAccessGroupLookup
	mem        memory.Allocator
}

// NewTransportDataset constructs a TransportDataset.
func NewTransportDataset(id DatasetID, mem memory.Allocator) *TransportDataset {
	return &TransportDataset{
		id:    id,
		cache: NewRandomAccessGroupLookup(),
		mem:   mem,
	}
}

// AddTransformation is used to add downstream Transformation nodes
// to this Transport.
func (d *TransportDataset) AddTransformation(t Transformation) {
	if t, ok := t.(Transport); ok {
		d.transports = append(d.transports, t)
		return
	}
	d.transports = append(d.transports, WrapTransformationInTransport(t, d.mem))
}

func (d *TransportDataset) sendMessage(m Message) error {
	if len(d.transports) == 1 {
		return d.transports[0].ProcessMessage(m)
	}

	defer m.Ack()
	for _, t := range d.transports {
		if err := t.ProcessMessage(m.Dup()); err != nil {
			return err
		}
	}
	return nil
}

// Process sends the given Chunk to be processed by the downstream transports.
func (d *TransportDataset) Process(chunk table.Chunk) error {
	m := &processChunkMsg{
		srcMessage: srcMessage(d.id),
		chunk:      chunk,
	}
	return d.sendMessage(m)
}

// FlushKey sends the flush key message to the downstream transports.
func (d *TransportDataset) FlushKey(key flux.GroupKey) error {
	m := &flushKeyMsg{
		srcMessage: srcMessage(d.id),
		key:        key,
	}
	return d.sendMessage(m)
}

func (d *TransportDataset) Lookup(key flux.GroupKey) (interface{}, bool) {
	return d.cache.Lookup(key)
}
func (d *TransportDataset) LookupOrCreate(key flux.GroupKey, fn func() interface{}) interface{} {
	if fn == nil {
		fn = func() interface{} {
			return nil
		}
	}
	return d.cache.LookupOrCreate(key, fn)
}
func (d *TransportDataset) Set(key flux.GroupKey, value interface{}) {
	d.cache.Set(key, value)
}
func (d *TransportDataset) Delete(key flux.GroupKey) (v interface{}, found bool) {
	return d.cache.Delete(key)
}
func (d *TransportDataset) Range(f func(key flux.GroupKey, value interface{}) error) (err error) {
	d.cache.Range(func(key flux.GroupKey, value interface{}) {
		if err != nil {
			return
		}
		err = f(key, value)
	})
	return err
}

func (d *TransportDataset) RetractTable(key flux.GroupKey) error { return nil }
func (d *TransportDataset) UpdateProcessingTime(t Time) error    { return nil }
func (d *TransportDataset) UpdateWatermark(mark Time) error      { return nil }
func (d *TransportDataset) Finish(err error) {
	m := &finishMsg{
		srcMessage: srcMessage(d.id),
		err:        err,
	}
	_ = d.sendMessage(m)
	d.cache.Clear()
}
func (d *TransportDataset) SetTriggerSpec(t plan.TriggerSpec) {}
