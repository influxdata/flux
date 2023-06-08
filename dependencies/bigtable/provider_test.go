package bigtable_test

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux/dependencies/bigtable"
)

func TestGetNoProvider(t *testing.T) {
	ctx := context.Background()

	got := bigtable.GetProvider(ctx)
	if _, ok := got.(bigtable.ErrorProvider); !ok {
		t.Fatalf("expected error provider, got:\n%T", got)
	}
}
