package executetest

import (
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
)

func NewTestExecuteDependencies() execute.Dependencies {
	return execute.Dependencies{dependencies.InterpreterDepsKey: dependenciestest.Default()}
}
