package main

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/repl"
)

func replE(ctx context.Context, deps flux.Dependencies) error {
	r := repl.New(ctx, deps)
	r.Run()
	return nil
}
