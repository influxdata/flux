package universe

import (
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("universe", "true", values.NewBool(true))
	runtime.RegisterPackageValue("universe", "false", values.NewBool(false))
}
