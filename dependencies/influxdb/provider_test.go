package influxdb_test

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux/dependencies/influxdb"
	"github.com/google/go-cmp/cmp"
)

func TestGetProvider(t *testing.T) {
	want := influxdb.HttpProvider{
		DefaultConfig: influxdb.Config{
			Host: "http://localhost:8086",
		},
	}
	ctx := influxdb.Dependency{
		Provider: want,
	}.Inject(context.Background())

	got := influxdb.GetProvider(ctx)
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected provider -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestGetNoProvider(t *testing.T) {
	ctx := context.Background()

	got := influxdb.GetProvider(ctx)
	if _, ok := got.(influxdb.ErrorProvider); !ok {
		t.Fatalf("expected error provider, got:\n%T", got)
	}
}
