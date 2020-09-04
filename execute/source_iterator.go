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
	return &sourceDecoder{decoder: decoder, id: dsid}, nil
}

type sourceDecoder struct {
	ExecutionNode
	decoder SourceDecoder
	id      DatasetID
	ts      []Transformation
}

func (c *sourceDecoder) Do(ctx context.Context, f func(flux.Table) error) error {
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

func (c *sourceDecoder) AddTransformation(t Transformation) {
	c.ts = append(c.ts, t)
}

func (c *sourceDecoder) Run(ctx context.Context) {
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

// CreateSourceFromIterator takes an implementation of a SourceIterator as well as a dataset ID
// and creates an execute.Source.
func CreateSourceFromIterator(iterator SourceIterator, dsid DatasetID) (Source, error) {
	return &sourceIterator{iterator: iterator, id: dsid}, nil
}

// SourceIterator is an interface for iterating over flux.Table values in
// a source. It provides a common interface for creating an execute.Source
// in an iterative way.
type SourceIterator interface {
	// Do will invoke the Source and cause each materialized flux.Table
	// to the given function.
	Do(ctx context.Context, f func(flux.Table) error) error
}

// sourceIterator implements execute.Source using the SourceIterator.
type sourceIterator struct {
	ExecutionNode
	id       DatasetID
	ts       []Transformation
	iterator SourceIterator
}

func (s *sourceIterator) AddTransformation(t Transformation) {
	s.ts = append(s.ts, t)
}

func (s *sourceIterator) Run(ctx context.Context) {
	err := s.iterator.Do(ctx, s.processTable)
	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}

// processTable will call Process on all of the transformations
// associated with this source.
//
// If there are multiple sources, it will copy the table so it
// can be properly buffered to the multiple transformations.
func (s *sourceIterator) processTable(tbl flux.Table) error {
	if len(s.ts) == 0 {
		tbl.Done()
		return nil
	} else if len(s.ts) == 1 {
		return s.ts[0].Process(s.id, tbl)
	}

	// There is more than one transformation so we need to
	// copy the table for each transformation.
	bufTable, err := CopyTable(tbl)
	if err != nil {
		return err
	}
	defer bufTable.Done()

	for _, t := range s.ts {
		if err := t.Process(s.id, bufTable.Copy()); err != nil {
			return err
		}
	}
	return nil
}
