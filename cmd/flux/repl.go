package main

import (
	"context"

	"github.com/influxdata/flux/repl"
)

func replE(ctx context.Context, suggestions bool) error {
	r := repl.New(ctx, suggestions)
	r.Run()
	return nil
}
