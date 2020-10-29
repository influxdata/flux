package influxdb_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/dependencies/influxdb"
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
