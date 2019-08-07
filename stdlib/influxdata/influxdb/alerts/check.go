package alerts

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
)

const CheckKind = "check"

type CheckOpSpec struct{}

func init() {
	// TODO(affo): implement check: https://github.com/influxdata/flux/issues/1639
	checkSignature := flux.FunctionSignature(make(map[string]semantic.PolyType), []string{})
	flux.RegisterPackageValue("influxdata/influxdb/alerts", CheckKind, flux.FunctionValue("check", nil, checkSignature))
}
