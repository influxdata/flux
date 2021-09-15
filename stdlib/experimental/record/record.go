package record

import (
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/record", "any", values.NewObjectWithValues(nil))
}
