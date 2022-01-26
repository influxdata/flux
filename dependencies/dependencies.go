package dependencies

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/bigtable"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/dependencies/mqtt"
	"github.com/influxdata/flux/dependency"
)

func NewDefaultDependencies(defaultInfluxDBHost string) flux.Dependency {
	deps := flux.NewDefaultDependencies()
	deps.Deps.FilesystemService = filesystem.SystemFS
	return dependency.List{
		deps,
		influxdb.Dependency{
			Provider: &influxdb.HttpProvider{
				DefaultConfig: influxdb.Config{
					Host: defaultInfluxDBHost,
				},
			},
		},
		bigtable.Dependency{
			Provider: bigtable.DefaultProvider{},
		},
		mqtt.Dependency{
			Dialer: mqtt.DefaultDialer{},
		},
	}
}

func NewErrorDependencies() flux.Dependency {
	deps := flux.NewEmptyDependencies()
	return dependency.List{
		deps,
		influxdb.Dependency{
			Provider: &influxdb.ErrorProvider{},
		},
		bigtable.Dependency{
			Provider: bigtable.ErrorProvider{},
		},
		mqtt.Dependency{
			Dialer: mqtt.ErrorDialer{},
		},
	}
}
