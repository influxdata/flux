package executetest

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
)

func NewTestExecuteDependencies() flux.Dependencies {
	return dependenciestest.Default()
}
