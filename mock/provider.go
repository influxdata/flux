package mock

import (
	"context"

	"github.com/influxdata/flux/dependencies/influxdb"
)

type MockProvider struct {
	influxdb.UnimplementedProvider
	WriterForFn func(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error)
}

var _ influxdb.Provider = &MockProvider{}

func (m MockProvider) WriterFor(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error) {
	return m.WriterForFn(ctx, conf)
}
