package options

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	flux.RegisterBuiltInOption("now", universe.SystemTime())
}
