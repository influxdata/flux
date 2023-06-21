package runtime

import (
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/values"
)

const versionFuncName = "version"

var errBuildInfoNotPresent = errors.New(codes.NotFound, "build info is not present")

func init() {
	runtime.RegisterPackageValue("runtime", versionFuncName, values.NewFunction(
		versionFuncName,
		runtime.MustLookupBuiltinType("runtime", versionFuncName),
		Version,
		false,
	))
}
