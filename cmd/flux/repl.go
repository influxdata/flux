package main

import (
	"context"

	"github.com/influxdata/flux/repl"
)

func replE(ctx context.Context, enableSuggestions bool) error {
	r := repl.New(ctx, enableSuggestions)
	r.Run()
	return nil
}
