package main

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux/repl"
)

func replE(ctx context.Context) error {
	r := repl.New(ctx)
	r.Run()
	return nil
}
