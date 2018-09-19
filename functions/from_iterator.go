package functions

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
)

// Source Decoder is an interface that generalizes the process of retrieving data from an unspecified data source.
//
// Connect implements the logic needed to connect directly to the data source.
//
// Fetch implements a single fetch of data from the source (may be called multiple times).  Should return false when
// there is no more data to retrieve.
//
// Decode implements the process of marshaling the data returned by the source into a flux.Table type.
//
// In executing the retrieval process, Connect is called once at the onset, and subsequent calls of Fetch() and Decode()
// are called iteratively until the data source is fully consumed.
type SourceDecoder interface {
	Connect() error
	Fetch() (bool, error)
	Decode() (flux.Table, error)
}

// CreateFromSourceIterator takes an implementation of a SourceDecoder, as well as a dataset ID and Administration type
// and creates a custom sourceIterator type that is a valid execute.Source type.
func CreateFromSourceIterator(decoder SourceDecoder, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	return &sourceIterator{decoder: decoder, id: dsid}, nil
}

type sourceIterator struct {
	// TODO: add fields you need to connect, fetch, etc.
	decoder SourceDecoder
	id   execute.DatasetID
	ts   []execute.Transformation
}

func (c *sourceIterator) Do(f func(flux.Table) error) error {
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

func (c *sourceIterator) AddTransformation(t execute.Transformation) {
	c.ts = append(c.ts, t)
}

func (c *sourceIterator) Run(ctx context.Context) {
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
