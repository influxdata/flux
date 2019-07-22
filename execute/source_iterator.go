package execute

import (
	"context"

	"github.com/influxdata/flux"
)

// SourceDecoder is an interface that generalizes the process of retrieving data from an unspecified data source.
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
	Connect(ctx context.Context) error
	Fetch(ctx context.Context) (bool, error)
	Decode(ctx context.Context) (flux.Table, error)
	Close() error
}

// CreateSourceFromDecoder takes an implementation of a SourceDecoder, as well as a dataset ID and Administration type
// and creates an execute.Source.
func CreateSourceFromDecoder(decoder SourceDecoder, dsid DatasetID, a Administration) (Source, error) {
	return &sourceIterator{decoder: decoder, id: dsid}, nil
}

type sourceIterator struct {
	decoder SourceDecoder
	id      DatasetID
	ts      []Transformation
}

func (c *sourceIterator) Do(ctx context.Context, f func(flux.Table) error) error {
	err := c.decoder.Connect(ctx)
	if err != nil {
		return err
	}
	defer c.decoder.Close()

	runOnce := true
	more, err := c.decoder.Fetch(ctx)
	if err != nil {
		return err
	}
	for runOnce || more {
		runOnce = false
		tbl, err := c.decoder.Decode(ctx)
		if err != nil {
			return err
		}
		if err := f(tbl); err != nil {
			return err
		}
		more, err = c.decoder.Fetch(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *sourceIterator) AddTransformation(t Transformation) {
	c.ts = append(c.ts, t)
}

func (c *sourceIterator) Run(ctx context.Context) {
	err := c.Do(ctx, func(tbl flux.Table) error {
		for _, t := range c.ts {
			err := t.Process(c.id, tbl)
			if err != nil {
				return err
			}
		}
		return nil
	})

	for _, t := range c.ts {
		t.Finish(c.id, err)
	}
}
