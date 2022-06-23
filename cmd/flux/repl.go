package main

import (
	"context"

	"github.com/influxdata/flux/repl"
)

func replE(ctx context.Context, opts ...repl.Option) error {
	r := repl.New(ctx, opts...)
	r.Run()
	return nil
}
