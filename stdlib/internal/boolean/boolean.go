package boolean

import (
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

func init() {
	runtime.RegisterPackageValue("internal/boolean", "true", values.NewBool(true))
	runtime.RegisterPackageValue("internal/boolean", "false", values.NewBool(false))
}
