package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

type dataset struct {
	id    execute.DatasetID
	ts    execute.TransformationSet
	cache *BuilderCache
}

// NewDataset constructs an execute.Dataset that is compatible with
// the BuilderCache.
//
// This dataset does not support triggers and will only flush tables
// when the dataset is finished.
func NewDataset(id execute.DatasetID, cache *BuilderCache) execute.Dataset {
	return &dataset{
		id:    id,
		cache: cache,
	}
}

func (d *dataset) AddTransformation(t execute.Transformation) {
	d.ts = append(d.ts, t)
}

func (d *dataset) SetTriggerSpec(spec plan.TriggerSpec) {
}

func (d *dataset) UpdateWatermark(mark execute.Time) error {
	return d.ts.UpdateWatermark(d.id, mark)
}

func (d *dataset) UpdateProcessingTime(time execute.Time) error {
	return d.ts.UpdateProcessingTime(d.id, time)
}

func (d *dataset) RetractTable(key flux.GroupKey) error {
	d.cache.DiscardTable(key)
	return d.ts.RetractTable(d.id, key)
}

func (d *dataset) Finish(err error) {
	if err == nil {
		err = d.cache.ForEach(func(key flux.GroupKey, builder Builder) error {
			tbl, err := builder.Table()
			if err != nil {
				return err
			}
			return d.ts.Process(d.id, tbl)
		})
	}
	d.ts.Finish(d.id, err)
}
