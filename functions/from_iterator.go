package functions

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
)

type SourceDecoder interface {
	Connect() error
	Fetch() (bool, error)
	Decode() (flux.Table, error)
}

func CreateFromSourceIterator(decoder SourceDecoder, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	return &SourceIterator{decoder: decoder, id: dsid}, nil
}

type SourceIterator struct {
	// TODO: add fields you need to connect, fetch, etc.
	decoder SourceDecoder
	id   execute.DatasetID
	ts   []execute.Transformation
}

func (c *SourceIterator) Do(f func(flux.Table) error) error {
	err := c.decoder.Connect()
	if err != nil {
		return err
	}
	runOnce := true
	more, err := c.decoder.Fetch()
	if err != nil {
		return err
	}
	for runOnce || more {
		runOnce = false
		tbl, err := c.decoder.Decode()
		if err != nil {
			return err
		}
		f(tbl)
		more, err = c.decoder.Fetch()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SourceIterator) AddTransformation(t execute.Transformation) {
	c.ts = append(c.ts, t)
}

func (c *SourceIterator) Run(ctx context.Context) {
	var err error
	var max execute.Time
	maxSet := false
	err = c.Do(func(tbl flux.Table) error {
		for _, t := range c.ts {
			err := t.Process(c.id, tbl)
			if err != nil {
				return err
			}
			if idx := execute.ColIdx(execute.DefaultStopColLabel, tbl.Key().Cols()); idx >= 0 {
				if stop := tbl.Key().ValueTime(idx); !maxSet || stop > max {
					max = stop
					maxSet = true
				}
			}
		}
		return nil
	})
	if err != nil {
		goto FINISH
	}

	if maxSet {
		for _, t := range c.ts {
			t.UpdateWatermark(c.id, max)
		}
	}

FINISH:
	for _, t := range c.ts {
		t.Finish(c.id, err)
	}
}
