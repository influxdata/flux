package v1

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
)

const DatabasesKind = "databases"

var DatabasesSignature = semantic.FunctionPolySignature{
	Return: flux.TableObjectType,
}

func init() {
	flux.RegisterPackageValue("influxdata/influxdb/v1", DatabasesKind, flux.FunctionValue(DatabasesKind, nil, DatabasesSignature))
}
