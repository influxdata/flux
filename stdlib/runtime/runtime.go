package runtime

import (
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
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
