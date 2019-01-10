package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("universe", "true", values.NewBool(true))
	flux.RegisterPackageValue("universe", "false", values.NewBool(false))
}
