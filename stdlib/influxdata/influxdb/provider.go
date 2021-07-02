package influxdb

import (
	"context"

	"github.com/influxdata/flux/dependencies/influxdb"
)

type Provider = influxdb.Provider

func GetProvider(ctx context.Context) Provider {
	return influxdb.GetProvider(ctx)
}
