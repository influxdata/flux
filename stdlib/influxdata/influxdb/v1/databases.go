package v1

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
)

const DatabasesKind = "databases"

var DatabasesSignature = semantic.MustLookupBuiltinType("influxdata/influxdb/v1", "databases")

func init() {
	flux.RegisterPackageValue("influxdata/influxdb/v1", DatabasesKind, flux.MustValue(flux.FunctionValue(DatabasesKind, nil, DatabasesSignature)))
}
