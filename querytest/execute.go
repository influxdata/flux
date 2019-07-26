package querytest

import (
	"context"
	"io"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
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
	return encoder.Encode(w, results)
}

func NewQuerier() *Querier {
	return &Querier{}
}
