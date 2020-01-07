package v1

import (
	"github.com/influxdata/flux"
)

const DatabasesKind = "databases"

var DatabasesSignature = semantic.LookupBuiltInType("influxdata/influxdb/v1", "databses")

func init() {
	flux.RegisterPackageValue("influxdata/influxdb/v1", DatabasesKind, flux.MustValue(flux.FunctionValue(DatabasesKind, nil, DatabasesSignature)))
}
