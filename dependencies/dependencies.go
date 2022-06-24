package dependencies

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/bigtable"
	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/filesystem"
	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/influxdb"
	"github.com/mvn-trinhnguyen2-dn/flux/dependencies/mqtt"
)

type Dependencies struct {
	flux.Deps
	influxdb influxdb.Dependency
	bigtable bigtable.Dependency
	mqtt     mqtt.Dependency
}

func (d Dependencies) Inject(ctx context.Context) context.Context {
	ctx = d.Deps.Inject(ctx)
	ctx = d.influxdb.Inject(ctx)
	ctx = d.bigtable.Inject(ctx)
	return d.mqtt.Inject(ctx)
}

func NewDefaultDependencies(defaultInfluxDBHost string) Dependencies {
	deps := flux.NewDefaultDependencies()
	deps.Deps.FilesystemService = filesystem.SystemFS

	return Dependencies{
		Deps: deps,

		influxdb: influxdb.Dependency{
			Provider: &influxdb.HttpProvider{
				DefaultConfig: influxdb.Config{
					Host: defaultInfluxDBHost,
				},
			},
		},

		bigtable: bigtable.Dependency{
			Provider: bigtable.DefaultProvider{},
		},

		mqtt: mqtt.Dependency{
			Dialer: mqtt.DefaultDialer{},
		},
	}
}

func NewErrorDependencies() Dependencies {
	deps := flux.NewEmptyDependencies()

	return Dependencies{
		Deps: deps,

		influxdb: influxdb.Dependency{
			Provider: &influxdb.ErrorProvider{},
		},

		bigtable: bigtable.Dependency{
			Provider: bigtable.ErrorProvider{},
		},

		mqtt: mqtt.Dependency{
			Dialer: mqtt.ErrorDialer{},
		},
	}
}
