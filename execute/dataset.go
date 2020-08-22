package execute

import (
	uuid "github.com/gofrs/uuid"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
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
	id DatasetID

	ts      TransformationSet
	accMode AccumulationMode

	watermark      Time
	processingTime Time

	cache DataCache
}

func NewDataset(id DatasetID, accMode AccumulationMode, cache DataCache) *dataset {
	return &dataset{
		id:      id,
		accMode: accMode,
		cache:   cache,
	}
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
