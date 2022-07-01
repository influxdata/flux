package mock

import (
	"context"

	"github.com/influxdata/flux/dependencies/influxdb"
)

type InfluxDBProvider struct {
	influxdb.UnimplementedProvider
	WriterForFn func(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error)
}

var _ influxdb.Provider = &InfluxDBProvider{}

func (m InfluxDBProvider) WriterFor(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error) {
	return m.WriterForFn(ctx, conf)
}
