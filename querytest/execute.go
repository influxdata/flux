package querytest

import (
	"context"
	"io"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
	"github.com/pkg/errors"
)

type Querier struct{}

func (q *Querier) Query(ctx context.Context, w io.Writer, c flux.Compiler, d flux.Dialect) (int64, error) {
	program, err := c.Compile(ctx)
	if err != nil {
		return 0, err
	}
	alloc := &memory.Allocator{}
	query, err := program.Start(ctx, alloc)
	if err != nil {
		return 0, err
	}
	results := flux.NewResultIteratorFromQuery(query)
	defer results.Release()

	encoder := d.Encoder()
	er, err := encoder.Encode(w, results)
	if err != nil {
		// An error occurred during encoding
		return er.BytesWritten, errors.Wrap(err, "error encoding result")
	}

	if len(er.Errs) > 0 {
		if er, err := encoder.EncodeErrors(w, er); err != nil {
			return er.BytesWritten, err
		} else {
			return er.BytesWritten, nil
		}
	}
	return er.BytesWritten, nil
}

func NewQuerier() *Querier {
	return &Querier{}
}
