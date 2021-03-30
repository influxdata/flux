package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/runtime"
)

const (
	APIFuncName = "api"
	PackagePath = "experimental/influxdb"
	APIKind     = PackagePath + "." + APIFuncName
)

var APISignature = runtime.MustLookupBuiltinType(PackagePath, APIFuncName)

func init() {
	runtime.RegisterPackageValue(PackagePath, APIFuncName,
		flux.MustValue(flux.FunctionValueWithSideEffect("api", nil, APISignature)),
	)
}
