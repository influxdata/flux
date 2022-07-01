package main

import (
	"context"

	"github.com/influxdata/flux/repl"
)

func replE(ctx context.Context) error {
	r := repl.New(ctx)
	r.Run()
	return nil
}
