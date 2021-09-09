package boolean

import (
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("internal/boolean", "true", values.NewBool(true))
	runtime.RegisterPackageValue("internal/boolean", "false", values.NewBool(false))
}
