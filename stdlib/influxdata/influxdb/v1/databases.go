package v1

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
)

const DatabasesKind = "databases"

var DatabasesSignature = runtime.MustLookupBuiltinType("influxdata/influxdb/v1", "databases")

func init() {
	runtime.RegisterPackageValue("influxdata/influxdb/v1", DatabasesKind, flux.MustValue(flux.FunctionValue(DatabasesKind, nil, DatabasesSignature)))
}
