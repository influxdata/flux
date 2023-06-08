package boolean

import (
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/values"
)

func init() {
	runtime.RegisterPackageValue("internal/boolean", "true", values.NewBool(true))
	runtime.RegisterPackageValue("internal/boolean", "false", values.NewBool(false))
}
